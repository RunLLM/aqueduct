import json
import re
import uuid
from typing import Any, Dict, List, Optional, Union

from aqueduct.integrations.parameters import BUILT_IN_EXPANSIONS, TAG_PATTERN
from aqueduct_executor.operators.connectors.data import common, models
from aqueduct_executor.operators.utils.enums import ArtifactType
from pydantic import parse_obj_as

# The TAG for 'previous table' when the user specifies a chained query.
PREV_TABLE_TAG = "$"


def _replace_builtin_tags(query: str) -> str:
    """Expands any builtin tags found in the raw query, eg. {{ today }}."""
    matches = re.findall(TAG_PATTERN, query)
    for match in matches:
        tag_name = match.strip(" {}")
        if tag_name in BUILT_IN_EXPANSIONS:
            expansion_func = BUILT_IN_EXPANSIONS[tag_name]
            query = query.replace(match, expansion_func())
    return query


def _replace_param_placeholders(query: str, parameter_vals: List[str]) -> str:
    """Replaces any user-defined placeholders in the query with the corresponding parameter value.

    Assumes that we've already validated that every parameter value has a corresponding placeholder in the query.
    """
    for i in range(len(parameter_vals)):
        query = query.replace("$" + str(i + 1), parameter_vals[i])
    return query


class RelationalParams(models.BaseParams):
    # The query cannot be used until `apply_placeholders()` is called on it. This flushes out
    # any user-defined tags like `{{today}}`.
    query_is_usable: Optional[bool] = False

    # Exactly one of 'query' and 'queries' will be set.
    # `query` represents a single query.
    query: Optional[str] = None
    # `queries` represents a chain of queries. We must first run _compile_chain
    # to compile it to a single query before further processing.
    queries: Optional[List[str]] = None

    # TODO: Consider not including github as part of relational params when it is JSON marshalled
    github_metadata: Optional[Any]

    def _compile_chain(self, queries: List[str]) -> str:
        """
        `_compile_chain` compiles a chain query to a single query using `WITH` clause.
        We generate temp_table_name and replace PREV_TABLE_TAG accordingly.

        Example:
        queries: [
            "SELECT * FROM my_table",
            "SELECT field_a, field_b FROM $",
            "SELECT * FROM $",
        ]

        returns: `
            WITH
                generated_tmp_a AS (SELECT * FROM my_table),
                generated_tmp_b AS (SELECT field_a, field_b FROM generated_tmp_a)
            SELECT * FROM generated_tmp_b
        `
        """
        if not queries:
            return ""

        if len(queries) == 1:
            return queries[0]

        with_clause = "WITH\n"
        prev_table_name = ""
        normalized_query = ""
        for idx, query in enumerate(queries):
            # Remove spaces and trailing semicolumns if any.
            normalized_query = query.strip().rstrip(";")

            # Replace tag except for the first query
            if idx == 0:
                if PREV_TABLE_TAG in normalized_query:
                    raise Exception(
                        f"Cannot compile chain. {PREV_TABLE_TAG} appears in the first query: {query}"
                    )
            else:
                normalized_query = normalized_query.replace(PREV_TABLE_TAG, prev_table_name)

            # Subquery goes to the 'WITH' clause except for the last one.
            if idx < len(queries) - 1:
                cur_table_name = f"aqueduct_{uuid.uuid4().hex}"
                with_clause += f"{cur_table_name} AS (\n{normalized_query}\n)"

                # There are more subqueries to append in this 'WITH' clause
                if idx < len(queries) - 2:
                    with_clause += ",\n"
                prev_table_name = cur_table_name

        # Returns `WITH` clause with the normalized final query.
        return f"{with_clause}\n{normalized_query}"

    def compile(self, parameter_vals: List[str]) -> None:
        """
        `compile` compiles this object to a single query that can be
        executed.
        """
        assert (
            int(bool(self.query)) + int(bool(self.queries)) == 1
        ), "Exactly one of .query and .queries fields should be set."

        queries = self.queries or []
        if self.query:
            queries = [self.query]

        # Expand the placeholders first, before collapsing the query chain, since $ is broader than $1, $2, etc.
        for i, q in enumerate(queries):
            q = _replace_param_placeholders(q, parameter_vals)
            queries[i] = _replace_builtin_tags(q)
        print(f"Expanded queries are `{queries}`.")

        print(f"Compiling queries {queries} .")
        query = self._compile_chain(queries)
        print(f"Compiled query is {query} .")

        self.query = query
        self.query_is_usable = True

    def usable(self) -> bool:
        """Denotes whether all placeholders have already been expanded for this query.

        Callers should check that `usable()` -> True before actually executing this query.
        """
        # We cannot return self.query_is_usable directly, since it is an Optional
        # and the method expects a bool to be returned.
        return bool(self.query_is_usable)


class S3Params(models.BaseParams):
    filepath: str
    artifact_type: ArtifactType
    format: Optional[common.S3TableFormat]
    merge: Optional[bool]


class MongoDBFindParams(models.BaseParams):
    args: Optional[List[Any]] = None
    kwargs: Optional[Dict[str, Any]] = None


class MongoDBParams(models.BaseParams):
    collection: str
    query_serialized: str
    query: Optional[MongoDBFindParams] = None

    def compile(self, parameters: List[str]) -> None:
        expanded = _replace_param_placeholders(self.query_serialized, parameters)
        self.query = parse_obj_as(MongoDBFindParams, json.loads(expanded))

    def usable(self) -> bool:
        return bool(self.query) and bool(self.collection)


Params = Union[RelationalParams, S3Params, MongoDBParams]

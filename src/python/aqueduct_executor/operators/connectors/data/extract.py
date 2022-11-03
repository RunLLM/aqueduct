import json
import re
import uuid
from datetime import date
from typing import Any, Dict, List, Optional, Union

from aqueduct.integrations.sql_integration import BUILT_IN_EXPANSIONS, PREV_TABLE_TAG, TAG_PATTERN
from aqueduct_executor.operators.connectors.data import common, models
from aqueduct_executor.operators.utils import enums
from pydantic import parse_obj_as


def replace_today() -> str:
    return "'" + date.today().strftime("%Y-%m-%d") + "'"


def _expand_placeholders(query: str, parameters: Dict[str, str]) -> str:
    """Expands any tags found in the raw query, eg. {{ today }}.

    Relational queries can be arbitrarily parameterized the same way operators are. The only
    requirement is that these parameters must be defined as strings.

    User-defined parameters are prioritized over built-in ones. Eg. if the user defines a parameter
    named "today" that they set with value "1234", the "{{today}}" will be expanded as "1234", even
    though there already is a built-in expansion.
    """
    matches = re.findall(TAG_PATTERN, query)
    for match in matches:
        tag_name = match.strip(" {}")

        if tag_name in parameters:
            query = query.replace(match, parameters[tag_name])
        elif tag_name in BUILT_IN_EXPANSIONS:
            expansion_func = BUILT_IN_EXPANSIONS[tag_name]
            query = query.replace(match, expansion_func())
        else:
            # If there's a tag in the query for which no expansion value is available, we error here.
            raise Exception("Unable to expand tag `%s` for query `%s`." % (tag_name, query))
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
        for (idx, query) in enumerate(queries):
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

    def compile(self, parameters: Dict[str, str]) -> None:
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

        print(f"Compiling queries {queries} .")
        query = self._compile_chain(queries)
        print(f"Compiled query is {query} .")

        query = _expand_placeholders(query, parameters)
        print(f"Expanded query is `{query}`.")
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
    artifact_type: enums.ArtifactType
    format: Optional[common.S3TableFormat]
    merge: Optional[bool]


class MongoDBFindParams(models.BaseParams):
    args: Optional[List[Any]] = None
    kwargs: Optional[Dict[str, Any]] = None


class MongoDBParams(models.BaseParams):
    collection: str
    query_serialized: str
    query: Optional[MongoDBFindParams] = None

    def compile(self, parameters: Dict[str, str]) -> None:
        expanded = _expand_placeholders(self.query_serialized, parameters)
        self.query = parse_obj_as(MongoDBFindParams, json.loads(expanded))

    def usable(self) -> bool:
        return bool(self.query) and bool(self.collection)


Params = Union[RelationalParams, S3Params, MongoDBParams]

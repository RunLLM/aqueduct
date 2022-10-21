import json
import re
import uuid
from datetime import date
from typing import Any, Dict, List, Optional, Union

from aqueduct_executor.operators.connectors.data import common, models
from aqueduct_executor.operators.utils import enums

# Regular Expression that matches any substring appearance with
# "{{ }}" and a word inside with optional space in front or after
# Potential Matches: "{{today}}", "{{ today  }}""
#
# Duplicated in the SDK at `sdk/aqueduct/integrations/sql_integration.py`.
# Make sure the two are in sync.
TAG_PATTERN = r"{{\s*[\w-]+\s*}}"
CHAIN_TABLE_TAG = "$"


def replace_today() -> str:
    return "'" + date.today().strftime("%Y-%m-%d") + "'"


# A dictionary of built-in tags to their replacement string functions.
#
# Duplicated in spirit by the SDK at `sdk/aqueduct/integrations/sql_integration.py`.
# Make sure the two are in sync.
BUILT_IN_EXPANSIONS = {
    "today": replace_today,
}


class RelationalParams(models.BaseParams):
    # The query cannot be used until `apply_placeholders()` is called on it. This flushes out
    # any user-defined tags like `{{today}}`.
    query_is_usable: Optional[bool] = False

    query: Optional[str] = None
    queries: Optional[List[str]] = None

    # TODO: Consider not including github as part of relational params when it is JSON marshalled
    github_metadata: Optional[Any]

    def _compile_chain(self, queries: List[str]) -> str:
        if not queries:
            return ""

        if len(queries) > 1:
            with_clause = "WITH\n"
            prev_table_name = ""
            for (idx, query) in enumerate(queries):
                # remove spaces and trailing semicolumns if any.
                normalized_query = query.strip().rstrip(";")

                # replace tag except for the first query
                if idx == 0:
                    if CHAIN_TABLE_TAG in normalized_query:
                        raise Exception(
                            f"Cannot compile chain. {CHAIN_TABLE_TAG} appears in the first query: {query}"
                        )
                else:
                    normalized_query = normalized_query.replace(CHAIN_TABLE_TAG, prev_table_name)

                # subquery goes to the 'WITH' clause except for the last one.
                if idx < len(queries) - 1:
                    cur_table_name = f"aqueduct_{uuid.uuid4().hex}"
                    with_clause += f"{cur_table_name} AS (\n{normalized_query}\n)"

                    # there are more subqueries to append in this 'WITH' clause
                    if idx < len(queries) - 2:
                        with_clause += ",\n"
                    prev_table_name = cur_table_name
                # otherwise, append the last query to 'WITH' clause
                else:
                    return f"{with_clause}\n{normalized_query}"

        return queries[0]

    def _expand_placeholders(self, query, parameters: Dict[str, str]) -> str:
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

    def compile(self, parameters: Dict[str, str]) -> None:
        assert (
            int(bool(self.query)) + int(bool(self.queries)) == 1
        ), "Exactly one of .query and .queries fields should be set."
        query = ""
        if bool(self.query):
            query = self.query
            print(f"Compiling query {query} .")
        else:
            print(f"Compiling chain queries {self.queries} .")
            query = self._compile_chain(self.queries)
            print(f"Compiled chain query is {query} .")

        query = self._expand_placeholders(query, parameters)
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


Params = Union[RelationalParams, S3Params]

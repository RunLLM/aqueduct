import re
from datetime import date
from typing import List, Optional, Union

import pandas as pd
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.preview import preview_artifact
from aqueduct.artifacts.table_artifact import TableArtifact
from aqueduct.constants.enums import ArtifactType, ExecutionMode, LoadUpdateMode, ServiceType
from aqueduct.error import InvalidUserActionException, InvalidUserArgumentException
from aqueduct.integrations.save import _save_artifact
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG
from aqueduct.models.integration import Integration, IntegrationInfo
from aqueduct.models.operators import (
    ExtractSpec,
    Operator,
    OperatorSpec,
    RelationalDBExtractParams,
    RelationalDBLoadParams,
)
from aqueduct.utils.dag_deltas import AddOperatorDelta, apply_deltas_to_dag
from aqueduct.utils.naming import default_artifact_name_from_op_name, sanitize_artifact_name
from aqueduct.utils.utils import generate_uuid

from aqueduct import globals

LIST_TABLES_QUERY_PG = "SELECT tablename, tableowner FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema';"
LIST_TABLES_QUERY_SNOWFLAKE = "SELECT table_name AS \"tablename\", table_owner AS \"tableowner\" FROM information_schema.tables WHERE table_schema != 'INFORMATION_SCHEMA' AND table_type = 'BASE TABLE';"
LIST_TABLES_QUERY_MYSQL = "SELECT table_name FROM INFORMATION_SCHEMA.TABLES WHERE table_type = 'BASE TABLE' AND table_schema NOT IN ('mysql', 'sys', 'performance_schema');"
LIST_TABLES_QUERY_MARIADB = "SELECT table_name AS \"tablename\" FROM INFORMATION_SCHEMA.TABLES WHERE table_type = 'BASE TABLE' AND table_schema NOT IN ('mysql', 'sys', 'performance_schema');"

LIST_TABLES_QUERY_SQLSERVER = (
    "SELECT table_name FROM INFORMATION_SCHEMA.TABLES WHERE table_type = 'BASE TABLE';"
)
GET_TABLE_QUERY = "select * from %s"
LIST_TABLES_QUERY_SQLITE = "SELECT name AS tablename FROM sqlite_master WHERE type='table';"
LIST_TABLES_QUERY_ATHENA = "AQUEDUCT_ATHENA_LIST_TABLE"

# Regular Expression that matches any substring appearance with
# "{{ }}" and a word inside with optional space in front or after
# Potential Matches: "{{today}}", "{{ today  }}""
TAG_PATTERN = r"{{\s*[\w-]+\s*}}"

# The TAG for 'previous table' when the user specifies a chained query.
PREV_TABLE_TAG = "$"


# A dictionary of built-in tags to their replacement string functions.
def replace_today() -> str:
    return "'" + date.today().strftime("%Y-%m-%d") + "'"


def find_parameter_names(query: str) -> List[str]:
    matches = re.findall(TAG_PATTERN, query)
    return [match.strip(" {}") for match in matches]


def find_parameter_artifacts(dag: DAG, names: List[str]) -> List[ArtifactMetadata]:
    """
    `find_parameter_artifacts` finds all parameter artifacts corresponding to given `names`.
    parameters:
        names: the list of names, repeating names are allowed.
    returns:
        a list of unique parameter artifacts for these names. Built-in names are omitted.

    raises: InvalidUserArgumentException if there's no parameter for the provided name.
    """
    artifacts = []
    for name in names:
        artf = dag.get_artifact_by_name(name)
        if artf is None:
            # If it is a built-in tag, we can ignore it for now, since the python operators will perform the expansion.
            if name in BUILT_IN_EXPANSIONS:
                continue

            raise InvalidUserArgumentException(
                "There is no parameter defined with name `%s`." % name,
            )
        artifacts.append(artf)

    return artifacts


# A dictionary of built-in tags to their replacement string functions.
BUILT_IN_EXPANSIONS = {
    "today": replace_today,
}


class RelationalDBIntegration(Integration):
    """
    Class for Relational integrations.
    """

    def __init__(self, dag: DAG, metadata: IntegrationInfo):
        self._dag = dag
        self._metadata = metadata

    def list_tables(self) -> pd.DataFrame:
        """
        Lists the tables available in the RelationalDB integration.

        Returns:
            pd.DataFrame of available tables.
        """
        if self.type() in [ServiceType.BIGQUERY, ServiceType.SNOWFLAKE]:
            # Use the list integration objects endpoint instead of
            # providing a hardcoded SQL query to execute
            tables = globals.__GLOBAL_API_CLIENT__.list_tables(str(self.id()))
            return pd.DataFrame(tables, columns=["tablename"])

        if self.type() in [
            ServiceType.POSTGRES,
            ServiceType.AQUEDUCTDEMO,
            ServiceType.REDSHIFT,
        ]:
            list_tables_query = LIST_TABLES_QUERY_PG
        elif self.type() == ServiceType.MYSQL:
            list_tables_query = LIST_TABLES_QUERY_MYSQL
        elif self.type() == ServiceType.MARIADB:
            list_tables_query = LIST_TABLES_QUERY_MARIADB
        elif self.type() == ServiceType.SQLSERVER:
            list_tables_query = LIST_TABLES_QUERY_SQLSERVER
        elif self.type() == ServiceType.SQLITE:
            list_tables_query = LIST_TABLES_QUERY_SQLITE
        elif self.type() == ServiceType.ATHENA:
            list_tables_query = LIST_TABLES_QUERY_ATHENA

        sql_artifact = self.sql(query=list_tables_query)
        return sql_artifact.get()

    def table(self, name: str) -> pd.DataFrame:
        """
        Retrieves a table from a RealtionalDB integration.

        Args:
            name:
                The name of the table to retrieve.

        Returns:
            pd.DataFrame of the table to retrieve.
        """
        sql_artifact = self.sql(query=GET_TABLE_QUERY % name)
        return sql_artifact.get()

    def sql(
        self,
        query: Union[str, List[str], RelationalDBExtractParams],
        name: Optional[str] = None,
        output: Optional[str] = None,
        description: str = "",
        lazy: bool = False,
    ) -> TableArtifact:
        """
        Runs a SQL query against the RelationalDB integration.

        Args:
            query:
                The query to run. When a list is provided, we run the list
                in a chain and return the result of the final query.
            name:
                Name of the query.
            output:
                Name to assign the output artifact. If not set, the default naming scheme will be used.
            description:
                Description of the query.
            lazy:
                Whether to run this operator lazily. See https://docs.aqueducthq.com/operators/lazy-vs.-eager-execution .

        Returns:
            TableArtifact representing result of the SQL query.
        """
        if globals.__GLOBAL_CONFIG__.lazy:
            lazy = True

        execution_mode = ExecutionMode.LAZY if lazy else ExecutionMode.EAGER

        op_name = name or "%s query" % self.name()
        artifact_name = output or default_artifact_name_from_op_name(op_name)

        extract_params = query
        if isinstance(extract_params, str):
            extract_params = RelationalDBExtractParams(
                query=extract_params,
            )
        elif isinstance(extract_params, list):
            for q in extract_params:
                assert isinstance(
                    q, str
                ), "When using a list of queries, it must be a list of strings."

            if len(extract_params) == 1:
                extract_params = RelationalDBExtractParams(
                    query=extract_params[0],
                )
            else:
                extract_params = RelationalDBExtractParams(
                    queries=extract_params,
                )
        elif isinstance(
            extract_params, RelationalDBExtractParams
        ):  # query is a RelationalDBExtractParams object
            if int(bool(extract_params.query)) + int(bool(extract_params.queries)) != 1:
                raise Exception(
                    "For a RelationalDBExtractParams object, exactly one of .query or .queries fields should be set."
                )

        # Find any tags that need to be expanded in the query, and add the parameters that correspond
        # to these tags as inputs to this operator. The orchestration engine will perform the replacement at runtime.
        sql_input_artifact_ids = []
        queries = []
        if extract_params.query is not None:
            queries = [extract_params.query]

        if extract_params.queries is not None:
            queries = extract_params.queries

        param_names = [name for q in queries for name in find_parameter_names(q)]
        param_artifacts = find_parameter_artifacts(self._dag, param_names)

        for artf in param_artifacts:
            # Check that the parameter corresponds to a string value.
            if artf.type != ArtifactType.STRING:
                raise InvalidUserArgumentException(
                    "The parameter `%s` must be defined as a string. Instead, got type %s"
                    % (artf.name, artf.type)
                )

            sql_input_artifact_ids.append(artf.id)

        sql_operator_id = generate_uuid()
        sql_output_artifact_id = generate_uuid()
        apply_deltas_to_dag(
            self._dag,
            deltas=[
                AddOperatorDelta(
                    op=Operator(
                        id=sql_operator_id,
                        name=op_name,
                        description=description,
                        spec=OperatorSpec(
                            extract=ExtractSpec(
                                service=self.type(),
                                integration_id=self.id(),
                                parameters=extract_params,
                            )
                        ),
                        inputs=sql_input_artifact_ids,
                        outputs=[sql_output_artifact_id],
                    ),
                    output_artifacts=[
                        ArtifactMetadata(
                            id=sql_output_artifact_id,
                            name=sanitize_artifact_name(artifact_name),
                            type=ArtifactType.TABLE,
                            explicitly_named=output is not None,
                        ),
                    ],
                ),
            ],
        )

        if execution_mode == ExecutionMode.EAGER:
            # Issue preview request since this is an eager execution.
            artifact = preview_artifact(self._dag, sql_output_artifact_id)
            assert isinstance(artifact, TableArtifact)
            return artifact
        else:
            # We are in lazy mode.
            return TableArtifact(self._dag, sql_output_artifact_id)

    def save(self, artifact: BaseArtifact, table_name: str, update_mode: LoadUpdateMode) -> None:
        """Registers a save operator of the given artifact, to be executed when it's computed in a published flow.

        Args:
            artifact:
                The artifact to save into this sql integration.
            table_name:
                The table to save the artifact to.
            update_mode:
                Defines the semantics of the save if a table already exists.
                Options are "replace", "append" (row-wise), or "fail" (if table already exists).
        """
        if self.type() == ServiceType.ATHENA:
            raise InvalidUserActionException(
                "Save operation not supported for %s." % self.type().value
            )
        # Non-tabular data cannot be saved into relational data stores.
        if artifact.type() not in [ArtifactType.UNTYPED, ArtifactType.TABLE]:
            raise InvalidUserActionException(
                "Unable to save non-relational data into relational data store `%s`." % self.name()
            )

        _save_artifact(
            artifact.id(),
            self._dag,
            self._metadata,
            save_params=RelationalDBLoadParams(table=table_name, update_mode=update_mode),
        )

    def describe(self) -> None:
        """
        Prints out a human-readable description of the SQL integration.
        """
        print("==================== SQL Integration =============================")
        print("Integration Information:")
        self._metadata.describe()
        print("Integration Table List Preview:")
        print(self.list_tables()["name"].head().to_string())
        print("(only first 5 tables are shown)")

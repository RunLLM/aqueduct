import re
from datetime import date
from typing import List, Optional, Union

import pandas as pd
from aqueduct.artifacts import utils as artifact_utils
from aqueduct.artifacts.metadata import ArtifactMetadata
from aqueduct.artifacts.table_artifact import TableArtifact
from aqueduct.dag import DAG
from aqueduct.dag_deltas import AddOrReplaceOperatorDelta, apply_deltas_to_dag
from aqueduct.enums import ArtifactType, ExecutionMode, LoadUpdateMode, ServiceType
from aqueduct.error import InvalidUserActionException, InvalidUserArgumentException
from aqueduct.integrations.integration import Integration, IntegrationInfo
from aqueduct.operators import (
    ExtractSpec,
    Operator,
    OperatorSpec,
    RelationalDBExtractParams,
    RelationalDBLoadParams,
    SaveConfig,
)
from aqueduct.utils import artifact_name_from_op_name, generate_uuid

from aqueduct import globals

LIST_TABLES_QUERY_PG = "SELECT tablename, tableowner FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema';"
LIST_TABLES_QUERY_SNOWFLAKE = "SELECT table_name AS \"tablename\", table_owner AS \"tableowner\" FROM information_schema.tables WHERE table_schema != 'INFORMATION_SCHEMA' AND table_type = 'BASE TABLE';"
LIST_TABLES_QUERY_MYSQL = "SELECT table_name FROM INFORMATION_SCHEMA.TABLES WHERE table_type = 'BASE TABLE' AND table_schema NOT IN ('mysql', 'sys', 'performance_schema');"
LIST_TABLES_QUERY_SQLSERVER = (
    "SELECT table_name FROM INFORMATION_SCHEMA.TABLES WHERE table_type = 'BASE TABLE';"
)
LIST_TABLES_QUERY_BIGQUERY = "SELECT schema_name FROM information_schema.schemata;"
GET_TABLE_QUERY = "select * from %s"
LIST_TABLES_QUERY_SQLITE = "SELECT name FROM sqlite_master WHERE type='table';"
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
    Class for RealtionalDB integrations.
    """

    def __init__(self, dag: DAG, metadata: IntegrationInfo):
        self._dag = dag
        self._metadata = metadata

    def list_tables(self) -> pd.DataFrame:
        """
        Lists the tables available in the RealtionalDB integration.

        Returns:
            pd.DataFrame of available tables.
        """
        if self._metadata.service in [
            ServiceType.POSTGRES,
            ServiceType.AQUEDUCTDEMO,
            ServiceType.REDSHIFT,
        ]:
            list_tables_query = LIST_TABLES_QUERY_PG
        elif self._metadata.service == ServiceType.SNOWFLAKE:
            list_tables_query = LIST_TABLES_QUERY_SNOWFLAKE
        elif self._metadata.service in [ServiceType.MYSQL, ServiceType.MARIADB]:
            list_tables_query = LIST_TABLES_QUERY_MYSQL
        elif self._metadata.service == ServiceType.SQLSERVER:
            list_tables_query = LIST_TABLES_QUERY_SQLSERVER
        elif self._metadata.service == ServiceType.BIGQUERY:
            list_tables_query = LIST_TABLES_QUERY_BIGQUERY
        elif self._metadata.service == ServiceType.SQLITE:
            list_tables_query = LIST_TABLES_QUERY_SQLITE
        elif self._metadata.service == ServiceType.ATHENA:
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
            description:
                Description of the query.
            lazy:
                Whether to run this operator lazily. See https://docs.aqueducthq.com/operators/lazy-vs.-eager-execution .

        Returns:
            TableArtifact representing result of the SQL query.
        """
        if globals.__GLOBAL_CONFIG__.lazy:
            lazy = True
        execution_mode = ExecutionMode.EAGER if not lazy else ExecutionMode.LAZY

        integration_info = self._metadata

        # The sql operator name defaults to "[integration name] query 1". If another
        # sql operator already exists with that name, we'll continue bumping the suffix
        # until the sql operator is unique. If an explicit name is provided, we will
        # overwrite the existing one.
        sql_op_name = name
        if sql_op_name is None:
            sql_op_name = self._dag.get_unclaimed_op_name(prefix="%s query" % integration_info.name)

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
                AddOrReplaceOperatorDelta(
                    op=Operator(
                        id=sql_operator_id,
                        name=sql_op_name,
                        description=description,
                        spec=OperatorSpec(
                            extract=ExtractSpec(
                                service=integration_info.service,
                                integration_id=integration_info.id,
                                parameters=extract_params,
                            )
                        ),
                        inputs=sql_input_artifact_ids,
                        outputs=[sql_output_artifact_id],
                    ),
                    output_artifacts=[
                        ArtifactMetadata(
                            id=sql_output_artifact_id,
                            name=artifact_name_from_op_name(sql_op_name),
                            type=ArtifactType.TABLE,
                        ),
                    ],
                ),
            ],
        )

        if execution_mode == ExecutionMode.EAGER:
            # Issue preview request since this is an eager execution.
            artifact = artifact_utils.preview_artifact(self._dag, sql_output_artifact_id)
            assert isinstance(artifact, TableArtifact)
            return artifact
        else:
            # We are in lazy mode.
            return TableArtifact(self._dag, sql_output_artifact_id)

    def config(self, table: str, update_mode: LoadUpdateMode) -> SaveConfig:
        """
        Configuration for saving to RelationalDB Integration.

        Arguments:
            table:
                Table to save to.
            update_mode:
                The update mode to use when saving the artifact as a relational table.
                Possible values are: APPEND, REPLACE, or FAIL.
        Returns:
            SaveConfig object to use in TableArtifact.save()
        """
        if self._metadata.service == ServiceType.ATHENA:
            raise InvalidUserActionException(
                "Save operation is not supported for integration type %s."
                % self._metadata.service.value
            )

        return SaveConfig(
            integration_info=self._metadata,
            parameters=RelationalDBLoadParams(table=table, update_mode=update_mode),
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

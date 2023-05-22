from typing import List, Optional, Union

import pandas as pd
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.preview import preview_artifact
from aqueduct.artifacts.table_artifact import TableArtifact
from aqueduct.constants.enums import ArtifactType, ExecutionMode, LoadUpdateMode, ServiceType
from aqueduct.error import InvalidUserActionException, InvalidUserArgumentException
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG
from aqueduct.models.operators import (
    ExtractSpec,
    Operator,
    OperatorSpec,
    RelationalDBExtractParams,
    RelationalDBLoadParams,
)
from aqueduct.models.resource import BaseResource, ResourceInfo
from aqueduct.resources.parameters import (
    _validate_artifact_is_string,
    _validate_builtin_expansions,
    _validate_parameters,
)
from aqueduct.resources.save import _save_artifact
from aqueduct.resources.validation import validate_is_connected
from aqueduct.utils.dag_deltas import AddOperatorDelta, apply_deltas_to_dag
from aqueduct.utils.naming import default_artifact_name_from_op_name, sanitize_artifact_name
from aqueduct.utils.utils import generate_uuid

from aqueduct import globals

LIST_TABLES_QUERY_PG = "SELECT tablename, tableowner FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema';"
LIST_TABLES_QUERY_SNOWFLAKE = "SELECT table_name AS \"tablename\", table_owner AS \"tableowner\" FROM information_schema.tables WHERE table_schema != 'INFORMATION_SCHEMA' AND table_type = 'BASE TABLE';"
LIST_TABLES_QUERY_MYSQL = "SELECT table_name AS tablename FROM INFORMATION_SCHEMA.TABLES WHERE table_type = 'BASE TABLE' AND table_schema NOT IN ('mysql', 'sys', 'performance_schema');"
LIST_TABLES_QUERY_MARIADB = "SELECT table_name AS \"tablename\" FROM INFORMATION_SCHEMA.TABLES WHERE table_type = 'BASE TABLE' AND table_schema NOT IN ('mysql', 'sys', 'performance_schema');"

LIST_TABLES_QUERY_SQLSERVER = (
    "SELECT table_name FROM INFORMATION_SCHEMA.TABLES WHERE table_type = 'BASE TABLE';"
)
GET_TABLE_QUERY = "select * from %s"
LIST_TABLES_QUERY_SQLITE = "SELECT name AS tablename FROM sqlite_master WHERE type='table';"
LIST_TABLES_QUERY_ATHENA = "AQUEDUCT_ATHENA_LIST_TABLE"


class RelationalDBResource(BaseResource):
    """
    Class for Relational resources.
    """

    def __init__(self, dag: DAG, metadata: ResourceInfo):
        self._dag = dag
        self._metadata = metadata

    @validate_is_connected()
    def list_tables(self) -> pd.DataFrame:
        """
        Lists the tables available in the RelationalDB resource.

        Returns:
            pd.DataFrame of available tables.
        """

        if self.type() in [ServiceType.BIGQUERY, ServiceType.SNOWFLAKE]:
            # Use the list resource objects endpoint instead of
            # providing a hardcoded SQL query to execute
            tables = globals.__GLOBAL_API_CLIENT__.list_tables(str(self.id()))
            return pd.DataFrame(tables, columns=["tablename"])

        if self.type() in [
            ServiceType.POSTGRES,
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

    @validate_is_connected()
    def table(self, name: str) -> pd.DataFrame:
        """
        Retrieves a table from a RelationalDB resource.

        Args:
            name:
                The name of the table to retrieve.

        Returns:
            pd.DataFrame of the table to retrieve.
        """
        sql_artifact = self.sql(query=GET_TABLE_QUERY % name)
        return sql_artifact.get()

    @validate_is_connected()
    def sql(
        self,
        query: Union[str, List[str], RelationalDBExtractParams],
        name: Optional[str] = None,
        output: Optional[str] = None,
        description: str = "",
        parameters: Optional[List[BaseArtifact]] = None,
        lazy: bool = False,
    ) -> TableArtifact:
        """
        Runs a SQL query against the RelationalDB resource.

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
            parameters:
                An optional list of string parameters to use in the query. We use the Postgres syntax of $1, $2 for placeholders.
                The number denotes which parameter in the list to use (one-indexed). These parameters feed into the
                sql query operator and will fill in the placeholders in the query with the actual values.

                For example, for the following query with parameters=[param1, param2]:
                    SELECT * FROM my_table where age = $1 and name = $2
                Assuming default values of "18" and "John" respectively, the default query will expand into
                    SELECT * FROM my_table where age = 18 and name = "John".

                If multiple of the same placeholders are used in the same query, the same value will be supplied for each.
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
        assert isinstance(extract_params, RelationalDBExtractParams)
        assert (
            extract_params.query
            or extract_params.queries
            and not (extract_params.query and extract_params.queries)
        )

        # Perform validations on the query.
        queries: Optional[List[str]] = (
            [extract_params.query] if extract_params.query is not None else extract_params.queries
        )
        assert isinstance(queries, list)
        _validate_builtin_expansions(queries)

        sql_input_artifact_ids = []
        if parameters is not None:
            if not isinstance(parameters, list) and any(
                not isinstance(param, BaseArtifact) for param in parameters
            ):
                raise InvalidUserArgumentException(
                    "`parameters` argument must be a list of artifacts."
                )
            _validate_parameters(queries, parameters)
            sql_input_artifact_ids = [param.id() for param in parameters]

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
                                resource_id=self.id(),
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

    @validate_is_connected()
    def save(
        self,
        artifact: BaseArtifact,
        table_name: Union[str, BaseArtifact],
        update_mode: LoadUpdateMode,
    ) -> None:
        """Registers a save operator of the given artifact, to be executed when it's computed in a published flow.

        Args:
            artifact:
                The artifact to save into this sql resource.
            table_name:
                The table to save the artifact to. You can also parameterize this field by passing
                a string parameter here. When this save is parameterized, the table name parameter
                will always be ordered before the artifact in the save operator's input list.
            update_mode:
                Defines the semantics of the save if a table already exists.
                Options are "replace", "append" (row-wise), or "fail" (if table already exists).
        """
        if self.type() == ServiceType.ATHENA:
            raise InvalidUserActionException(
                "Save operation not supported for %s." % self.type().value
            )

        if not isinstance(table_name, str) and not isinstance(table_name, BaseArtifact):
            raise InvalidUserArgumentException(
                "`table_name` must either be a string or a string parameter artifact."
            )

        # Non-tabular data cannot be saved into relational data stores.
        if artifact.type() not in [ArtifactType.UNTYPED, ArtifactType.TABLE]:
            raise InvalidUserActionException(
                "Unable to save non-relational data into relational data store `%s`." % self.name()
            )

        # Resolve the table name if it is parameterized.
        table_name_str = table_name
        artifact_ids = [artifact.id()]
        if isinstance(table_name, BaseArtifact):
            table_name_artifact = table_name
            _validate_artifact_is_string(table_name_artifact)

            # This is unset in the LoadParams, since we're parameterizing it.
            table_name_str = ""

            # Assumption: All parameter artifacts are prepended to the operator's input list.
            artifact_ids = [table_name_artifact.id()] + artifact_ids
        else:
            if table_name_str == "":
                raise InvalidUserArgumentException("Cannot save to an empty table name.")

        _save_artifact(
            artifact_ids,
            self._dag,
            self._metadata,
            save_params=RelationalDBLoadParams(table=table_name_str, update_mode=update_mode),
        )

    def describe(self) -> None:
        """
        Prints out a human-readable description of the SQL resource.
        """
        print("==================== SQL Resource =============================")
        print("Resource Information:")
        self._metadata.describe()

        # Only list the tables if the resource is connected.
        try:
            print("Resource Table List Preview:")
            print(self.list_tables()["name"].head().to_string())
            print("(only first 5 tables are shown)")
        except:
            pass

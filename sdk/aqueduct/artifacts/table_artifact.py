from __future__ import annotations

import json
import uuid
from typing import Any, Dict, List, Optional, Union

import pandas as pd
from aqueduct.artifacts import bool_artifact, numeric_artifact
from aqueduct.artifacts import utils as artifact_utils
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.metadata import ArtifactMetadata
from aqueduct.constants.metrics import SYSTEM_METRICS_INFO
from aqueduct.dag import DAG
from aqueduct.dag_deltas import (
    AddOrReplaceOperatorDelta,
    RemoveCheckOperatorDelta,
    apply_deltas_to_dag,
)
from aqueduct.enums import (
    ArtifactType,
    CheckSeverity,
    ExecutionMode,
    FunctionGranularity,
    FunctionType,
    OperatorType,
)
from aqueduct.error import AqueductError, ArtifactNeverComputedException
from aqueduct.operators import (
    CheckSpec,
    FunctionSpec,
    MetricSpec,
    Operator,
    OperatorSpec,
    SystemMetricSpec,
)
from aqueduct.utils import (
    artifact_name_from_op_name,
    format_header_for_print,
    generate_uuid,
    get_checks_for_op,
    get_description_for_check,
    get_description_for_metric,
    serialize_function,
)
from ruamel import yaml

from aqueduct import globals


class TableArtifact(BaseArtifact):
    """This class represents a computed table within the flow's DAG.

    Any `@op`-annotated python function that returns a dataframe will
    return this class when that function is called.

    Any SQL query will also return this class.

    Examples:
        >>> @op
        >>> def predict(df):
        >>>     return predictions
        >>>
        >>> output_artifact = predict(input_artifact)

        The contents of these artifacts can be manifested locally or written to an
        integration:

        >>> df = output_artifact.get()
        >>> print(df.head())
        >>> output_artifact.save(warehouse.config(table_name="output_table"))
    """

    def __init__(
        self,
        dag: DAG,
        artifact_id: uuid.UUID,
        content: Optional[pd.DataFrame] = None,
        from_flow_run: bool = False,
    ):
        self._dag = dag
        self._artifact_id = artifact_id

        # This parameter indicates whether the artifact is fetched from flow-run or not.
        self._from_flow_run = from_flow_run
        self._set_content(content)

    def get(self, parameters: Optional[Dict[str, Any]] = None) -> pd.DataFrame:
        """Materializes TableArtifact into an actual dataframe.

        Args:
            parameters:
                A map from parameter name to its custom value, to be used when evaluating
                this artifact.

        Returns:
            A dataframe containing the table contents of this artifact.

        Raises:
            InvalidRequestError:
                An error occurred because of an issue with the user's code or inputs.
            InternalServerError:
                An unexpected error occurred within the Aqueduct cluster.
        """
        self._dag.must_get_artifact(self._artifact_id)

        if self._from_flow_run:
            if self._get_content() is None:
                raise ArtifactNeverComputedException(
                    "This artifact was part of an existing flow run but was never computed successfully!",
                )
            elif parameters is not None:
                raise NotImplementedError(
                    "Parameterizing historical artifacts is not currently supported."
                )

        content = self._get_content()
        if parameters is not None or content is None:
            previewed_artifact = artifact_utils.preview_artifact(
                self._dag, self._artifact_id, parameters
            )
            content = previewed_artifact._get_content()

            # If the artifact was previously generated lazily, materialize the contents.
            if parameters is None and self._get_content() is None:
                self._set_content(content)

        assert isinstance(content, pd.DataFrame)
        return content

    def head(self, n: int = 5, parameters: Optional[Dict[str, Any]] = None) -> pd.DataFrame:
        """Returns a preview of the table artifact.

        >>> db = client.integration(name="demo/")
        >>> customer_data = db.sql("SELECT * from customers")
        >>> churn_predictions = predict_churn(customer_data)
        >>> churn_predictions.head()

        Args:
            n:
                the number of row previewed. Default to 5.
        Returns:
            A dataframe containing the table contents of this artifact.
        """
        df = self.get(parameters=parameters)
        return df.head(n)

    PRESET_METRIC_LIST = ["number_of_missing_values", "number_of_rows", "max", "min", "mean", "std"]

    def list_preset_metrics(self) -> List[str]:
        """Returns a list of all preset metrics available on the table artifact.
        These preset metrics can be set via the invoking the preset method on a artifact.

        Returns:
            A list of available preset metrics on a table
        """
        return self.PRESET_METRIC_LIST

    def list_system_metrics(self) -> List[str]:
        """Returns a list of all system metrics available on the table artifact.
        These system metrics can be set via the invoking the system_metric() method the table.

        Returns:
            A list of available system metrics on a table
        """
        return list(SYSTEM_METRICS_INFO.keys())

    def validate_with_expectation(
        self,
        expectation_name: str,
        expectation_args: Optional[Dict[str, Any]] = None,
        severity: CheckSeverity = CheckSeverity.WARNING,
        lazy: bool = False,
    ) -> bool_artifact.BoolArtifact:
        """Creates a check that validates with the table with great_expectations and its set of internal expectations.
        The expectations supported can be found here:
        https://great-expectations.readthedocs.io/en/latest/reference/glossary_of_expectations.html
        The expectations supported are only those that can support Pandas on the great_expectations backend.
        E.g. Use a expectation to check all column values are unique.
        ge_check = table.validate_with_expectation("expect_column_values_to_be_unique", {"column": "fixed_acidity"})
        ge_check.get() // True or False based on expectation passing

        Args:
            expectation_name:
                Name of built-in expectation to run with great_expectations
            expectation_args:
                Dictionary of args to pass into the expectation suite for the expectation being run.
            severity:
                Optional severity associated with the check created with this expectations

        Returns:
            A bool artifact that represent the validation result of running the expectation provided on the table.
        """
        if globals.__GLOBAL_CONFIG__.lazy:
            lazy = True
        execution_mode = ExecutionMode.EAGER if not lazy else ExecutionMode.LAZY

        # We import on demand since this takes 1-2 seconds, and we don't want to incur that every time
        # we reference the SDK.
        from great_expectations.core import ExpectationConfiguration
        from great_expectations.core.batch import RuntimeBatchRequest
        from great_expectations.data_context import BaseDataContext
        from great_expectations.data_context.types.base import (
            DataContextConfig,
            DatasourceConfig,
            FilesystemStoreBackendDefaults,
        )
        from great_expectations.validator.validator import Validator

        def great_expectations_check_method(table: pd.DataFrame) -> bool:
            data_context_config = DataContextConfig(
                datasources={
                    "my_pandas_datasource": DatasourceConfig(
                        class_name="PandasDatasource",
                    )
                },
                store_backend_defaults=FilesystemStoreBackendDefaults(root_directory="/tmp"),
            )

            context = BaseDataContext(project_config=data_context_config)

            # Create and load Expectation Suite
            expectation_suite_name = "aq_check_expectation"
            suite = context.create_expectation_suite(
                expectation_suite_name=expectation_suite_name, overwrite_existing=True
            )
            expectation_config = ExpectationConfiguration(expectation_name, expectation_args)
            suite.add_expectation(expectation_config)

            # We create a custom datasource that will allow us to load a in-memory dataframe to validate the expectation quite
            datasource_yaml = f"""
            name: my_pandas_datasource
            class_name: Datasource
            module_name: great_expectations.datasource
            execution_engine:
                module_name: great_expectations.execution_engine
                class_name: PandasExecutionEngine
            data_connectors:
                aq_runtime_dataconnecter:
                    class_name: RuntimeDataConnector
                    batch_identifiers:
                        - default_identifier_name
            """
            context.add_datasource(**yaml.load(datasource_yaml, Loader=yaml.Loader))

            df: pd.DataFrame = table
            runtime_batch_request = RuntimeBatchRequest(
                datasource_name="my_pandas_datasource",
                data_connector_name="aq_runtime_dataconnecter",
                data_asset_name="aq_table_check",
                runtime_parameters={"batch_data": df},
                batch_identifiers={"default_identifier_name": "aq_table_check"},
            )

            # Constructing Validator by passing in RuntimeBatchRequest
            my_validator: Validator = context.get_validator(
                batch_request=runtime_batch_request,
                expectation_suite=suite,
            )

            # Run validation to return the result
            result = my_validator.validate()
            return bool(result.success)

        check_name = "ge_table_check: {%s}" % expectation_name

        zip_file = serialize_function(great_expectations_check_method, check_name)
        function_spec = FunctionSpec(
            type=FunctionType.FILE,
            granularity=FunctionGranularity.TABLE,
            file=zip_file,
        )
        check_spec = OperatorSpec(check=CheckSpec(level=severity, function=function_spec))
        check_description = "Check table with built in expectations from great expectations"
        new_artifact = self._apply_operator_to_table(
            check_spec,
            check_name,
            check_description,
            output_artifact_type_hint=ArtifactType.BOOL,
            execution_mode=execution_mode,
        )

        assert isinstance(new_artifact, bool_artifact.BoolArtifact)

        return new_artifact

    def number_of_missing_values(
        self,
        column_id: Any = None,
        row_id: Any = None,
        lazy: bool = False,
    ) -> numeric_artifact.NumericArtifact:
        """Creates a metric that represents the number of missing values over a given column or row.

        Note: takes a scalar column_id/row_id and uses pandas.DataFrame.isnull() to compute value.

        Args:
            column_id:
                column identifier to find missing values for
            row_id:
                row identifier to find missing values for

        Returns:
            A numeric artifact that represents the number of missing values for the row/column on the applied table artifact.
        """
        if globals.__GLOBAL_CONFIG__.lazy:
            lazy = True
        execution_mode = ExecutionMode.EAGER if not lazy else ExecutionMode.LAZY

        table_name = self._get_table_name()
        if column_id is not None and row_id is not None:
            raise AqueductError(
                "Cannot choose both a row and a column for counting missing values over"
            )

        if column_id is None and row_id is None:
            raise AqueductError("Specify either a row or column for counting missing values over")

        if column_id is not None:

            def interal_num_missing_val_col(table: pd.DataFrame) -> float:
                return float(table[column_id].isnull().sum())

            metric_func = interal_num_missing_val_col
            metric_name = "num_col_missing_val(%s)" % column_id
            metric_description = "compute number of missing values for col %s on %s" % (
                column_id,
                table_name,
            )
        else:

            def interal_num_missing_val_row(table: pd.DataFrame) -> float:
                return float(table.loc[row_id].isnull().sum())

            metric_func = interal_num_missing_val_row
            metric_name = "num_row_missing_val(%s)" % row_id
            metric_description = "compute number of missing values for row %s on %s" % (
                column_id,
                table_name,
            )

        zip_file = serialize_function(metric_func, metric_name)

        function_spec = FunctionSpec(
            type=FunctionType.FILE,
            granularity=FunctionGranularity.TABLE,
            file=zip_file,
        )
        op_spec = OperatorSpec(metric=MetricSpec(function=function_spec))
        new_artifact = self._apply_operator_to_table(
            op_spec,
            metric_name,
            metric_description,
            output_artifact_type_hint=ArtifactType.NUMERIC,
            execution_mode=execution_mode,
        )

        assert isinstance(new_artifact, numeric_artifact.NumericArtifact)

        return new_artifact

    def number_of_rows(self, lazy: bool = False) -> numeric_artifact.NumericArtifact:
        """Creates a metric that represents the number of rows of this table

        Note: uses len() to determine row count over the pandas.DataFrame.

        Returns:
            A numeric artifact that represents the number of rows on this table.
        """
        if globals.__GLOBAL_CONFIG__.lazy:
            lazy = True
        execution_mode = ExecutionMode.EAGER if not lazy else ExecutionMode.LAZY

        table_name = self._get_table_name()

        def internal_num_rows_metric(table: pd.DataFrame) -> float:
            return float(len(table))

        metric_name = "num_rows"
        metric_description = "compute number of rows for %s" % table_name
        zip_file = serialize_function(internal_num_rows_metric, metric_name)

        function_spec = FunctionSpec(
            type=FunctionType.FILE,
            granularity=FunctionGranularity.TABLE,
            file=zip_file,
        )
        op_spec = OperatorSpec(metric=MetricSpec(function=function_spec))
        new_artifact = self._apply_operator_to_table(
            op_spec,
            metric_name,
            metric_description,
            output_artifact_type_hint=ArtifactType.NUMERIC,
            execution_mode=execution_mode,
        )

        assert isinstance(new_artifact, numeric_artifact.NumericArtifact)

        return new_artifact

    def max(self, column_id: Any, lazy: bool = False) -> numeric_artifact.NumericArtifact:
        """Creates a metric that represents the maximum value over the given column

        Note: takes a scalar column_id and uses pandas.DataFrame.max to compute value.

        Args:
            column_id:
                column identifier to find max of

        Returns:
            A numeric artifact that represents the max for the given column on the applied table artifact.
        """
        if globals.__GLOBAL_CONFIG__.lazy:
            lazy = True
        execution_mode = ExecutionMode.EAGER if not lazy else ExecutionMode.LAZY

        table_name = self._get_table_name()

        def internal_max_metric(table: pd.DataFrame) -> float:
            return float(table[column_id].max())

        metric_name = "max(%s)" % column_id
        metric_description = "Max for column %s for %s" % (
            column_id,
            table_name,
        )
        zip_file = serialize_function(internal_max_metric, metric_name)

        function_spec = FunctionSpec(
            type=FunctionType.FILE,
            granularity=FunctionGranularity.TABLE,
            file=zip_file,
        )
        op_spec = OperatorSpec(metric=MetricSpec(function=function_spec))
        new_artifact = self._apply_operator_to_table(
            op_spec,
            metric_name,
            metric_description,
            output_artifact_type_hint=ArtifactType.NUMERIC,
            execution_mode=execution_mode,
        )

        assert isinstance(new_artifact, numeric_artifact.NumericArtifact)

        return new_artifact

    def min(self, column_id: Any, lazy: bool = False) -> numeric_artifact.NumericArtifact:
        """Creates a metric that represents the minimum value over the given column

        Note: takes a scalar column_id and uses pandas.DataFrame.min to compute value.

        Args:
            column_id:
                column identifier to find min of

        Returns:
            A numeric artifact that represents the min for the given column on the applied table artifact.
        """
        if globals.__GLOBAL_CONFIG__.lazy:
            lazy = True
        execution_mode = ExecutionMode.EAGER if not lazy else ExecutionMode.LAZY

        table_name = self._get_table_name()

        def internal_min_metric(table: pd.DataFrame) -> float:
            return float(table[column_id].min())

        metric_name = "min(%s)" % column_id
        metric_description = "Min for column %s for %s" % (
            column_id,
            table_name,
        )
        zip_file = serialize_function(internal_min_metric, metric_name)

        function_spec = FunctionSpec(
            type=FunctionType.FILE,
            granularity=FunctionGranularity.TABLE,
            file=zip_file,
        )
        op_spec = OperatorSpec(metric=MetricSpec(function=function_spec))
        new_artifact = self._apply_operator_to_table(
            op_spec,
            metric_name,
            metric_description,
            output_artifact_type_hint=ArtifactType.NUMERIC,
            execution_mode=execution_mode,
        )

        assert isinstance(new_artifact, numeric_artifact.NumericArtifact)

        return new_artifact

    def mean(self, column_id: Any, lazy: bool = False) -> numeric_artifact.NumericArtifact:
        """Creates a metric that represents the mean value over the given column

        Note: takes a scalar column_id and uses pandas.DataFrame.mean to compute value.

        Args:
            column_id:
                column identifier to compute mean of

        Returns:
            A numeric artifact that represents the mean for the given column on the applied table artifact.
        """
        if globals.__GLOBAL_CONFIG__.lazy:
            lazy = True
        execution_mode = ExecutionMode.EAGER if not lazy else ExecutionMode.LAZY

        table_name = self._get_table_name()

        def internal_mean_metric(table: pd.DataFrame) -> float:
            return float(table[column_id].mean())

        metric_name = "mean(%s)" % column_id
        metric_description = "Mean for column %s for %s" % (
            column_id,
            table_name,
        )
        zip_file = serialize_function(internal_mean_metric, metric_name)

        function_spec = FunctionSpec(
            type=FunctionType.FILE,
            granularity=FunctionGranularity.TABLE,
            file=zip_file,
        )
        op_spec = OperatorSpec(metric=MetricSpec(function=function_spec))
        new_artifact = self._apply_operator_to_table(
            op_spec,
            metric_name,
            metric_description,
            output_artifact_type_hint=ArtifactType.NUMERIC,
            execution_mode=execution_mode,
        )

        assert isinstance(new_artifact, numeric_artifact.NumericArtifact)

        return new_artifact

    def std(self, column_id: Any, lazy: bool = False) -> numeric_artifact.NumericArtifact:
        """Creates a metric that represents the standard deviation value over the given column

        takes a scalar column_id and uses pandas.DataFrame.std to compute value

        Args:
            column_id:
                column identifier to compute standard deviation of

        Returns:
            A numeric artifact that represents the standard deviation for the given column on the applied table artifact.
        """
        if globals.__GLOBAL_CONFIG__.lazy:
            lazy = True
        execution_mode = ExecutionMode.EAGER if not lazy else ExecutionMode.LAZY

        table_name = self._get_table_name()

        def internal_std_metric(table: pd.DataFrame) -> float:
            std: float
            std = table[column_id].std()
            return std

        metric_name = "std(%s)" % column_id
        metric_description = "std for column %s for %s" % (
            column_id,
            table_name,
        )
        zip_file = serialize_function(internal_std_metric, metric_name)

        function_spec = FunctionSpec(
            type=FunctionType.FILE,
            granularity=FunctionGranularity.TABLE,
            file=zip_file,
        )
        op_spec = OperatorSpec(metric=MetricSpec(function=function_spec))
        new_artifact = self._apply_operator_to_table(
            op_spec,
            metric_name,
            metric_description,
            output_artifact_type_hint=ArtifactType.NUMERIC,
            execution_mode=execution_mode,
        )

        assert isinstance(new_artifact, numeric_artifact.NumericArtifact)

        return new_artifact

    def system_metric(
        self, metric_name: str, lazy: bool = False
    ) -> numeric_artifact.NumericArtifact:
        """Creates a system metric that represents the given system information from the previous @op that ran on the table.

        Args:
            metric_name:
                name of system metric to retrieve for the table.
                valid metrics are:
                    runtime: runtime of previous @op func in seconds
                    max_memory: maximum memory usage of previous @op func in Mb

        Returns:
            A numeric artifact that represents the requested system metric
        """
        if globals.__GLOBAL_CONFIG__.lazy:
            lazy = True
        execution_mode = ExecutionMode.EAGER if not lazy else ExecutionMode.LAZY

        operator = self._dag.must_get_operator(with_output_artifact_id=self._artifact_id)
        system_metric_description, system_metric_unit = SYSTEM_METRICS_INFO[metric_name]
        system_metric_name = "%s %s(%s) metric" % (operator.name, metric_name, system_metric_unit)
        op_spec = OperatorSpec(system_metric=SystemMetricSpec(metric_name=metric_name))
        new_artifact = self._apply_operator_to_table(
            op_spec,
            system_metric_name,
            system_metric_description,
            output_artifact_type_hint=ArtifactType.NUMERIC,
            execution_mode=execution_mode,
        )

        assert isinstance(new_artifact, numeric_artifact.NumericArtifact)

        return new_artifact

    def _apply_operator_to_table(
        self,
        op_spec: OperatorSpec,
        op_name: str,
        op_description: str,
        output_artifact_type_hint: ArtifactType,
        execution_mode: ExecutionMode = ExecutionMode.EAGER,
    ) -> Union[numeric_artifact.NumericArtifact, bool_artifact.BoolArtifact]:
        assert (
            output_artifact_type_hint == ArtifactType.NUMERIC
            or output_artifact_type_hint == ArtifactType.BOOL
        )

        operator_id = generate_uuid()
        output_artifact_id = generate_uuid()

        apply_deltas_to_dag(
            self._dag,
            deltas=[
                AddOrReplaceOperatorDelta(
                    op=Operator(
                        id=operator_id,
                        name=op_name,
                        description=op_description,
                        spec=op_spec,
                        inputs=[self._artifact_id],
                        outputs=[output_artifact_id],
                    ),
                    output_artifacts=[
                        ArtifactMetadata(
                            id=output_artifact_id,
                            name=artifact_name_from_op_name(op_name),
                            type=output_artifact_type_hint,
                        )
                    ],
                ),
            ],
        )

        if execution_mode == ExecutionMode.EAGER:
            # Issue preview request since this is an eager execution.
            artifact = artifact_utils.preview_artifact(self._dag, output_artifact_id)

            assert isinstance(artifact, numeric_artifact.NumericArtifact) or isinstance(
                artifact, bool_artifact.BoolArtifact
            )
            return artifact
        else:
            # We are in lazy mode.
            if output_artifact_type_hint == ArtifactType.NUMERIC:
                return numeric_artifact.NumericArtifact(self._dag, output_artifact_id)
            else:
                return bool_artifact.BoolArtifact(self._dag, output_artifact_id)

    def _get_table_name(self) -> str:
        return self._dag.must_get_artifact(self._artifact_id).name

    def __str__(self) -> str:
        """Prints out a human-readable description of the table artifact."""

        input_operator = self._dag.must_get_operator(with_output_artifact_id=self._artifact_id)
        load_operators = self._dag.list_operators(
            filter_to=[OperatorType.LOAD],
            on_artifact_id=self._artifact_id,
        )

        metric_operators = self._dag.list_operators(
            filter_to=[OperatorType.METRIC],
            on_artifact_id=self._artifact_id,
        )

        check_operators = get_checks_for_op(input_operator, self._dag)

        readable_dict = super()._describe()
        readable_dict.update(
            {
                "Metrics": [get_description_for_metric(op, self._dag) for op in metric_operators],
                "Checks": [get_description_for_check(op) for op in check_operators],
                "Destinations": [
                    {
                        "Name": op.name,
                        "Description": op.description,
                        "Spec": op.spec.json(exclude_none=True),
                    }
                    for op in load_operators
                    if op.spec.load is not None
                ],
            }
        )

        return f"""
{format_header_for_print(f"'{input_operator.name}' Table Artifact")}
{json.dumps(readable_dict, sort_keys=False, indent=4)}
        """

    def describe(self) -> None:
        """Prints the stringified description of the table artifact to stdout."""
        print(self.__str__())

    def remove_check(self, name: str) -> None:
        """Remove a check on this artifact by name.

        Raises:
            InvalidUserActionException: if a matching check operator could not be found.
        """
        apply_deltas_to_dag(
            self._dag,
            deltas=[
                RemoveCheckOperatorDelta(check_name=name, artifact_id=self._artifact_id),
            ],
        )

import inspect
import warnings
from functools import wraps
from typing import Any, Callable, List, Mapping, Optional, Union, cast

import numpy as np
from aqueduct.artifacts import utils as artifact_utils
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.bool_artifact import BoolArtifact
from aqueduct.artifacts.metadata import ArtifactMetadata
from aqueduct.artifacts.numeric_artifact import NumericArtifact
from aqueduct.artifacts.utils import to_artifact_class
from aqueduct.dag_deltas import AddOrReplaceOperatorDelta, apply_deltas_to_dag
from aqueduct.enums import (
    ArtifactType,
    CheckSeverity,
    ExecutionMode,
    FunctionGranularity,
    FunctionType,
    OperatorType,
)
from aqueduct.error import InvalidUserActionException, InvalidUserArgumentException
from aqueduct.operators import CheckSpec, FunctionSpec, MetricSpec, Operator, OperatorSpec
from aqueduct.parameter_utils import create_param
from aqueduct.utils import (
    CheckFunction,
    MetricFunction,
    Number,
    UserFunction,
    artifact_name_from_op_name,
    generate_uuid,
    serialize_function,
)

from aqueduct import dag as dag_module
from aqueduct import globals

OutputArtifactFunction = Callable[..., BaseArtifact]

# For functions that can handle multiple outputs (eg. `op()`)
OutputArtifactsFunction = Callable[..., Union[BaseArtifact, List[BaseArtifact]]]

# Type declarations for functions
DecoratedFunction = Callable[[UserFunction], OutputArtifactsFunction]

# Type declarations for metrics
DecoratedMetricFunction = Callable[[MetricFunction], OutputArtifactFunction]

# Type declarations for checks
DecoratedCheckFunction = Callable[[CheckFunction], OutputArtifactFunction]


def _is_input_artifact(elem: Any) -> bool:
    return isinstance(elem, BaseArtifact)


def wrap_spec(
    spec: OperatorSpec,
    *input_artifacts: BaseArtifact,
    op_name: str,
    output_artifact_type_hints: List[ArtifactType],
    description: str = "",
    execution_mode: ExecutionMode = ExecutionMode.EAGER,
) -> Union[BaseArtifact, List[BaseArtifact]]:
    """Applies a python function to existing artifacts.
    The function must be named predict() on a class named "Function",
    in a file named "model.py":

    >>> class Function:
    >>>     def predict(self, *args):
    >>>         ...

    Args:
        spec:
            The spec of the operator.
        *input_artifacts:
            All the artifacts that will serve as input to the python function.
            The function must have the same number of parameters as input
            artifacts.
        op_name:
            The name of the operator that generated this artifact.
        output_artifact_type_hints:
            The artifact types that the function is expected to output, in the correct order.
        description:
            The description for this operator.

    Returns:
        A list of artifacts, representing the outputs of the function.
    """
    assert (
        len(output_artifact_type_hints) > 0
    ), "Non-load operators must have at least one output artifact."

    for artifact in input_artifacts:
        if artifact._from_flow_run:
            raise InvalidUserActionException(
                "Artifact %s fetched from flow run cannot be reused for computation."
                % artifact.name
            )

    dag = dag_module.__GLOBAL_DAG__

    # Check that all the artifact ids exist in the dag.
    for artifact in input_artifacts:
        _ = dag.must_get_artifact(artifact.id())

    operator_id = generate_uuid()
    output_artifact_ids = [generate_uuid() for _ in output_artifact_type_hints]

    def _construct_artifact_name(i: int, op_name: str) -> str:
        """The artifact naming policy is "<op_name> artifact <optional counter>".

        The counter starts at 1, and is only present in the multi-output case.
        """
        assert i < len(output_artifact_ids)
        if len(output_artifact_ids) == 1:
            return artifact_name_from_op_name(op_name)
        return artifact_name_from_op_name(op_name) + " %d" % (i + 1)

    apply_deltas_to_dag(
        dag,
        deltas=[
            AddOrReplaceOperatorDelta(
                op=Operator(
                    id=operator_id,
                    name=op_name,
                    description=description,
                    spec=spec,
                    inputs=[artifact.id() for artifact in input_artifacts],
                    outputs=output_artifact_ids,
                ),
                output_artifacts=[
                    ArtifactMetadata(
                        id=output_artifact_id,
                        name=_construct_artifact_name(i, op_name),
                        type=output_artifact_type_hints[i],
                    )
                    for i, output_artifact_id in enumerate(output_artifact_ids)
                ],
            ),
        ],
    )

    if execution_mode == ExecutionMode.EAGER:
        # Issue preview request since this is an eager execution.
        output_artifacts = artifact_utils.preview_artifacts(dag, output_artifact_ids)
    else:
        # We are in lazy mode.
        output_artifacts = [
            to_artifact_class(dag, output_artifact_id, output_artifact_type_hints[i])
            for i, output_artifact_id in enumerate(output_artifact_ids)
        ]

    # Return a singular artifact if `num_outputs` == 1.
    return output_artifacts if len(output_artifacts) > 1 else output_artifacts[0]


def _type_check_decorator_arguments(
    description: Optional[str],
    file_dependencies: Optional[List[str]],
    requirements: Optional[Union[str, List[str]]],
) -> None:
    """
    Raises an InvalidUserArgumentException if any issues are found.
    """
    if description is not None and not isinstance(description, str):
        raise InvalidUserArgumentException("A supplied description must be of string type.")

    if file_dependencies is not None:
        if not isinstance(file_dependencies, list):
            raise InvalidUserArgumentException("File dependencies must be specified as a list.")
        if any(not isinstance(file_dep, str) for file_dep in file_dependencies):
            raise InvalidUserArgumentException("Each file dependency must be a string.")

    if requirements is not None:
        is_list = isinstance(requirements, list)
        if not isinstance(requirements, str) and not is_list:
            raise InvalidUserArgumentException(
                "Requirements must either be a path string or a list of pip requirements specifiers."
            )
        if is_list and any(not isinstance(req, str) for req in requirements):
            raise InvalidUserArgumentException("Each pip requirements specifier must be a string.")


def _type_check_decorated_function_arguments(
    operator_type: OperatorType, *input_artifacts: BaseArtifact
) -> None:
    for artifact in input_artifacts:
        if not isinstance(artifact, BaseArtifact):
            raise InvalidUserArgumentException(
                "Input to decorated must be an Aqueduct artifact, got type %s." % type(artifact)
            )

        if operator_type == OperatorType.FUNCTION:
            if (
                artifact._from_operator_type == OperatorType.METRIC
                or artifact._from_operator_type == OperatorType.CHECK
            ):
                raise InvalidUserActionException(
                    "Artifact from metric or check operator cannot be used as input to %s operator."
                    % operator_type
                )
        if operator_type == OperatorType.METRIC or operator_type == OperatorType.CHECK:
            if artifact._from_operator_type == OperatorType.CHECK:
                raise InvalidUserActionException(
                    "Artifact from check operator cannot be used as input to %s operator."
                    % operator_type
                )


def _convert_input_arguments_to_parameters(
    *input_artifacts: Any, func_params: Mapping[str, inspect.Parameter]
) -> List[BaseArtifact]:
    """
    Converts non-artifact inputs to parameters.

    Errors if the implicitly-created parameter collides with an existing one. Also errors if the function has a
    variable-length parameter, since we don't know what name to attribute to those.

    `func_params` maps from parameter name to an `inspect.Parameter` object containing additional information.
    """
    # KEYWORD_ONLY parameters are allowed, since they are guaranteed to have a name.
    # Note that we only accept them after "*" arguments, since we error out on VAR_POSITIONAL (eg. *args).
    # For example, `foo(*, positional_arg)` is allowed, but `foo(*args, positional_arg)` is not.
    # See https://peps.python.org/pep-0362/#parameter-object for a description of each parameter kind.
    disallowed_kinds = [inspect.Parameter.VAR_POSITIONAL, inspect.Parameter.VAR_KEYWORD]
    implicit_params_disallowed = any(
        param.kind in disallowed_kinds for param in func_params.values()
    )

    dag = dag_module.__GLOBAL_DAG__
    param_names = list(func_params.keys())

    artifacts = list(input_artifacts)
    for idx, artifact in enumerate(artifacts):
        if not isinstance(artifact, BaseArtifact):
            if implicit_params_disallowed:
                raise InvalidUserArgumentException(
                    """Input at index %d to function is not an artifact. Creating an Aqueduct parameter implicitly for a """
                    """function that takes in variable-length parameters (eg. *args or **kwargs) is currently unsupported."""
                    % idx
                )

            # We assume that the parameter name exists here, since we've disallowed any variable-length parameters.
            arg_name = param_names[idx]
            if dag.get_operator(with_name=arg_name) is not None:
                raise InvalidUserArgumentException(
                    """Input to function argument "%s" is not an artifact. We tried implicitly \
creating a parameter named "%s", but an existing operator or parameter with the same name already exists."""
                    % (arg_name, arg_name)
                )

            new_artifact = create_param(dag=dag, name=arg_name, default=artifact)
            warnings.warn(
                """Input to function argument "%s" is not an artifact type. We have implicitly \
created a parameter named "%s" and your input will be used as its default value. This parameter \
will be used when running the function."""
                % (arg_name, arg_name)
            )
            artifacts[idx] = new_artifact
    return artifacts


def op(
    name: Optional[Union[str, UserFunction]] = None,
    description: Optional[str] = None,
    file_dependencies: Optional[List[str]] = None,
    requirements: Optional[Union[str, List[str]]] = None,
    num_outputs: int = 1,
) -> Union[DecoratedFunction, OutputArtifactsFunction]:
    """Decorator that converts regular python functions into an operator.

    Calling the decorated function returns an Artifact. The decorated function
    can take any number of artifact inputs.

    To run the wrapped code locally, without Aqueduct, use the `local` attribute. Eg:
    >>> compute_recommendations.local(customer_profiles, recent_clicks)

    Args:
        name:
            Operator name. Defaults to the function name if not provided (or is of a non-string type).
        description:
            A description for the operator.
        file_dependencies:
            A list of relative paths to files that the function needs to access.
            Python classes/methods already imported within the function's file
            need not be included.
        requirements:
            Defines the python package requirements that this operator will run with.
            Can be either a path to the requirements.txt file or a list of pip requirements specifiers.
            (eg. ["transformers==4.21.0", "numpy==1.22.4"]. If not supplied, we'll first
            look for a `requirements.txt` file in the same directory as the decorated function
            and install those. Otherwise, we'll attempt to infer the requirements with
            `pip freeze`.
        num_outputs:
            The number of outputs the decorated function is expected to return.
            Will fail at runtime if a different number of outputs is returned by the function.

    Examples:
        The op name is inferred from the function name. The description is pulled from the function
        docstring or can be explicitly set in the decorator.

        >>> @op
        ... def compute_recommendations(customer_profiles, recent_clicks):
        ...     return recommendations
        >>> customer_profiles = db.sql("SELECT * from user_profile", db="postgres")
        >>> recent_clicks = db.sql("SELECT * recent_clicks", db="google_analytics/shopping")
        >>> recommendations = compute_recommendations(customer_profiles, recent_clicks)

        `recommendations` is an Artifact representing the result of `compute_recommendations()`.

        >>> recommendations.get()
    """
    _type_check_decorator_arguments(description, file_dependencies, requirements)
    if num_outputs < 1:
        raise InvalidUserArgumentException("`num_outputs` must be set to a positive integer.")

    def inner_decorator(func: UserFunction) -> OutputArtifactsFunction:
        nonlocal name
        nonlocal description
        if name is None or not isinstance(name, str):
            name = func.__name__
        if description is None:
            description = func.__doc__ or ""

        def _wrapped_util(
            *input_artifacts: BaseArtifact, execution_mode: ExecutionMode
        ) -> Union[BaseArtifact, List[BaseArtifact]]:
            """
            Creates the following files in the zipped folder structure:
             - model.py
             - model.pkl
             - requirements.txt
             - python_version.txt
             - <any file dependencies>
            """
            assert isinstance(name, str)
            assert isinstance(description, str)

            artifacts = _convert_input_arguments_to_parameters(
                *input_artifacts,
                func_params=inspect.signature(func).parameters,
            )

            _type_check_decorated_function_arguments(OperatorType.FUNCTION, *artifacts)

            zip_file = serialize_function(func, name, file_dependencies, requirements)
            function_spec = FunctionSpec(
                type=FunctionType.FILE,
                granularity=FunctionGranularity.TABLE,
                file=zip_file,
            )
            return wrap_spec(
                OperatorSpec(function=function_spec),
                *artifacts,
                op_name=name,
                output_artifact_type_hints=[ArtifactType.UNTYPED for _ in range(num_outputs)],
                description=description,
                execution_mode=execution_mode,
            )

        def wrapped(*input_artifacts: BaseArtifact) -> Union[BaseArtifact, List[BaseArtifact]]:
            return _wrapped_util(*input_artifacts, execution_mode=ExecutionMode.EAGER)

        # Enable the .local(*args) attribute, which calls the original function with the raw inputs.
        def local_func(*inputs: Any) -> Any:
            raw_inputs = [elem.get() if _is_input_artifact(elem) else elem for elem in inputs]
            return func(*raw_inputs)

        setattr(wrapped, "local", local_func)

        def lazy_mode(*input_artifacts: BaseArtifact) -> Union[BaseArtifact, List[BaseArtifact]]:
            return _wrapped_util(*input_artifacts, execution_mode=ExecutionMode.LAZY)

        setattr(wrapped, "lazy", lazy_mode)

        if globals.__GLOBAL_CONFIG__.lazy:
            return lazy_mode

        return wrapped

    if callable(name):
        # This only happens when the decorator is used without parenthesis, eg: @op.
        return inner_decorator(name)
    else:
        return inner_decorator


def metric(
    name: Optional[Union[str, MetricFunction]] = None,
    description: Optional[str] = None,
    file_dependencies: Optional[List[str]] = None,
    requirements: Optional[Union[str, List[str]]] = None,
) -> Union[DecoratedMetricFunction, OutputArtifactFunction]:
    """Decorator that converts regular python functions into a metric.

    Calling the decorated function returns a NumericArtifact. The decorated function
    can take any number of artifact inputs.

    The requirements.txt file in the current directory is used, if it exists.

    To run the wrapped code locally, without Aqueduct, use the `local` attribute. Eg:
    >>> compute_recommendations.local(customer_profiles, recent_clicks)

    Args:
        name:
            Operator name. Defaults to the function name if not provided (or is of a non-string type).
        description:
            A description for the metric.
        file_dependencies:
            A list of relative paths to files that the function needs to access.
            Python classes/methods already imported within the function's file
            need not be included.
        requirements:
            Defines the python package requirements that this operator will run with.
            Can be either a path to the requirements.txt file or a list of pip requirements specifiers.
            (eg. ["transformers==4.21.0", "numpy==1.22.4"]. If not supplied, we'll first
            look for a `requirements.txt` file in the same directory as the decorated function
            and install those. Otherwise, we'll attempt to infer the requirements with
            `pip freeze`.

    Examples:
        The metric name is inferred from the function name. The description is pulled from the function
        docstring or can be explicitly set in the decorator.

        >>> @metric()
        ... def avg_churn(churn_table):
        ...     return churn_table['pred_churn'].mean()
        >>> churn_table = db.sql("SELECT * from churn_table")
        >>> churn_metric = avg_churn(churn_table)

        `churn_metric` is a NumericArtifact representing the result of `avg_churn()`.

        >>> churn_metric.get()
    """
    _type_check_decorator_arguments(description, file_dependencies, requirements)

    def inner_decorator(func: MetricFunction) -> OutputArtifactFunction:
        nonlocal name
        nonlocal description
        if name is None or not isinstance(name, str):
            name = func.__name__
        if description is None:
            description = func.__doc__ or ""

        def _wrapped_util(
            *input_artifacts: BaseArtifact, execution_mode: ExecutionMode
        ) -> NumericArtifact:
            """
            Creates the following files in the zipped folder structure:
             - model.py
             - model.pkl
             - requirements.txt
             - python_version.txt
             - <any file dependencies>
            """
            assert isinstance(name, str)
            assert isinstance(description, str)

            artifacts = _convert_input_arguments_to_parameters(
                *input_artifacts,
                func_params=inspect.signature(func).parameters,
            )

            _type_check_decorated_function_arguments(OperatorType.METRIC, *artifacts)

            zip_file = serialize_function(func, name, file_dependencies, requirements)

            # TODO(ENG-735): Support granularity=FunctionGranularity.TABLE & granularity=FunctionGranularity.ROW
            function_spec = FunctionSpec(
                type=FunctionType.FILE,  # TODO(ENG-811): Support type=FunctionType.GITHUB
                granularity=FunctionGranularity.TABLE,
                file=zip_file,
            )

            metric_spec = MetricSpec(function=function_spec)

            numeric_artifact = wrap_spec(
                OperatorSpec(metric=metric_spec),
                *artifacts,
                op_name=name,
                output_artifact_type_hints=[ArtifactType.NUMERIC],
                description=description,
                execution_mode=execution_mode,
            )
            assert isinstance(numeric_artifact, NumericArtifact)

            numeric_artifact.set_operator_type(OperatorType.METRIC)

            return numeric_artifact

        """
        @wraps(...): updates `wrapped()` to look like `func()` by copying attributes (eg. __name__, __doc__, etc.)
        """

        @wraps(func)
        def wrapped(
            *input_artifacts: BaseArtifact,
        ) -> NumericArtifact:
            return _wrapped_util(*input_artifacts, execution_mode=ExecutionMode.EAGER)

        # Enable the .local(*args) attribute, which calls the original function with the raw inputs.
        def local_func(*inputs: Any) -> Number:
            raw_inputs = [elem.get() if _is_input_artifact(elem) else elem for elem in inputs]
            return func(*raw_inputs)

        setattr(wrapped, "local", local_func)

        def lazy_mode(*input_artifacts: BaseArtifact) -> NumericArtifact:
            return _wrapped_util(*input_artifacts, execution_mode=ExecutionMode.LAZY)

        setattr(wrapped, "lazy", lazy_mode)

        if globals.__GLOBAL_CONFIG__.lazy:
            return lazy_mode

        return wrapped

    if callable(name):
        # This only happens when the decorator is used without parenthesis, eg: @metric.
        return inner_decorator(name)
    else:
        return inner_decorator


def check(
    name: Optional[Union[str, CheckFunction]] = None,
    description: Optional[str] = None,
    severity: CheckSeverity = CheckSeverity.WARNING,
    file_dependencies: Optional[List[str]] = None,
    requirements: Optional[Union[str, List[str]]] = None,
) -> Union[DecoratedCheckFunction, OutputArtifactFunction]:
    """Decorator that converts a regular python function into a check.

    Calling the decorated function returns a BoolArtifact. The decorated python function
    can have any number of artifact inputs.

    A check can be set with either WARNING or ERROR severity. A failing check with ERROR severity
    will fail the workflow when run in our system.

    To run the wrapped code locally, without Aqueduct, use the `local` attribute. Eg:
    >>> compute_recommendations.local(customer_profiles, recent_clicks)

    Args:
        name:
            Operator name. Defaults to the function name if not provided (or is of a non-string type).
        description:
            A description for the check.
        severity:
            The severity level of the check if it fails.
        file_dependencies:
            A list of relative paths to files that the function needs to access.
            Python classes/methods already imported within the function's file
            need not be included.
        requirements:
            Defines the python package requirements that this operator will run with.
            Can be either a path to the requirements.txt file or a list of pip requirements specifiers.
            (eg. ["transformers==4.21.0", "numpy==1.22.4"]. If not supplied, we'll first
            look for a `requirements.txt` file in the same directory as the decorated function
            and install those. Otherwise, we'll attempt to infer the requirements with
            `pip freeze`.

    Examples:
        The check name is inferred from the function name. The description is pulled from the function
        docstring or can be explicitly set in the decorator.

        >>> @check(
        ... description="The average predicted churn is below 0.1",
        ... severity=CheckSeverity.ERROR,
        ... )
        ... def avg_churn_is_low(churn_table):
        ...     return churn_table['pred_churn'].mean() < 0.1
        >>> churn_is_low_check = avg_churn_is_low(churn_table_artifact)

        `churn_is_low_check` is a BoolArtifact representing the result of `avg_churn_is_low()`.

        >>> churn_is_low_check.get()
    """
    _type_check_decorator_arguments(description, file_dependencies, requirements)

    def inner_decorator(func: CheckFunction) -> OutputArtifactFunction:
        nonlocal name
        nonlocal description
        if name is None or not isinstance(name, str):
            name = func.__name__
        if description is None:
            description = func.__doc__ or ""

        def _wrapped_util(
            *input_artifacts: BaseArtifact, execution_mode: ExecutionMode
        ) -> BoolArtifact:
            """
            Creates the following files in the zipped folder structure:
             - model.py
             - model.pkl
             - requirements.txt
             - python_version.txt
             - <any file dependencies>
            """
            assert isinstance(name, str)
            assert isinstance(description, str)

            artifacts = _convert_input_arguments_to_parameters(
                *input_artifacts,
                func_params=inspect.signature(func).parameters,
            )

            _type_check_decorated_function_arguments(OperatorType.CHECK, *artifacts)

            zip_file = serialize_function(func, name, file_dependencies, requirements)
            function_spec = FunctionSpec(
                type=FunctionType.FILE,
                granularity=FunctionGranularity.TABLE,
                file=zip_file,
            )
            check_spec = CheckSpec(level=severity, function=function_spec)

            bool_artifact = wrap_spec(
                OperatorSpec(check=check_spec),
                *artifacts,
                op_name=name,
                output_artifact_type_hints=[ArtifactType.BOOL],
                description=description,
                execution_mode=execution_mode,
            )

            assert isinstance(bool_artifact, BoolArtifact)

            bool_artifact.set_operator_type(OperatorType.CHECK)

            return bool_artifact

        """
        @wraps(...): updates `wrapped()` to look like `func()` by copying attributes (eg. __name__, __doc__, etc.)
        """

        @wraps(func)
        def wrapped(
            *input_artifacts: BaseArtifact,
        ) -> BoolArtifact:
            return _wrapped_util(*input_artifacts, execution_mode=ExecutionMode.EAGER)

        # Enable the .local(*args) attribute, which calls the original function with the raw inputs.
        def local_func(*inputs: Any) -> Union[bool, np.bool_]:
            raw_inputs = [elem.get() if _is_input_artifact(elem) else elem for elem in inputs]
            return func(*raw_inputs)

        setattr(wrapped, "local", local_func)

        def lazy_mode(*input_artifacts: BaseArtifact) -> BoolArtifact:
            return _wrapped_util(*input_artifacts, execution_mode=ExecutionMode.LAZY)

        setattr(wrapped, "lazy", lazy_mode)

        if globals.__GLOBAL_CONFIG__.lazy:
            return lazy_mode
        return wrapped

    if callable(name):
        # This only happens when the decorator is used without parenthesis, eg: @check.
        return inner_decorator(name)
    else:
        return inner_decorator


def to_operator(
    func: UserFunction,
    name: Optional[str] = None,
    description: Optional[str] = None,
    file_dependencies: Optional[List[str]] = None,
    requirements: Optional[Union[str, List[str]]] = None,
) -> OutputArtifactsFunction:
    """Convert a function that returns a dataframe into an Aqueduct operator.

    Args:
        func:
            the python function that is to be converted into operator.
        name:
            Operator name.
        description:
            A description for the operator.
        file_dependencies:
            A list of relative paths to files that the function needs to access.
            Python classes/methods already imported within the function's file
            need not be included.
        requirements:
            Defines the python package requirements that this operator will run with.
            Can be either a path to the requirements.txt file or a list of pip requirements specifiers.
            (eg. ["transformers==4.21.0", "numpy==1.22.4"]. If not supplied, we'll first
            look for a `requirements.txt` file in the same directory as the decorated function
            and install those. Otherwise, we'll attempt to infer the requirements with
            `pip freeze`.
    Returns:
        An Aqueduct operator that can be used just like any decorated operator.
    """
    if callable(name):
        # This only happens when the decorator is used without parenthesis, eg: @op.
        # We use `op()` like a normal function, so we can rule out this case.
        raise InvalidUserArgumentException("Supplied name must be a string.")

    # Since `name` must be a string, we know that only one of the return values of `op()`
    # is possible.
    func_op = cast(
        DecoratedFunction,
        op(
            name=name,
            description=description,
            file_dependencies=file_dependencies,
            requirements=requirements,
        ),
    )
    return func_op(func)

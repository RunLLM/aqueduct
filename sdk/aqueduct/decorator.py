import inspect
import uuid
import warnings
from functools import wraps
from typing import Any, Callable, Dict, List, Mapping, Optional, Union, cast

import numpy as np
from aqueduct.artifacts._create import create_metric_or_check_artifact
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.bool_artifact import BoolArtifact
from aqueduct.artifacts.create import create_param_artifact, to_artifact_class
from aqueduct.artifacts.numeric_artifact import NumericArtifact
from aqueduct.artifacts.preview import preview_artifacts
from aqueduct.constants.enums import (
    ArtifactType,
    CheckSeverity,
    CustomizableResourceType,
    ExecutionMode,
    FunctionGranularity,
    FunctionType,
    OperatorType,
)
from aqueduct.error import InvalidUserActionException, InvalidUserArgumentException
from aqueduct.logger import logger
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.operators import (
    CheckSpec,
    FunctionSpec,
    MetricSpec,
    Operator,
    OperatorSpec,
    ResourceConfig,
    get_operator_type,
)
from aqueduct.type_annotations import CheckFunction, MetricFunction, Number, UserFunction
from aqueduct.utils.dag_deltas import (
    AddOperatorDelta,
    DAGDelta,
    RemoveOperatorDelta,
    apply_deltas_to_dag,
)
from aqueduct.utils.function_packaging import serialize_function
from aqueduct.utils.naming import default_artifact_name_from_op_name, sanitize_artifact_name
from aqueduct.utils.utils import generate_engine_config, generate_uuid

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
    output_artifact_names: Optional[List[str]] = None,
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
        output_artifact_names:
            If set, provides the custom output artifact names.
        output_artifact_type_hints:
            The artifact types that the function is expected to output, in the correct order.
        description:
            The description for this operator.
        execution_mode:
            Whether the operator should be executed eagerly or lazily.
        overwrite_op:
            The operator to overwrite, if any. Should only be set for metrics and checks.

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

    dag = globals.__GLOBAL_DAG__

    # Check that all the artifact ids exist in the dag.
    for artifact in input_artifacts:
        _ = dag.must_get_artifact(artifact.id())

    operator_id = generate_uuid()
    output_artifact_ids = [generate_uuid() for _ in output_artifact_type_hints]

    # Even if there are multiple outputs, we give them all the same artifact names (in the default case).
    # These will be deduplicated at publish time.
    artifact_names = output_artifact_names or [
        default_artifact_name_from_op_name(op_name) for _ in range(len(output_artifact_ids))
    ]

    new_op = Operator(
        id=operator_id,
        name=op_name,
        description=description,
        spec=spec,
        inputs=[artifact.id() for artifact in input_artifacts],
        outputs=output_artifact_ids,
    )

    new_output_artifacts = [
        ArtifactMetadata(
            id=output_artifact_id,
            name=sanitize_artifact_name(artifact_names[i]),
            type=output_artifact_type_hints[i],
            explicitly_named=output_artifact_names is not None,
        )
        for i, output_artifact_id in enumerate(output_artifact_ids)
    ]

    # Update the dag to reflect the newly created operator.
    if get_operator_type(new_op) in [OperatorType.METRIC, OperatorType.CHECK]:
        create_metric_or_check_artifact(dag, new_op, new_output_artifacts)
    else:
        apply_deltas_to_dag(
            dag,
            deltas=[
                AddOperatorDelta(
                    op=new_op,
                    output_artifacts=new_output_artifacts,
                )
            ],
        )

    if execution_mode == ExecutionMode.EAGER:
        # Issue preview request since this is an eager execution.
        output_artifacts = preview_artifacts(dag, output_artifact_ids)
    else:
        # We are in lazy mode.
        output_artifacts = [
            to_artifact_class(dag, output_artifact_id, output_artifact_type_hints[i])
            for i, output_artifact_id in enumerate(output_artifact_ids)
        ]

    # Return a singular artifact if `num_outputs` == 1.
    return output_artifacts if len(output_artifacts) > 1 else output_artifacts[0]


def _typecheck_op_decorator_arguments(
    description: Optional[str],
    file_dependencies: Optional[List[str]],
    requirements: Optional[Union[str, List[str]]],
    engine: Optional[str],
    num_outputs: int,
    outputs: Optional[List[str]],
) -> None:
    _typecheck_common_decorator_arguments(description, file_dependencies, requirements)

    if engine is not None and not isinstance(engine, str):
        raise InvalidUserArgumentException("`engine` must be a string.")

    if num_outputs is not None:
        if not isinstance(num_outputs, int) or num_outputs < 1:
            raise InvalidUserArgumentException("`num_outputs` must be set to a positive integer.")

    if outputs is not None:
        if not (isinstance(outputs, str) or isinstance(outputs, List)):
            raise InvalidUserArgumentException("`outputs` must be either a string or a list.")


def _typecheck_common_decorator_arguments(
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
    *input_artifacts: Any, op_name: str, func_params: Mapping[str, inspect.Parameter]
) -> List[BaseArtifact]:
    """
    Converts non-artifact inputs to parameters.

    Errors if the function has a variable-length parameter, since we don't know what name to attribute to those.

    For parameters created this way, the naming collision policy is as follows: we will error
    if there exists other operators or artifacts with the same name, unless we are overwriting
    another implicit parameter being used by the same operator. An implicit parameter is named
    "<op_name>:<param_name>".

    Args:
        input_artifacts:
            Entries in this list are not artifacts if the corresponding argument was supplied as an
            implicit parameter.
        op_name:
            The name of the operator that will consume this implicit parameter as input. Necessary only
            for resolving implicit parameter naming collisions.
        func_params:
            Maps from parameter name to an `inspect.Parameter` object containing additional information.
    """
    # KEYWORD_ONLY parameters are allowed, since they are guaranteed to have a name.
    # Note that we only accept them after "*" arguments, since we error out on VAR_POSITIONAL (eg. *args).
    # For example, `foo(*, positional_arg)` is allowed, but `foo(*args, positional_arg)` is not.
    # See https://peps.python.org/pep-0362/#parameter-object for a description of each parameter kind.
    disallowed_kinds = [inspect.Parameter.VAR_POSITIONAL, inspect.Parameter.VAR_KEYWORD]
    implicit_params_disallowed = any(
        param.kind in disallowed_kinds for param in func_params.values()
    )

    dag = globals.__GLOBAL_DAG__
    fn_param_names = list(func_params.keys())

    artifacts = list(input_artifacts)
    for idx, artifact in enumerate(artifacts):
        if not isinstance(artifact, BaseArtifact):
            if implicit_params_disallowed:
                raise InvalidUserArgumentException(
                    """Input at index %d to function is not an artifact. Creating an Aqueduct parameter implicitly for a """
                    """function that takes in variable-length parameters (eg. *args or **kwargs) is currently unsupported."""
                    % idx
                )

            # We assume that the user-function's parameter name exists here, since we've disallowed any variable-length parameters.
            # An implicit parameter will have the operator name prepended to it.
            param_name = op_name + ":" + fn_param_names[idx]

            # Implicit parameters are only ever created (or error). They never replace anything.
            logger().warning(
                """Operator `%s`'s argument `%s` is not an artifact type. We have implicitly created a parameter named `%s` and your input will be used as its default value. This parameter will be used when running the function."""
                % (op_name, fn_param_names[idx], param_name)
            )
            artifacts[idx] = create_param_artifact(
                dag=dag,
                param_name=param_name,
                default=artifact,
                description="Parameter corresponding to argument `%s` of function `%s`."
                % (fn_param_names[idx], op_name),
                explicitly_named=False,
            )

    # If the user has supplied fewer arguments than the function takes, we check if the remaining
    # arguments have default values. If they do, we create implicit parameters for them.
    if len(artifacts) < len(fn_param_names):
        for idx in range(len(artifacts), len(fn_param_names)):
            default_value = func_params[fn_param_names[idx]].default
            if default_value is inspect.Parameter.empty:
                raise InvalidUserArgumentException(
                    """No input was provided for argument `%s` of function `%s`, and no default value was specified."""
                    % (fn_param_names[idx], op_name)
                )

            param_name = op_name + ":" + fn_param_names[idx]

            # Implicit parameters are only ever created (or error). They never replace anything.
            logger().warning(
                """Operator `%s`'s argument `%s` is not an artifact type. We have implicitly created a parameter named `%s` and your input will be used as its default value. This parameter will be used when running the function."""
                % (op_name, fn_param_names[idx], param_name)
            )
            artifacts.append(
                create_param_artifact(
                    dag=dag,
                    param_name=param_name,
                    default=default_value,
                    description="Parameter corresponding to argument `%s` of function `%s`."
                    % (fn_param_names[idx], op_name),
                    explicitly_named=False,
                )
            )

    return artifacts


def _convert_memory_string_to_mbs(memory_str: str) -> int:
    """Converts a memory string supplied by the user into the equivalent number in MBs.

    Only "MB" and "GB" suffixes are supported, case-insensitive.
    """
    memory_str = memory_str.strip()
    if len(memory_str) <= 2:
        raise InvalidUserArgumentException(
            "Memory value `%s` not long enough, it must be a number and a two character suffix (eg. 100MB)."
            % memory_str
        )

    if memory_str[-2:].upper() == "MB":
        multiplier = 1
    elif memory_str[-2:].upper() == "GB":
        multiplier = 1000
    else:
        raise InvalidUserArgumentException(
            "Memory value `%s` is invalid. It must have a suffix that is one of mb/MB/gb/GB."
            % memory_str,
        )

    memory_scalar_str = memory_str[:-2].strip()
    if not memory_scalar_str.isnumeric():
        raise InvalidUserArgumentException(
            "Memory value `%s` has an invalid value. `%s` must be a positive integer."
            % (memory_str, memory_scalar_str),
        )

    return multiplier * int(memory_scalar_str)


def _update_operator_spec_with_engine(
    spec: OperatorSpec,
    engine: Optional[str] = None,
) -> None:
    if engine is not None:
        if globals.__GLOBAL_API_CLIENT__ is None:
            raise InvalidUserActionException(
                "Aqueduct Client was not instantiated! Please create a client and retry."
            )

        spec.engine_config = generate_engine_config(
            globals.__GLOBAL_API_CLIENT__.list_integrations(),
            engine,
        )


def _update_operator_spec_with_resources(
    spec: OperatorSpec,
    resources: Optional[Dict[str, Any]] = None,
) -> None:
    if resources is not None:
        if not isinstance(resources, Dict) or any(not isinstance(k, str) for k in resources):
            raise InvalidUserArgumentException("`resources` must be a dictionary with string keys.")

        num_cpus = resources.get(CustomizableResourceType.NUM_CPUS)
        memory = resources.get(CustomizableResourceType.MEMORY)
        gpu_resource_name = resources.get(CustomizableResourceType.GPU_RESOURCE_NAME)
        cuda_version = resources.get(CustomizableResourceType.CUDA_VERSION)

        if num_cpus is not None and (not isinstance(num_cpus, int) or num_cpus < 0):
            raise InvalidUserArgumentException(
                "`num_cpus` value must be set to a positive integer."
            )

        # `memory` value can be either an int (in MBs) or a string. We will convert it into an integer
        # representing the number of MBs.
        if memory is not None:
            if not isinstance(memory, int) and not isinstance(memory, str):
                raise InvalidUserArgumentException(
                    "`memory` value must be either an integer or string."
                )

            if isinstance(memory, int) and memory < 0:
                raise InvalidUserArgumentException(
                    "If `memory` value is set as an integer, it must be positive."
                )

            # We'll need to convert the string value into an integer (in MBs).
            if isinstance(memory, str):
                memory = _convert_memory_string_to_mbs(memory)

            assert isinstance(memory, int)

        if gpu_resource_name is not None and (not isinstance(gpu_resource_name, str)):
            raise InvalidUserArgumentException("`gpu_resource_name` value must be set to a string.")

        if cuda_version is not None and (not isinstance(cuda_version, str)):
            raise InvalidUserArgumentException("`cuda_version` value must be set to a string.")

        if cuda_version is not None and gpu_resource_name is None:
            raise InvalidUserArgumentException(
                "`cuda_version` can only be set if a `gpu_resource_name` is specified."
            )

        spec.resources = ResourceConfig(
            num_cpus=num_cpus,
            memory_mb=memory,
            gpu_resource_name=gpu_resource_name,
            cuda_version=cuda_version,
        )


def op(
    name: Optional[Union[str, UserFunction]] = None,
    description: Optional[str] = None,
    engine: Optional[str] = None,
    file_dependencies: Optional[List[str]] = None,
    requirements: Optional[Union[str, List[str]]] = None,
    num_outputs: Optional[int] = None,
    outputs: Optional[List[str]] = None,
    resources: Optional[Dict[str, Any]] = None,
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
        engine:
            The name of the compute integration this operator will run on. Defaults to the Aqueduct engine.
        file_dependencies:
            A list of relative paths to files that the function needs to access.
            Python classes/methods already imported within the function's file
            need not be included.
        requirements:
            Defines the python package requirements that this operator will run with.
            Can be either a path to the requirements.txt file or a list of pip requirements specifiers.
            (eg. ["transformers==4.21.0", "numpy==1.22.4"]. If not supplied, we'll first
            look for a `requirements.txt` file in the same directory as the decorated function
            and install those. Otherwise, the method raises RequirementsMissingError exception.
        num_outputs:
            The number of outputs the decorated function is expected to return.
            Will fail at runtime if a different number of outputs is returned by the function.
        outputs:
            The name to assign the output artifacts of for this operator. The number of names provided
            must match the number of return values of the decorated function. If not set, the artifact
            names will default to "<op_name> artifact <optional counter>".
        resources:
            A dictionary containing the custom resource configurations that this operator will run with.
            These configurations are guaranteed to be followed, we will not silently ignore any of them.
            If a resource configuration is unsupported by a particular execution engine, we will fail at
            execution time. The supported keys are:

            "num_cpus" (int):
                The number of cpus that this operator will run with. This operator will execute with *exactly*
                this number of cpus. If not enough cpus are available, operator execution will fail.
            "memory" (int, str):
                The amount of memory this operator will run with. This operator will execute with *exactly*
                this amount of memory. If not enough memory is available, operator execution will fail.

                If an integer value is supplied, the memory unit is assumed to be MB. If a string is supplied,
                a suffix indicating the memory unit must be supplied. Supported memory units are "MB" and "GB",
                case-insensitive.

                For example, the following values are valid: 100, "100MB", "1GB", "100mb", "1gb".
            "gpu_resource_name" (str):
                Name of the gpu resource to use (only applicable for Kubernetes engine).

                For example, the following value is valid: "nvidia.com/gpu".
            "cuda_version" (str):
                Version of CUDA to use with GPU (only applicable for Kubernetes engine). The currently supported
                values are "11.4.1" and "11.8.0".

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
    # Establish parity between `num_outputs` and `outputs`, or raise exception if there is a mismatch.
    if num_outputs is None and outputs is None:
        num_outputs = 1
    elif num_outputs is not None and outputs is not None:
        if len(outputs) != num_outputs:
            raise InvalidUserArgumentException(
                "`len(outputs) must be equivalent to `num_outputs`. Getting %d and %d"
                % (len(outputs), num_outputs)
            )
    elif num_outputs is None and outputs is not None:
        num_outputs = len(outputs)

    # If not set, default number of outputs is one.
    if num_outputs is None:
        num_outputs = 1

    _typecheck_op_decorator_arguments(
        description, file_dependencies, requirements, engine, num_outputs, outputs
    )

    def inner_decorator(func: UserFunction) -> OutputArtifactsFunction:
        nonlocal name
        nonlocal description
        if name is None or not isinstance(name, str):
            name = func.__name__
        if description is None:
            description = func.__doc__ or ""

        def _wrapped_util(
            *input_artifacts: BaseArtifact, execution_mode: Optional[ExecutionMode] = None
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

            if execution_mode == None:
                execution_mode = _get_global_execution_mode()

            assert isinstance(execution_mode, ExecutionMode)

            artifacts = _convert_input_arguments_to_parameters(
                *input_artifacts,
                op_name=name,
                func_params=inspect.signature(func).parameters,
            )

            _type_check_decorated_function_arguments(OperatorType.FUNCTION, *artifacts)

            zip_file = serialize_function(func, name, file_dependencies, requirements)
            function_spec = FunctionSpec(
                type=FunctionType.FILE,
                granularity=FunctionGranularity.TABLE,
                file=zip_file,
            )

            op_spec = OperatorSpec(
                function=function_spec,
            )

            _update_operator_spec_with_engine(op_spec, engine)
            _update_operator_spec_with_resources(op_spec, resources)

            assert isinstance(num_outputs, int)
            return wrap_spec(
                op_spec,
                *artifacts,
                op_name=name,
                output_artifact_names=outputs,
                output_artifact_type_hints=[ArtifactType.UNTYPED for _ in range(num_outputs)],
                description=description,
                execution_mode=execution_mode,
            )

        def wrapped(*input_artifacts: BaseArtifact) -> Union[BaseArtifact, List[BaseArtifact]]:
            return _wrapped_util(*input_artifacts)

        # Enable the .local(*args) attribute, which calls the original function with the raw inputs.
        def local_func(*inputs: Any) -> Any:
            raw_inputs = [elem.get() if _is_input_artifact(elem) else elem for elem in inputs]
            return func(*raw_inputs)

        setattr(wrapped, "local", local_func)

        def lazy_mode(*input_artifacts: BaseArtifact) -> Union[BaseArtifact, List[BaseArtifact]]:
            return _wrapped_util(*input_artifacts, execution_mode=ExecutionMode.LAZY)

        setattr(wrapped, "lazy", lazy_mode)

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
    output: Optional[str] = None,
    engine: Optional[str] = None,
    resources: Optional[Dict[str, Any]] = None,
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
            and install those. Otherwise, the method raises RequirementsMissingError exception.
        output:
            An optional custom name for the output metric artifact. Otherwise, the default naming scheme
            will be used.
        engine:
            The name of the compute integration this operator will run on.
        resources:
            A dictionary containing the custom resource configurations that this operator will run with.
            These configurations are guaranteed to be followed, we will not silently ignore any of them.
            If a resource configuration is unsupported by a particular execution engine, we will fail at
            execution time. The supported keys are:

            "num_cpus" (int):
                The number of cpus that this operator will run with. This operator will execute with *exactly*
                this number of cpus. If not enough cpus are available, operator execution will fail.
            "memory" (int, str):
                The amount of memory this operator will run with. This operator will execute with *exactly*
                this amount of memory. If not enough memory is available, operator execution will fail.

                If an integer value is supplied, the memory unit is assumed to be MB. If a string is supplied,
                a suffix indicating the memory unit must be supplied. Supported memory units are "MB" and "GB",
                case-insensitive.

                For example, the following values are valid: 100, "100MB", "1GB", "100mb", "1gb".
            "gpu_resource_name" (str):
                Name of the gpu resource to use (only applicable for Kubernetes engine).

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
    _typecheck_common_decorator_arguments(description, file_dependencies, requirements)

    if output is not None and not isinstance(output, str):
        raise InvalidUserArgumentException("`output` must be of type string if set.")

    def inner_decorator(func: MetricFunction) -> OutputArtifactFunction:
        nonlocal name
        nonlocal description
        if name is None or not isinstance(name, str):
            name = func.__name__
        if description is None:
            description = func.__doc__ or ""

        def _wrapped_util(
            *input_artifacts: BaseArtifact, execution_mode: Optional[ExecutionMode] = None
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

            if execution_mode == None:
                execution_mode = _get_global_execution_mode()

            assert isinstance(execution_mode, ExecutionMode)

            if len(input_artifacts) == 0:
                raise InvalidUserArgumentException(
                    "Metrics must have an input. Did you forget to call this metric on an artifact?"
                )

            artifacts = _convert_input_arguments_to_parameters(
                *input_artifacts,
                op_name=name,
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
            op_spec = OperatorSpec(metric=metric_spec)
            _update_operator_spec_with_engine(op_spec, engine)
            _update_operator_spec_with_resources(op_spec, resources)

            output_names = [output] if output is not None else None
            numeric_artifact = wrap_spec(
                op_spec,
                *artifacts,
                op_name=name,
                output_artifact_names=output_names,
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
            return _wrapped_util(*input_artifacts)

        # Enable the .local(*args) attribute, which calls the original function with the raw inputs.
        def local_func(*inputs: Any) -> Number:
            raw_inputs = [elem.get() if _is_input_artifact(elem) else elem for elem in inputs]
            return func(*raw_inputs)

        setattr(wrapped, "local", local_func)

        def lazy_mode(*input_artifacts: BaseArtifact) -> NumericArtifact:
            return _wrapped_util(*input_artifacts, execution_mode=ExecutionMode.LAZY)

        setattr(wrapped, "lazy", lazy_mode)

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
    output: Optional[str] = None,
    engine: Optional[str] = None,
    resources: Optional[Dict[str, Any]] = None,
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
            and install those. Otherwise, the method raises RequirementsMissingError exception.
        output:
            An optional custom name for the output metric artifact. Otherwise, the default naming scheme
            will be used.
        engine:
            The name of the compute integration this operator will run on.
        resources:
            A dictionary containing the custom resource configurations that this operator will run with.
            These configurations are guaranteed to be followed, we will not silently ignore any of them.
            If a resource configuration is unsupported by a particular execution engine, we will fail at
            execution time. The supported keys are:

            "num_cpus" (int):
                The number of cpus that this operator will run with. This operator will execute with *exactly*
                this number of cpus. If not enough cpus are available, operator execution will fail.
            "memory" (int, str):
                The amount of memory this operator will run with. This operator will execute with *exactly*
                this amount of memory. If not enough memory is available, operator execution will fail.

                If an integer value is supplied, the memory unit is assumed to be MB. If a string is supplied,
                a suffix indicating the memory unit must be supplied. Supported memory units are "MB" and "GB",
                case-insensitive.

                For example, the following values are valid: 100, "100MB", "1GB", "100mb", "1gb".
            "gpu_resource_name" (str):
                Name of the gpu resource to use (only applicable for Kubernetes engine).

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
    _typecheck_common_decorator_arguments(description, file_dependencies, requirements)

    if output is not None and not isinstance(output, str):
        raise InvalidUserArgumentException("`output` must be of type string if set.")

    def inner_decorator(func: CheckFunction) -> OutputArtifactFunction:
        nonlocal name
        nonlocal description
        if name is None or not isinstance(name, str):
            name = func.__name__
        if description is None:
            description = func.__doc__ or ""

        def _wrapped_util(
            *input_artifacts: BaseArtifact, execution_mode: Optional[ExecutionMode] = None
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

            if execution_mode == None:
                execution_mode = _get_global_execution_mode()

            assert isinstance(execution_mode, ExecutionMode)

            if len(input_artifacts) == 0:
                raise InvalidUserArgumentException(
                    "Check must have an input. Did you forget to call this check on an artifact?"
                )

            artifacts = _convert_input_arguments_to_parameters(
                *input_artifacts,
                op_name=name,
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
            op_spec = OperatorSpec(check=check_spec)
            _update_operator_spec_with_engine(op_spec, engine)
            _update_operator_spec_with_resources(op_spec, resources)

            output_names = [output] if output is not None else None
            bool_artifact = wrap_spec(
                op_spec,
                *artifacts,
                op_name=name,
                output_artifact_names=output_names,
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
            return _wrapped_util(*input_artifacts)

        # Enable the .local(*args) attribute, which calls the original function with the raw inputs.
        def local_func(*inputs: Any) -> Union[bool, np.bool_]:
            raw_inputs = [elem.get() if _is_input_artifact(elem) else elem for elem in inputs]
            return func(*raw_inputs)

        setattr(wrapped, "local", local_func)

        def lazy_mode(*input_artifacts: BaseArtifact) -> BoolArtifact:
            return _wrapped_util(*input_artifacts, execution_mode=ExecutionMode.LAZY)

        setattr(wrapped, "lazy", lazy_mode)

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


def _get_global_execution_mode() -> ExecutionMode:
    if globals.__GLOBAL_CONFIG__.lazy:
        return ExecutionMode.LAZY
    else:
        return ExecutionMode.EAGER

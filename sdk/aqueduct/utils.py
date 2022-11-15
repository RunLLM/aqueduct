import base64
import inspect
import json
import os
import shutil
import subprocess
import sys
import uuid
from datetime import datetime
from pathlib import Path
from typing import Any, Callable, Dict, List, Mapping, Optional, Union

import cloudpickle as pickle
import multipart
import numpy as np
import pkg_resources
import requests
from aqueduct.config import (
    AirflowEngineConfig,
    EngineConfig,
    FlowConfig,
    K8sEngineConfig,
    LambdaEngineConfig,
)
from aqueduct.dag import DAG, RetentionPolicy, Schedule
from aqueduct.enums import ArtifactType, OperatorType, RuntimeType, ServiceType, TriggerType
from aqueduct.error import *
from aqueduct.integrations.airflow_integration import AirflowIntegration
from aqueduct.integrations.integration import IntegrationInfo
from aqueduct.integrations.k8s_integration import K8sIntegration
from aqueduct.integrations.lambda_integration import LambdaIntegration
from aqueduct.logger import logger
from aqueduct.operators import Operator, ParamSpec
from aqueduct.serialization import (
    DEFAULT_ENCODING,
    artifact_type_to_serialization_type,
    make_temp_dir,
    serialization_function_mapping,
    serialize_val,
)
from aqueduct.templates import op_file_content
from croniter import croniter
from pandas import DataFrame
from PIL import Image
from requests_toolbelt.multipart import decoder

GITHUB_ISSUE_LINK = "https://github.com/aqueducthq/aqueduct/issues/new?assignees=&labels=bug&template=bug_report.md&title=%5BBUG%5D"


def format_header_for_print(header: str) -> str:
    """Used to print the header of a section in "describe()" with a consistent length.

    Sandwiches the "header" argument with a repeating sequence of "="s. Eg.

    ============================ "predict" artifact ==========================
    [       prefix len          ]
    [                              full_len                                   ]
    """
    prefix_len = 20
    full_len = 80
    return f"{'=' * prefix_len} {header} {'=' * max(0, full_len - prefix_len - len(header))}"


def generate_uuid() -> uuid.UUID:
    return uuid.uuid4()


WORKFLOW_UI_ROUTE_TEMPLATE = "/workflow/%s"
WORKFLOW_RUN_UI_ROUTE_TEMPLATE = "?workflowDagResultId=%s"


def generate_ui_url(
    aqueduct_base_address: str, workflow_id: str, result_id: Optional[str] = None
) -> str:
    if result_id:
        url = "%s%s%s" % (
            aqueduct_base_address,
            WORKFLOW_UI_ROUTE_TEMPLATE % workflow_id,
            WORKFLOW_RUN_UI_ROUTE_TEMPLATE % result_id,
        )
    else:
        url = "%s%s" % (
            aqueduct_base_address,
            WORKFLOW_UI_ROUTE_TEMPLATE % workflow_id,
        )
    return url


def is_string_valid_uuid(value: str) -> bool:
    try:
        uuid.UUID(str(value))
        return True
    except ValueError:
        return False


def raise_errors(response: requests.Response) -> None:
    def _extract_err_msg() -> str:
        resp_json = response.json()
        if "error" not in resp_json:
            raise Exception("No 'error' field on response: %s" % json.dumps(resp_json))
        return str(resp_json["error"])

    if response.status_code == 400:
        raise InvalidRequestError(_extract_err_msg())
    if response.status_code == 403:
        raise ClientValidationError(_extract_err_msg())
    elif response.status_code == 422:
        raise UnprocessableEntityError(_extract_err_msg())
    elif response.status_code == 500:
        raise InternalServerError(_extract_err_msg())
    elif response.status_code == 404:
        raise ResourceNotFoundError(_extract_err_msg())
    elif response.status_code != 200:
        raise AqueductError(_extract_err_msg())


def schedule_from_cron_string(schedule_str: str) -> Schedule:
    if len(schedule_str) == 0:
        return Schedule(trigger=TriggerType.MANUAL)

    if not croniter.is_valid(schedule_str):
        raise InvalidCronStringException("%s is not a valid cron string!" % schedule_str)

    return Schedule(trigger=TriggerType.PERIODIC, cron_schedule=schedule_str)


def retention_policy_from_latest_runs(k_latest_runs: int) -> RetentionPolicy:
    return RetentionPolicy(k_latest_runs=k_latest_runs)


# Helpers for creating model zip file

MODEL_FILE_NAME = "model.py"
MODEL_PICKLE_FILE_NAME = "model.pkl"
PYTHON_VERSION_FILE_NAME = "python_version.txt"
CONDA_VERSION_FILE_NAME = "conda_version.txt"
RESERVED_FILE_NAMES = [
    MODEL_FILE_NAME,
    MODEL_PICKLE_FILE_NAME,
    PYTHON_VERSION_FILE_NAME,
    CONDA_VERSION_FILE_NAME,
]
REQUIREMENTS_FILE = "requirements.txt"
BLACKLISTED_REQUIREMENTS = ["aqueduct_ml", "aqueduct_sdk", "aqueduct-ml", "aqueduct-sdk"]

UserFunction = Callable[..., Any]
Number = Union[int, float, np.number]
MetricFunction = Callable[..., Number]
CheckFunction = Callable[..., Union[bool, np.bool_]]


def get_zip_file_path(dir_name: str) -> str:
    return dir_name + ".zip"


def delete_zip_folder_and_file(dir_name: str) -> None:
    zip_file_path = get_zip_file_path(dir_name)

    if os.path.isfile(zip_file_path):
        os.remove(zip_file_path)

    if os.path.exists(dir_name):
        shutil.rmtree(dir_name)


def serialize_function(
    func: Union[UserFunction, MetricFunction, CheckFunction],
    op_name: str,
    file_dependencies: Optional[List[str]] = None,
    requirements: Optional[Union[str, List[str]]] = None,
) -> bytes:
    """
    Takes a user-defined function and packages it into a zip file structure expected by the backend.

    Arguments:
        func:
            The function to package
        op_name:
            The name of the function operator to package.
        file_dependencies:
            A list of relative paths to files that the function needs to access.
        requirements:
            Defines the python package requirements that this function will run with.
            Can be either a path to the requirements.txt file or a list of pip requirements specifiers.
            (eg. ["transformers==4.21.0", "numpy==1.22.4"]. If not supplied, we'll first
            look for a `requirements.txt` file in the same directory as the decorated function
            and install those. Otherwise, we'll attempt to infer the requirements with
            `pip freeze`.

    Returns:
        filepath of zip file in string format
    """
    dir_path = None
    try:
        dir_path = make_temp_dir()
        _package_files_and_requirements(
            func, os.path.join(os.getcwd(), dir_path), file_dependencies, requirements
        )

        # Figure out the python version
        python_version = ".".join((str(x) for x in sys.version_info[:2]))
        with open(os.path.join(dir_path, PYTHON_VERSION_FILE_NAME), "w") as f:
            f.write(python_version)

        with open(os.path.join(dir_path, MODEL_FILE_NAME), "w") as model_file:
            model_file.write(op_file_content())
        with open(os.path.join(dir_path, MODEL_PICKLE_FILE_NAME), "wb") as f:
            pickle.dump(func, f)

        # Write function source code to file
        source_file = "{}.py".format(op_name)
        with open(os.path.join(dir_path, source_file), "w") as f:
            try:
                source = inspect.getsource(func)
            except Exception:
                source = "unknown"
            f.write(source)

        zip_file_path = get_zip_file_path(dir_path)
        _make_archive(dir_path, zip_file_path)
        return open(zip_file_path, "rb").read()
    finally:
        if dir_path:
            delete_zip_folder_and_file(dir_path)


def _package_files_and_requirements(
    func: Union[UserFunction, MetricFunction, CheckFunction],
    dir_path: str,
    file_dependencies: Optional[List[str]] = None,
    requirements: Optional[Union[str, List[str]]] = None,
) -> None:
    """
    Populates the given dir_path directory with all the file dependencies and requirements.txt.

    Arguments:
        func:
            User-defined function to package
        dir_path:
            Absolute path of directory we'll be using
        file_dependencies:
            A list of relative paths to files that the function needs to access.
        requirements:
            Can be either a path to the requirements.txt file or a list of pip requirements specifiers.
            (eg. ["transformers==4.21.0", "numpy==1.22.4"]. If not supplied, we'll first
            look for a `requirements.txt` file in the same directory as the decorated function
            and install those. Otherwise, we'll attempt to infer the requirements with
            `pip freeze`.
    """
    if not file_dependencies:
        file_dependencies = []

    current_directory_path = os.getcwd()

    func_filepath = inspect.getsourcefile(func)
    if not func_filepath:
        raise Exception("Unable to find source file of function.")
    # In Python3.8, `inspect.getsourcefile` only returns the file's relative path,
    # so we need the line below to get the absolute path.
    func_filepath = os.path.abspath(func_filepath)
    func_dirpath = os.path.dirname(func_filepath)

    # We check if the directory `func_dirpath` exists. If not, this means `func` is from within a
    # Jupyter notebook that the user is currently running, so we don't switch the working directory.
    # The goal of switching the working directory is that if a user specifies relative paths
    # in `file_dependencies` and if `func` is imported from a Python script located in another
    # directory, we can locate them.
    if os.path.isdir(func_dirpath):
        os.chdir(func_dirpath)

    for file_index, file_path in enumerate(file_dependencies):
        if file_path in RESERVED_FILE_NAMES:
            # If the user uploads a `model.py` file as a dependency, we will error out.
            raise ReservedFileNameException(
                "%s is a reserved file name in our system. Please rename your file. " % file_path
            )
        if not os.path.exists(file_path):
            raise InvalidFunctionException("File %s does not exist" % file_path)

        if not os.path.abspath(file_path).startswith(os.getcwd()):
            raise InvalidDependencyFilePath(
                "File %s cannot be outside of the directory containing the function" % file_path
            )

        dstfolder = os.path.dirname(os.path.join(dir_path, file_path))
        if not os.path.exists(dstfolder):
            os.makedirs(dstfolder)
        shutil.copy(file_path, os.path.join(dir_path, file_path))

    # This is the absolute path to the requirements file we are sending to the backend.
    packaged_requirements_path = os.path.join(dir_path, REQUIREMENTS_FILE)
    if requirements is not None:
        # The operator has a custom requirements specification.
        assert isinstance(requirements, str) or all(isinstance(req, str) for req in requirements)

        if isinstance(requirements, str):
            if os.path.exists(requirements):
                logger().info("Installing requirements found at {path}".format(path=requirements))
                shutil.copy(requirements, packaged_requirements_path)
            else:
                raise FileNotFoundError(
                    "Requirements file provided at %s does not exist." % requirements
                )
        else:
            # User has given us a list of pip requirement strings.
            with open(packaged_requirements_path, "x") as f:
                f.write("\n".join(requirements))

    elif os.path.exists(REQUIREMENTS_FILE):
        # There exists a workflow-level requirements file (need to reside in the same directory as the function).
        logger().info(
            "%s: requirements.txt file detected in current directory %s, will not self-generate by inferring package dependencies."
            % (func.__name__, os.getcwd())
        )
        shutil.copy(REQUIREMENTS_FILE, packaged_requirements_path)

    else:
        # No requirements have been provided, so we use `pip freeze` to infer.
        logger().info(
            "%s: No requirements.txt file detected, self-generating file by inferring package dependencies."
            % func.__name__
        )
        with open(packaged_requirements_path, "x") as f:
            f.write("\n".join(_infer_requirements()))

    # Prune out any blacklisted requirements.
    _filter_out_blacklisted_requirements(packaged_requirements_path)

    _add_cloudpickle_to_requirements(packaged_requirements_path)

    os.chdir(current_directory_path)


def _filter_out_blacklisted_requirements(packaged_requirements_path: str) -> None:
    """Opens the requirements.txt file and removes any packages that we don't support."""
    with open(packaged_requirements_path, "r") as f:
        req_lines = f.readlines()

    with open(packaged_requirements_path, "w") as f:
        for line in req_lines:
            if any(blacklisted_req in line for blacklisted_req in BLACKLISTED_REQUIREMENTS):
                continue
            f.write(line)


def _add_cloudpickle_to_requirements(packaged_requirements_path: str) -> None:
    """
    Regardless of how we detect dependencies, we must include cloudpickle (with client's version
    number) as a requirement because the server needs to install the same version of cloudpickle as
    the client.

    If the user-specified requirements file already contains a cloudpickle entry and the version
    number matches, the installation process will still succeed. If there is a mismatch, the installation
    will fail and the user should fix the version number to match the version installed on the client.
    """
    with open(packaged_requirements_path, "a") as f:
        cloudpickle_requirement = (
            "\ncloudpickle==%s" % pkg_resources.get_distribution("cloudpickle").version
        )
        f.write(cloudpickle_requirement)


def _infer_requirements() -> List[str]:
    """
    Obtains the list of pip requirements specifiers from the current python environment using `pip freeze`.

    Returns:
        A list, for example, ["transformers==4.21.0", "numpy==1.22.4"].
    """
    try:
        process = subprocess.Popen(
            f"{sys.executable} -m pip freeze",
            shell=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )
        stdout_raw, stderr_raw = process.communicate()
        logger().debug("Inferred requirements raw stdout: %s", stdout_raw)
        logger().debug("Inferred requirements raw stderr: %s", stderr_raw)

        return stdout_raw.decode("utf-8").split("\n")
    except Exception as e:
        raise InternalAqueductError("Unable to infer requirements. Error: %s" % e)


def _make_archive(source: str, destination: str) -> None:
    """Creates zip file from source directory to destination file."""
    base = os.path.basename(destination)
    split_base = base.split(".")
    name = Path(".".join(split_base[:-1])).name
    format = split_base[-1]
    archive_from = os.path.dirname(source)
    if not archive_from:
        archive_from = "."
    archive_to = os.path.basename(source.strip(os.sep))
    shutil.make_archive(name, format, archive_from, archive_to)
    shutil.move("%s.%s" % (name, format), destination)


def artifact_name_from_op_name(op_name: str) -> str:
    return op_name + " artifact"


def generate_extract_op_name(
    dag: DAG,
    integration_name: str,
    name: Optional[str],
) -> str:
    """
    Generates name for extract operators to avoid operators with the same name.

    Arguments:
        dag:
            DAG that operator will be a part of.
        integration_name:
            Name of integration to run extract on.
        name:
            Optinally provided operator name.
    Returns:
        Name for extract operator.
    """

    op_name = name

    default_op_prefix = "%s query" % integration_name
    default_op_index = 1
    while op_name is None:
        candidate_op_name = default_op_prefix + " %d" % default_op_index
        colliding_op = dag.get_operator(with_name=candidate_op_name)
        if colliding_op is None:
            op_name = candidate_op_name  # break out of the loop!
        default_op_index += 1

    assert op_name is not None

    return op_name


def get_checks_for_op(op: Operator, dag: DAG) -> List[Operator]:
    check_operators = []
    for artf in op.outputs:
        check_operators.extend(
            dag.list_operators(
                filter_to=[OperatorType.CHECK],
                on_artifact_id=artf,
            )
        )
    return check_operators


def get_metrics_for_op(op: Operator, dag: DAG) -> List[Operator]:
    metric_operators = []
    for artf in op.outputs:
        metric_operators.extend(
            dag.list_operators(
                filter_to=[OperatorType.METRIC],
                on_artifact_id=artf,
            )
        )
    return metric_operators


def get_description_for_check(check: Operator) -> Dict[str, str]:
    check_spec = check.spec.check
    if check_spec:
        level = check_spec.level
    else:
        raise AqueductError("Check artifact malformed.")
    return {
        "Label": check.name,
        "Description": check.description,
        "Level": level,
    }


def get_description_for_metric(
    metric: Operator, dag: DAG
) -> Dict[str, Union[str, List[Mapping[str, Any]]]]:
    metric_spec = metric.spec.metric
    if metric_spec:
        granularity = metric_spec.function.granularity
    else:
        raise AqueductError("Metric artifact malformed.")
    return {
        "Label": metric.name,
        "Description": metric.description,
        "Granularity": granularity,
        "Checks": [
            get_description_for_check(check_op) for check_op in get_checks_for_op(metric, dag)
        ],
        "Metrics": [
            get_description_for_metric(metric_op, dag)
            for metric_op in get_metrics_for_op(metric, dag)
        ],
    }


def human_readable_timestamp(ts: int) -> str:
    format = "%Y-%m-%d %H:%M:%S"
    return datetime.utcfromtimestamp(ts).strftime(format)


def indent_multiline_string(content: str) -> str:
    """Indents every line of a multiline string block."""
    return "\t" + "\t".join(content.splitlines(True))


def parse_user_supplied_id(id: Union[str, uuid.UUID]) -> str:
    """Verifies that a user-defined id is of the expected types, returning the string version of the id."""
    if not isinstance(id, str) and not isinstance(id, uuid.UUID):
        raise InvalidUserArgumentException("Provided id must be either str or uuid.")

    if isinstance(id, uuid.UUID):
        return str(id)
    return id


def infer_artifact_type(value: Any) -> ArtifactType:
    if isinstance(value, DataFrame):
        return ArtifactType.TABLE
    elif isinstance(value, Image.Image):
        return ArtifactType.IMAGE
    elif isinstance(value, bytes):
        return ArtifactType.BYTES
    elif isinstance(value, str):
        # We first check if the value is a valid JSON string.
        try:
            json.loads(value)
            return ArtifactType.JSON
        except:
            return ArtifactType.STRING
    elif isinstance(value, bool) or isinstance(value, np.bool_):
        return ArtifactType.BOOL
    elif isinstance(value, int) or isinstance(value, float) or isinstance(value, np.number):
        return ArtifactType.NUMERIC
    elif isinstance(value, dict):
        return ArtifactType.DICT
    elif isinstance(value, tuple):
        return ArtifactType.TUPLE
    elif isinstance(value, list):
        return ArtifactType.LIST
    else:
        try:
            pickle.dumps(value)
            return ArtifactType.PICKLABLE
        except:
            pass

        try:
            # tf.keras.Model's can be pickled, but some classes that inherit from it cannot (eg. `tfrs.Model`)
            from tensorflow import keras

            if isinstance(value, keras.Model):
                return ArtifactType.TF_KERAS
        except:
            pass

        raise Exception("Failed to map type %s to supported artifact type." % type(value))


def _bytes_to_base64_string(content: bytes) -> str:
    """Helper to convert bytes to a base64-string.

    For example, image-serialized bytes are not `utf8` encoded, so if we want to convert
    such bytes to string, we must use this function.
    """
    return base64.b64encode(content).decode(DEFAULT_ENCODING)


def construct_param_spec(val: Any, artifact_type: ArtifactType) -> ParamSpec:
    serialization_type = artifact_type_to_serialization_type(artifact_type, val)
    assert serialization_type in serialization_function_mapping

    # We must base64 encode the resulting bytes, since we can't be sure
    # what encoding it was written in (eg. Image types are not encoded as "utf8").
    return ParamSpec(
        val=_bytes_to_base64_string(serialize_val(val, serialization_type)),
        serialization_type=serialization_type,
    )


def parse_artifact_result_response(response: requests.Response) -> Dict[str, Any]:
    multipart_data = decoder.MultipartDecoder.from_response(response)
    parse = multipart.parse_options_header

    result = {}

    for part in multipart_data.parts:
        field_name = part.headers[b"Content-Disposition"].decode(multipart_data.encoding)
        field_name = parse(field_name)[1]["name"]

        if field_name == "metadata":
            result[field_name] = json.loads(part.content.decode(multipart_data.encoding))
        elif field_name == "data":
            result[field_name] = part.content
        else:
            raise AqueductError(
                "Unexpected form field %s for artifact result response" % field_name
            )

    return result


def generate_engine_config(integration: IntegrationInfo) -> EngineConfig:
    """Generates an EngineConfig from an integration info object."""
    if integration.service == ServiceType.AIRFLOW:
        return EngineConfig(
            type=RuntimeType.AIRFLOW,
            airflow_config=AirflowEngineConfig(
                integration_id=integration.id,
            ),
        )
    elif integration.service == ServiceType.K8S:
        return EngineConfig(
            type=RuntimeType.K8S,
            k8s_config=K8sEngineConfig(
                integration_id=integration.id,
            ),
        )
    elif integration.service == ServiceType.LAMBDA:
        return EngineConfig(
            type=RuntimeType.LAMBDA,
            lambda_config=LambdaEngineConfig(
                integration_id=integration.id,
            ),
        )
    else:
        raise AqueductError("Unsupported engine configuration.")

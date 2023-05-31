import uuid
from typing import Any, Dict, List, Optional, Tuple, Union

from aqueduct.constants.enums import ArtifactType, RuntimeType, ServiceType, TriggerType
from aqueduct.error import *
from aqueduct.models.config import (
    AirflowEngineConfig,
    DatabricksEngineConfig,
    EngineConfig,
    K8sEngineConfig,
    LambdaEngineConfig,
    SparkEngineConfig,
)
from aqueduct.models.dag import Schedule
from aqueduct.models.operators import ParamSpec
from aqueduct.models.resource import ResourceInfo
from aqueduct.resources.dynamic_k8s import DynamicK8sResource
from aqueduct.utils.resource_validation import validate_resource_is_connected
from croniter import croniter

from ..models.execution_state import Logs
from .serialization import artifact_type_to_serialization_type, serialize_val
from .type_inference import _bytes_to_base64_string


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
WORKFLOW_RUN_UI_ROUTE_TEMPLATE = "/result/%s"


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


def generate_flow_schedule(
    schedule_str: str, source_flow_id: Optional[uuid.UUID] = None
) -> Schedule:
    """Generates a flow schedule using the provided cron string and the source flow id if present."""
    if source_flow_id:
        return Schedule(trigger=TriggerType.CASCADE, source_id=source_flow_id)

    if len(schedule_str) == 0:
        return Schedule(trigger=TriggerType.MANUAL)

    if not croniter.is_valid(schedule_str):
        raise InvalidCronStringException("%s is not a valid cron string!" % schedule_str)

    return Schedule(trigger=TriggerType.PERIODIC, cron_schedule=schedule_str)


def indent_multiline_string(content: str) -> str:
    """Indents every line of a multiline string block."""
    return "\t" + "\t".join(content.splitlines(True))


def print_logs(logs: Logs) -> None:
    """Prints out the logs with the following format:

    stdout:
        {logs}
        {logs}
    ----------------------------------
    stderr:
        {logs}
        {logs}
    """
    if len(logs.stdout) > 0:
        print("stdout:")
        print(indent_multiline_string(logs.stdout).rstrip("\n"))

    if len(logs.stdout) > 0 and len(logs.stderr) > 0:
        print("----------------------------------")

    if len(logs.stderr) > 0:
        print("stderr:")
        print(indent_multiline_string(logs.stderr).rstrip("\n"))


def parse_user_supplied_id(id: Union[str, uuid.UUID]) -> str:
    """Verifies that a user-defined id is of the expected types, returning the string version of the id."""
    if not isinstance(id, str) and not isinstance(id, uuid.UUID):
        raise InvalidUserArgumentException("Provided id must be either str or uuid.")

    if isinstance(id, uuid.UUID):
        return str(id)
    return id


def construct_param_spec(
    val: Any,
    artifact_type: ArtifactType,
) -> ParamSpec:
    # Not derived from bson.
    # For now, bson_table applies only to tables read from mongo.
    derived_from_bson = False

    serialization_type = artifact_type_to_serialization_type(
        artifact_type,
        derived_from_bson,
        val,
    )

    # We must base64 encode the resulting bytes, since we can't be sure
    # what encoding it was written in (eg. Image types are not encoded as "utf8").
    return ParamSpec(
        val=_bytes_to_base64_string(serialize_val(val, serialization_type, derived_from_bson)),
        serialization_type=serialization_type,
    )


def generate_engine_config(
    resources: Dict[str, ResourceInfo],
    resource_name: Optional[Union[str, DynamicK8sResource]],
) -> Optional[EngineConfig]:
    """Generates an EngineConfig from an resource info object.

    Both None and "Aqueduct" (case-insensitive) map to the Aqueduct Engine.
    """
    if isinstance(resource_name, DynamicK8sResource):
        resource_name = resource_name.name()

    if resource_name is None or resource_name.lower() == "aqueduct":
        return None

    if resource_name not in resources.keys():
        raise InvalidResourceException("Not connected to compute resource `%s`!" % resource_name)

    resource = resources[resource_name]
    validate_resource_is_connected(resource_name, resource.exec_state)

    if resource.service == ServiceType.AIRFLOW:
        return EngineConfig(
            type=RuntimeType.AIRFLOW,
            name=resource_name,
            airflow_config=AirflowEngineConfig(
                resource_id=resource.id,
            ),
        )
    elif resource.service == ServiceType.K8S:
        return EngineConfig(
            type=RuntimeType.K8S,
            name=resource_name,
            k8s_config=K8sEngineConfig(
                resource_id=resource.id,
            ),
        )
    elif resource.service == ServiceType.LAMBDA:
        return EngineConfig(
            type=RuntimeType.LAMBDA,
            name=resource_name,
            lambda_config=LambdaEngineConfig(
                resource_id=resource.id,
            ),
        )
    elif resource.service == ServiceType.DATABRICKS:
        return EngineConfig(
            type=RuntimeType.DATABRICKS,
            name=resource_name,
            databricks_config=DatabricksEngineConfig(
                resource_id=resource.id,
            ),
        )
    elif resource.service == ServiceType.SPARK:
        return EngineConfig(
            type=RuntimeType.SPARK,
            name=resource_name,
            spark_config=SparkEngineConfig(
                resource_id=resource.id,
            ),
        )
    else:
        raise AqueductError("Unsupported engine configuration.")


def find_flow_with_user_supplied_id_and_name(
    flows: List[Tuple[uuid.UUID, str]], flow_identifier: Optional[Union[str, uuid.UUID]] = None
) -> str:
    """Verifies that the user supplied flow_identifier correspond
    to an actual flow in `flows`. Must be either uuid or name that
    must match to the same flow. It returns the string version of
    the matching flow's id.
    """
    if not flow_identifier:
        raise InvalidUserArgumentException(
            "Must supply a valid flow identifier, either name or uuid"
        )

    flow_id_str = parse_user_supplied_id(flow_identifier)

    if isinstance(flow_identifier, uuid.UUID) or is_string_valid_uuid(flow_identifier):
        if all(uuid.UUID(flow_id_str) != flow[0] for flow in flows):
            raise InvalidUserArgumentException("Unable to find a flow with id %s" % flow_identifier)
    elif isinstance(flow_identifier, str):
        if all(flow_id_str != flow[1] for flow in flows):
            raise InvalidUserArgumentException("Unable to find a flow with name %s" % flow_id_str)
        # You land here if found matching flow name, return corresponding uuid
        # backend api's look for uuid instead of name
        for flow in flows:
            if flow_id_str == flow[1]:
                flow_id_str = str(flow[0])
                break

    return flow_id_str

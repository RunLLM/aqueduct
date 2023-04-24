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
from aqueduct.models.integration import IntegrationInfo
from aqueduct.models.operators import ParamSpec
from aqueduct.utils.integration_validation import validate_integration_is_connected
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
    integrations: Dict[str, IntegrationInfo], integration_name: Optional[str]
) -> Optional[EngineConfig]:
    """Generates an EngineConfig from an integration info object.

    Both None and "Aqueduct" (case-insensitive) map to the Aqueduct Engine.
    """
    if integration_name is None or integration_name.lower() == "aqueduct":
        return None

    if integration_name not in integrations.keys():
        raise InvalidIntegrationException(
            "Not connected to compute integration `%s`!" % integration_name
        )

    integration = integrations[integration_name]
    validate_integration_is_connected(integration_name, integration.exec_state)

    if integration.service == ServiceType.AIRFLOW:
        return EngineConfig(
            type=RuntimeType.AIRFLOW,
            name=integration_name,
            airflow_config=AirflowEngineConfig(
                integration_id=integration.id,
            ),
        )
    elif integration.service == ServiceType.K8S:
        return EngineConfig(
            type=RuntimeType.K8S,
            name=integration_name,
            k8s_config=K8sEngineConfig(
                integration_id=integration.id,
            ),
        )
    elif integration.service == ServiceType.LAMBDA:
        return EngineConfig(
            type=RuntimeType.LAMBDA,
            name=integration_name,
            lambda_config=LambdaEngineConfig(
                integration_id=integration.id,
            ),
        )
    elif integration.service == ServiceType.DATABRICKS:
        return EngineConfig(
            type=RuntimeType.DATABRICKS,
            name=integration_name,
            databricks_config=DatabricksEngineConfig(
                integration_id=integration.id,
            ),
        )
    elif integration.service == ServiceType.SPARK:
        return EngineConfig(
            type=RuntimeType.SPARK,
            name=integration_name,
            spark_config=SparkEngineConfig(
                integration_id=integration.id,
            ),
        )
    else:
        raise AqueductError("Unsupported engine configuration.")


def find_flow_with_user_supplied_id_and_name(
    flows: List[Tuple[uuid.UUID, str]],
    flow_id: Optional[Union[str, uuid.UUID]] = None,
    flow_name: Optional[str] = None,
) -> str:
    """Verifies that the user supplied flow id and name correspond
    to an actual flow in `flows`. Only one of `flow_id` and `flow_name` is necessary,
    but if both are provided, they must match to the same flow. It returns the
    string version of the matching flow's id.
    """
    if not flow_id and not flow_name:
        raise InvalidUserArgumentException(
            "Must supply at least one of the following:`flow_id` or `flow_name`"
        )

    if flow_id:
        flow_id_str = parse_user_supplied_id(flow_id)
        if all(uuid.UUID(flow_id_str) != flow[0] for flow in flows):
            raise InvalidUserArgumentException("Unable to find a flow with id %s" % flow_id)

    if flow_name:
        flow_id_str_from_name = None
        for flow in flows:
            if flow[1] == flow_name:
                flow_id_str_from_name = str(flow[0])
                break

        if not flow_id_str_from_name:
            raise InvalidUserArgumentException("Unable to find a flow with name %s" % flow_name)

        if flow_id and flow_id_str != flow_id_str_from_name:
            # User supplied both flow_id and flow_name, but they do not
            # correspond to the same flow
            raise InvalidUserArgumentException(
                "The flow with id %s does not correspond to the flow with name %s"
                % (flow_id, flow_name)
            )

        return flow_id_str_from_name

    return flow_id_str

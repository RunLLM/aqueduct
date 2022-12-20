import uuid
from datetime import datetime
from typing import Any, Dict, Optional, Union

from aqueduct.constants.enums import ArtifactType, RuntimeType, ServiceType, TriggerType
from aqueduct.error import *
from aqueduct.models.config import (
    AirflowEngineConfig,
    EngineConfig,
    K8sEngineConfig,
    LambdaEngineConfig,
)
from aqueduct.models.dag import Schedule
from aqueduct.models.integration import IntegrationInfo
from aqueduct.models.operators import ParamSpec
from croniter import croniter

from .serialization import (
    artifact_type_to_serialization_type,
    serialization_function_mapping,
    serialize_val,
)
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


def schedule_from_cron_string(schedule_str: str) -> Schedule:
    if len(schedule_str) == 0:
        return Schedule(trigger=TriggerType.MANUAL)

    if not croniter.is_valid(schedule_str):
        raise InvalidCronStringException("%s is not a valid cron string!" % schedule_str)

    return Schedule(trigger=TriggerType.PERIODIC, cron_schedule=schedule_str)


def artifact_name_from_op_name(op_name: str) -> str:
    return op_name + " artifact"


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


def construct_param_spec(val: Any, artifact_type: ArtifactType) -> ParamSpec:
    serialization_type = artifact_type_to_serialization_type(
        artifact_type,
        # Not derived from bson.
        # For now, bson_table applies only to tables read from mongo.
        False,
        val,
    )
    assert serialization_type in serialization_function_mapping

    # We must base64 encode the resulting bytes, since we can't be sure
    # what encoding it was written in (eg. Image types are not encoded as "utf8").
    return ParamSpec(
        val=_bytes_to_base64_string(serialize_val(val, serialization_type)),
        serialization_type=serialization_type,
    )


def generate_engine_config(
    integrations: Dict[str, IntegrationInfo], integration_name: Optional[str]
) -> Optional[EngineConfig]:
    """Generates an EngineConfig from an integration info object."""
    if integration_name is None:
        return None

    if integration_name not in integrations.keys():
        raise InvalidIntegrationException(
            "Not connected to compute integration `%s`!" % integration_name
        )

    integration = integrations[integration_name]
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

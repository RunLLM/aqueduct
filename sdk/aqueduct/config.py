import uuid
from typing import Optional, Union

from aqueduct.enums import RuntimeType
from aqueduct.integrations.airflow_integration import AirflowIntegration
from aqueduct.integrations.k8s_integration import K8sIntegration
from pydantic import BaseModel


class AqueductEngineConfig(BaseModel):
    pass


class AirflowEngineConfig(BaseModel):
    integration_id: uuid.UUID


class K8sEngineConfig(BaseModel):
    integration_id: uuid.UUID


class EngineConfig(BaseModel):
    # The runtime type dictates the engine config that is set.
    # We default to the AqueductEngine.
    type: RuntimeType = RuntimeType.AQUEDUCT
    aqueduct_config: Optional[AqueductEngineConfig]
    airflow_config: Optional[AirflowEngineConfig]
    k8s_config: Optional[K8sEngineConfig]


class FlowConfig(BaseModel):
    engine: Optional[Union[AirflowIntegration, K8sIntegration]]

    class Config:
        # Necessary to allow an engine field
        arbitrary_types_allowed = True

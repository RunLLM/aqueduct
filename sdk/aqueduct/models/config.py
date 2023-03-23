import uuid
from typing import Optional, Union

from aqueduct.constants.enums import RuntimeType
from aqueduct.integrations.airflow_integration import AirflowIntegration
from aqueduct.integrations.databricks_integration import DatabricksIntegration
from aqueduct.integrations.k8s_integration import K8sIntegration
from aqueduct.integrations.lambda_integration import LambdaIntegration
from aqueduct.integrations.spark_integration import SparkIntegration
from pydantic import BaseModel


class AqueductEngineConfig(BaseModel):
    pass


class AqueductCondaEngineConfig(BaseModel):
    env: str


class AirflowEngineConfig(BaseModel):
    integration_id: uuid.UUID


class K8sEngineConfig(BaseModel):
    integration_id: uuid.UUID


class LambdaEngineConfig(BaseModel):
    integration_id: uuid.UUID


class DatabricksEngineConfig(BaseModel):
    integration_id: uuid.UUID


class SparkEngineConfig(BaseModel):
    integration_id: uuid.UUID


class EngineConfig(BaseModel):
    # The runtime type dictates the engine config that is set.
    # We default to the AqueductEngine.
    type: RuntimeType = RuntimeType.AQUEDUCT
    aqueduct_config: Optional[AqueductEngineConfig]
    aqueduct_conda_config: Optional[AqueductCondaEngineConfig]
    airflow_config: Optional[AirflowEngineConfig]
    k8s_config: Optional[K8sEngineConfig]
    lambda_config: Optional[LambdaEngineConfig]
    databricks_config: Optional[DatabricksEngineConfig]
    spark_config: Optional[SparkEngineConfig]

    # The name of the compute integration. This not consumed by the backend,
    # but is instead only used for logging purposes in the SDK.
    name: str = "Aqueduct"

    class Config:
        fields = {
            "name": {"exclude": ...},
        }

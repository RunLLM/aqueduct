import uuid
from typing import Optional, Union

from aqueduct.constants.enums import RuntimeType
from aqueduct.resources.airflow import AirflowResource
from aqueduct.resources.aws_lambda import LambdaResource
from aqueduct.resources.databricks import DatabricksResource
from aqueduct.resources.k8s import K8sResource
from aqueduct.resources.spark import SparkResource
from pydantic import BaseModel


class AqueductEngineConfig(BaseModel):
    pass


class AqueductCondaEngineConfig(BaseModel):
    env: str


class AirflowEngineConfig(BaseModel):
    resource_id: uuid.UUID


class K8sEngineConfig(BaseModel):
    resource_id: uuid.UUID


class LambdaEngineConfig(BaseModel):
    resource_id: uuid.UUID


class DatabricksEngineConfig(BaseModel):
    resource_id: uuid.UUID


class SparkEngineConfig(BaseModel):
    resource_id: uuid.UUID


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

    # The name of the compute resource. This not consumed by the backend,
    # but is instead only used for logging purposes in the SDK.
    name: str = "Aqueduct"

    class Config:
        fields = {
            "name": {"exclude": ...},
        }

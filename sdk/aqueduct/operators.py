import uuid
from typing import List, Optional, Union

from aqueduct.enums import (
    ArtifactType,
    CheckSeverity,
    FunctionGranularity,
    FunctionType,
    GithubRepoConfigContentType,
    GoogleSheetsSaveMode,
    LoadUpdateMode,
    OperatorType,
    S3TableFormat,
    SalesforceExtractType,
    SerializationType,
    ServiceType,
)
from aqueduct.error import AqueductError
from aqueduct.integrations.integration import IntegrationInfo
from pydantic import BaseModel, Extra


class GithubMetadata(BaseModel):
    """
    Specifies a destination in github integration.
    There are two ways to specify the content:
    -   by `path`, which points to a file or dir in the github repo.
    -   from `repo_config_content_type` and `repo_config_content_name`, which points to
        information stored in the repo's `.aqconfig`.
    If using `repo_config` content, backend will ignore `path` and overwrite it with
    the `path` specified in `.aqconfig`.
    """

    owner: str
    repo: str
    branch: str
    path: Optional[str] = None
    repo_config_content_type: Optional[GithubRepoConfigContentType] = None
    repo_config_content_name: Optional[str] = None
    commit_id: Optional[str] = None


class RelationalDBExtractParams(BaseModel):
    """
    Specifies the query to run when extracting from a relational DB.
    Exactly one of the 3 fields should be set.

    query: the string to run a single query.
    queries: a list of strings to run a chain of queries.
    github_metadata: Github information to run a query stored in github.
    """

    query: Optional[str] = None
    queries: Optional[List[str]] = None
    github_metadata: Optional[GithubMetadata] = None


class SalesforceExtractParams(BaseModel):
    type: SalesforceExtractType
    query: str


class GoogleSheetsExtractParams(BaseModel):
    spreadsheet_id: str


class S3ExtractParams(BaseModel):
    # Note that since we expect the path to be either a string or a list of strings, we need to json
    # serialize the path before we pass it to initialize this field.
    filepath: str
    artifact_type: ArtifactType
    format: Optional[S3TableFormat]
    merge: Optional[bool]


class MongoExtractParams(BaseModel):
    collection: str
    query_serialized: str


UnionExtractParams = Union[
    SalesforceExtractParams,
    S3ExtractParams,
    GoogleSheetsExtractParams,
    MongoExtractParams,
    RelationalDBExtractParams,
]


class ExtractSpec(BaseModel):
    service: ServiceType
    integration_id: uuid.UUID
    parameters: UnionExtractParams


class RelationalDBLoadParams(BaseModel):
    table: str
    update_mode: LoadUpdateMode


class SalesforceLoadParams(BaseModel):
    object: str


class GoogleSheetsLoadParams(BaseModel):
    filepath: str
    save_mode: GoogleSheetsSaveMode


class S3LoadParams(BaseModel):
    filepath: str
    format: Optional[S3TableFormat]

    # Must do this to prevent confusion with GoogleSheetsLoadParams.
    class Config:
        extra = Extra.forbid


UnionLoadParams = Union[
    SalesforceLoadParams, S3LoadParams, GoogleSheetsLoadParams, RelationalDBLoadParams
]


# Internal class used by SDK to represent the config for loading to an integration.
class SaveConfig(BaseModel):
    integration_info: IntegrationInfo
    parameters: UnionLoadParams


# Class expected by backend for a load operator.
class LoadSpec(BaseModel):
    service: ServiceType
    integration_id: uuid.UUID
    parameters: UnionLoadParams


class EntryPoint(BaseModel):
    file: str
    class_name: Optional[str]
    method: str


class FunctionSpec(BaseModel):
    type: FunctionType
    language = "Python"
    granularity: FunctionGranularity
    github_metadata: Optional[GithubMetadata]
    entry_point: Optional[EntryPoint] = None

    # Function zip file.
    file: Optional[bytes] = None

    class Config:
        fields = {"file": {"exclude": ...}}


class MetricSpec(BaseModel):
    function: FunctionSpec


class SystemMetricSpec(BaseModel):
    metric_name: str


class CheckSpec(BaseModel):
    level: CheckSeverity
    function: FunctionSpec


class ParamSpec(BaseModel):
    # `val` is the base64-encoded version of the serialized param value.
    val: str
    serialization_type: SerializationType


# https://docs.aws.amazon.com/lambda/latest/operatorguide/computing-power.html
LAMBDA_MIN_MEMORY_MB = 128
LAMBDA_MAX_MEMORY_MB = 10240


class ResourceConfig(BaseModel):
    # These resources are configured exactly. The user is not given any more
    # or any less. If the requested resources exceeds capacity, an error
    # will be thrown at execution time.
    num_cpus: Optional[int]
    memory_mb: Optional[int]
    gpu_resource_name: Optional[str]


class OperatorSpec(BaseModel):
    extract: Optional[ExtractSpec]
    load: Optional[LoadSpec]
    function: Optional[FunctionSpec]
    metric: Optional[MetricSpec]
    check: Optional[CheckSpec]
    param: Optional[ParamSpec]
    system_metric: Optional[SystemMetricSpec]
    resources: Optional[ResourceConfig]


class Operator(BaseModel):
    id: uuid.UUID
    name: str
    description: str
    spec: OperatorSpec
    inputs: List[uuid.UUID] = []
    outputs: List[uuid.UUID] = []

    def file(self) -> Optional[bytes]:
        if self.spec.function:
            return self.spec.function.file
        if self.spec.metric:
            return self.spec.metric.function.file
        if self.spec.check:
            return self.spec.check.function.file

        return None

    def update_serialized_function(self, serialized_function: bytes) -> None:
        if self.spec.function:
            self.spec.function.file = serialized_function
        if self.spec.metric:
            self.spec.metric.function.file = serialized_function
        if self.spec.check:
            self.spec.check.function.file = serialized_function


def get_operator_type(operator: Operator) -> OperatorType:
    return get_operator_type_from_spec(operator.spec)


def get_operator_type_from_spec(spec: OperatorSpec) -> OperatorType:
    if spec.extract is not None:
        return OperatorType.EXTRACT
    if spec.load is not None:
        return OperatorType.LOAD
    if spec.function is not None:
        return OperatorType.FUNCTION
    if spec.metric is not None:
        return OperatorType.METRIC
    if spec.check is not None:
        return OperatorType.CHECK
    if spec.param is not None:
        return OperatorType.PARAM
    if spec.system_metric is not None:
        return OperatorType.SYSTEM_METRIC
    else:
        raise AqueductError("Invalid operator type")

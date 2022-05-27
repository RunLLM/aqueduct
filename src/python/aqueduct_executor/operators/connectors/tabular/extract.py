from typing import Any, Optional, Union

from aqueduct_executor.operators.connectors.tabular import common, models


class RelationalParams(models.BaseParams):
    query: str
    # TODO: Consider not including github as part of relational params when it is JSON marshalled
    github_metadata: Optional[Any]


class S3Params(models.BaseParams):
    filepath: str
    format: common.S3FileFormat


Params = Union[RelationalParams, S3Params]

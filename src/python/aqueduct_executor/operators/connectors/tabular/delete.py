from typing import Optional, Union

from pydantic import validator

from aqueduct_executor.operators.connectors.tabular import common, models


class RelationalParams(models.BaseParams):
    table: str


class S3Params(models.BaseParams):
    key: str


Params = Union[RelationalParams, S3Params]

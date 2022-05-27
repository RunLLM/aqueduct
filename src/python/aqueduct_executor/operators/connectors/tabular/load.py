from typing import Optional, Union

from pydantic import validator

from aqueduct_executor.operators.connectors.tabular import common, models


class RelationalParams(models.BaseParams):
    table: str
    update_mode: Optional[common.UpdateMode] = common.UpdateMode.REPLACE

    class Config:
        validate_assignment = True

    @validator("update_mode")
    def set_update_mode(cls, update_mode):
        if update_mode == "":
            return common.UpdateMode.REPLACE
        return update_mode


class S3Params(models.BaseParams):
    filepath: str
    format: common.S3FileFormat


Params = Union[RelationalParams, S3Params]

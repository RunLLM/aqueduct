from pydantic import BaseModel, Extra


class BaseSpec(BaseModel):
    """
    BaseSpec defines the Pydantic Config shared by all connector Spec's, e.g.
    AuthenticateSpec, ExtractSpec, etc.
    """

    class Config:
        extra = Extra.forbid  # Ensures extra fields are not allowed
        smart_union = True  # Prevents undesired coercion does not occur.


class BaseConfig(BaseModel):
    """
    BaseConfig defines the Pydantic Config shared by all connector Config's, e.g.
    postgres.Config, mysql.Config, etc.
    """

    class Config:
        extra = Extra.forbid


class BaseParams(BaseModel):
    """
    BaseParams defines the Pydantic Config shared by all ExtractParams and LoadParams, e.g.
    relational.ExtractParams, relational.LoadParams, etc.
    """

    class Config:
        extra = Extra.forbid

from typing import Optional, Union

from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.constants.enums import ArtifactType
from aqueduct.resources.s3 import S3Resource
from aqueduct.resources.sql import RelationalDBResource

from sdk.shared.data_objects import DataObject


def extract(
    integration,
    obj_identifier: Union[DataObject, str],
    op_name: Optional[str] = None,
    output_name: Optional[str] = None,
    lazy: bool = False,
) -> BaseArtifact:
    """Reads the specified object in from the integration and returns it as an artifact.

    Assumption: the object is a pandas dataframe, serialized in a particular fashion in each integration.
    This serialization method should match what is done in `save()`.
    """
    if isinstance(obj_identifier, DataObject):
        obj_identifier = obj_identifier.value

    assert isinstance(obj_identifier, str)
    if isinstance(integration, RelationalDBResource):
        return integration.sql(
            query="SELECT * from %s" % obj_identifier, name=op_name, output=output_name, lazy=lazy
        )
    elif isinstance(integration, S3Resource):
        return integration.file(
            obj_identifier,
            ArtifactType.TABLE,
            "parquet",
            name=op_name,
            output=output_name,
            lazy=lazy,
        )
    raise Exception("Unexpected data integration type provided in test: %s", type(integration))

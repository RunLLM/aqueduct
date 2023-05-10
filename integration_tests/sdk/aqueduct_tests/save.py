from typing import Optional

from aqueduct.artifacts.table_artifact import TableArtifact
from aqueduct.constants.enums import S3TableFormat
from aqueduct.resources.mongodb import MongoDBResource
from aqueduct.resources.s3 import S3Resource
from aqueduct.resources.sql import RelationalDBResource

from aqueduct import LoadUpdateMode

from ..shared.globals import artifact_id_to_saved_identifier
from ..shared.naming import generate_table_name


def save(
    integration,
    artifact: TableArtifact,
    name: Optional[str] = None,
    update_mode: Optional[LoadUpdateMode] = None,
):
    """Saves an artifact into the integration.

    If `name` is set, make sure that it is set to a globally unique value, since test cases can be run concurrently.

    Assumption: the artifact represents a pandas dataframe. Each type of integration is serialized in a particular fashion.
    It should match the deserialization method in extract().
    """
    if name is None:
        name = generate_table_name()
    if update_mode is None:
        update_mode = LoadUpdateMode.REPLACE

    if isinstance(integration, RelationalDBResource):
        integration.save(artifact, name, update_mode)

    elif isinstance(integration, S3Resource):
        assert update_mode == LoadUpdateMode.REPLACE, "S3 only supports replacement update."
        integration.save(artifact, name, S3TableFormat.PARQUET)
    elif isinstance(integration, MongoDBResource):
        integration.collection(name).save(artifact, update_mode)
    else:
        raise Exception("Unexpected data integration type provided in test: %s", type(integration))

    # Record where the artifact was saved, so we can validate the data later, after the flow is published.
    artifact_id_to_saved_identifier[str(artifact.id())] = name

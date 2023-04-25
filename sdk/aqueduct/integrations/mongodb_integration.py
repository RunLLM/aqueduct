import json
from typing import Any, Dict, List, Optional

from aqueduct.artifacts import preview as artifact_utils
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.table_artifact import TableArtifact
from aqueduct.constants.enums import ArtifactType, ExecutionMode, LoadUpdateMode
from aqueduct.integrations.parameters import _validate_parameters
from aqueduct.integrations.save import _save_artifact
from aqueduct.integrations.validation import validate_is_connected
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG
from aqueduct.models.integration import Integration, IntegrationInfo
from aqueduct.models.operators import (
    ExtractSpec,
    MongoExtractParams,
    Operator,
    OperatorSpec,
    RelationalDBLoadParams,
)
from aqueduct.utils.dag_deltas import AddOperatorDelta, apply_deltas_to_dag
from aqueduct.utils.naming import default_artifact_name_from_op_name, sanitize_artifact_name
from aqueduct.utils.utils import generate_uuid

from aqueduct import globals


class MongoDBCollectionIntegration(Integration):
    _collection_name: str
    _dag: DAG

    def __init__(self, dag: DAG, metadata: IntegrationInfo, collection_name: str) -> None:
        self._metadata = metadata
        self._dag = dag
        self._collection_name = collection_name

    @validate_is_connected()
    def find(
        self,
        *args: List[Any],
        name: Optional[str] = None,
        output: Optional[str] = None,
        description: str = "",
        parameters: Optional[List[BaseArtifact]] = None,
        lazy: bool = False,
        **kwargs: Dict[str, Any],
    ) -> BaseArtifact:
        """
        `find` accepts almost exactly the same input signature as the `find` exposed by mongo:
        https://www.mongodb.com/docs/manual/tutorial/query-documents/ .

        Under the hood, we call mongo SDK's `find` API to extract from DB, using arguments you
        provided to this function.

        You can additionally provide the following keyword arguments:
            name:
                Name of the query.
            description:
                Description of the query.
            output:
                Name to assign the output artifact. If not set, the default naming scheme will be used.
            parameters:
                An optional list of string parameters to use in the query.  We use the Postgres syntax of $1, $2 for placeholders.
                The number denotes which parameter in the list to use (one-indexed). These parameters feed into the
                sql query operator and will fill in the placeholders in the query with the actual values.

                Example:
                    country1 = client.create_param("UK", default=" United Kingdom ")
                    country2 = client.create_param("Thailand", default=" Thailand ")
                    mongo_db_integration.collection("hotel_reviews").find(
                        {
                            "reviewer_nationality": {
                                "$in": [$1, $2],
                           }
                        },
                        parameters=[country1, country2],
                    )

                    The query will then be executed with:
                        "reviewer_nationality": {
                            "$in": [" United Kingdom ", " Thailand "],
                       }


            lazy:
                Whether to run this operator lazily. See https://docs.aqueducthq.com/operators/lazy-vs.-eager-execution .
        """
        op_name = name or "%s query" % self.name()
        artifact_name = output or default_artifact_name_from_op_name(op_name)

        if globals.__GLOBAL_CONFIG__.lazy:
            lazy = True
        execution_mode = ExecutionMode.EAGER if not lazy else ExecutionMode.LAZY

        try:
            serialized_args = json.dumps(
                {
                    "args": args or [],
                    "kwargs": kwargs or {},
                }
            )
        except Exception as e:
            raise Exception(
                f"Cannot serialize arguments for `find`."
                "Please refer to "
                "https://www.mongodb.com/docs/manual/tutorial/query-documents/ "
                "to pass proper parameters to your query."
            ) from e

        mongo_extract_params = MongoExtractParams(
            collection=self._collection_name, query_serialized=serialized_args
        )

        # Perform validations on any parameters.
        param_artf_ids = []
        if parameters is not None:
            _validate_parameters(queries=[serialized_args], parameters=parameters)
            param_artf_ids = [artf.id() for artf in parameters]

        op_id = generate_uuid()
        output_artf_id = generate_uuid()
        apply_deltas_to_dag(
            self._dag,
            deltas=[
                AddOperatorDelta(
                    op=Operator(
                        id=op_id,
                        name=op_name,
                        description=description,
                        spec=OperatorSpec(
                            extract=ExtractSpec(
                                service=self.type(),
                                integration_id=self.id(),
                                parameters=mongo_extract_params,
                            )
                        ),
                        inputs=param_artf_ids,
                        outputs=[output_artf_id],
                    ),
                    output_artifacts=[
                        ArtifactMetadata(
                            id=output_artf_id,
                            name=sanitize_artifact_name(artifact_name),
                            type=ArtifactType.TABLE,
                            explicitly_named=output is not None,
                        ),
                    ],
                ),
            ],
        )

        if execution_mode == ExecutionMode.EAGER:
            # Issue preview request since this is an eager execution.
            artifact = artifact_utils.preview_artifact(self._dag, output_artf_id)
            assert isinstance(artifact, TableArtifact)
            return artifact
        else:
            # We are in lazy mode.
            return TableArtifact(self._dag, output_artf_id)

    @validate_is_connected()
    def save(self, artifact: BaseArtifact, update_mode: LoadUpdateMode) -> None:
        """Registers a save operator of the given artifact, to be executed when it's computed in a published flow.

        Args:
            artifact:
                The artifact to save into this collection.
            update_mode:
                Defines the semantics of the save if a table already exists.
                Options are "replace", "append" (row-wise), or "fail" (if table already exists).
        """
        _save_artifact(
            artifact.id(),
            self._dag,
            self._metadata,
            save_params=RelationalDBLoadParams(
                table=self._collection_name, update_mode=update_mode
            ),
        )


class MongoDBIntegration(Integration):
    """
    Class for MongoDB integration. This works similar to mongo's `Database` object:

    mongo_integration = client.integration("my_integration_name")
    my_table_artifact = mongo_integration.collection("my_collection").find({})
    """

    def __init__(self, dag: DAG, metadata: IntegrationInfo):
        self._dag = dag
        self._metadata = metadata

    @validate_is_connected()
    def collection(self, name: str) -> MongoDBCollectionIntegration:
        """Returns a specific collection object to call `.find()` method.

        Example:

        mongo_integration = client.integration("my_integration_name")
        my_table_artifact = mongo_integration.collection("my_collection").find({})
        """
        return MongoDBCollectionIntegration(self._dag, self._metadata, name)

    def describe(self) -> None:
        """Prints out a human-readable description of the MongoDB integration."""
        print("==================== MongoDB Integration  =============================")
        self._metadata.describe()

    @validate_is_connected()
    def save(self, artifact: BaseArtifact, collection: str, update_mode: LoadUpdateMode) -> None:
        """Registers a save operator of the given artifact, to be executed when it's computed in a published flow.

        Args:
            artifact:
                The artifact to save into the given collection.
            collection:
                The name of the collection to save to.
            update_mode:
                Defines the semantics of the save if a collection already exists.
                Options are "replace", "append" (row-wise), or "fail" (if table already exists).
        """
        _save_artifact(
            artifact.id(),
            self._dag,
            self._metadata,
            save_params=RelationalDBLoadParams(table=collection, update_mode=update_mode),
        )

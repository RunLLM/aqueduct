import json
from typing import Any, Dict, List, Optional

from aqueduct.artifacts import preview as artifact_utils
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.artifacts.save import save_artifact
from aqueduct.artifacts.table_artifact import TableArtifact
from aqueduct.constants.enums import ArtifactType, ExecutionMode, LoadUpdateMode
from aqueduct.error import InvalidUserArgumentException
from aqueduct.integrations.sql_integration import find_parameter_artifacts, find_parameter_names
from aqueduct.logger import logger
from aqueduct.models.artifact import ArtifactMetadata
from aqueduct.models.dag import DAG
from aqueduct.models.integration import Integration, IntegrationInfo
from aqueduct.models.operators import (
    ExtractSpec,
    MongoExtractParams,
    Operator,
    OperatorSpec,
    RelationalDBLoadParams,
    SaveConfig,
)
from aqueduct.utils.dag_deltas import AddOrReplaceOperatorDelta, apply_deltas_to_dag
from aqueduct.utils.utils import artifact_name_from_op_name, generate_uuid

from aqueduct import globals


class MongoDBCollectionIntegration(Integration):
    _collection_name: str
    _dag: DAG

    def __init__(self, dag: DAG, metadata: IntegrationInfo, collection_name: str) -> None:
        self._metadata = metadata
        self._dag = dag
        self._collection_name = collection_name

    def find(
        self,
        *args: List[Any],
        name: Optional[str] = None,
        description: str = "",
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
            lazy:
                Whether to run this operator lazily. See https://docs.aqueducthq.com/operators/lazy-vs.-eager-execution .
        """
        op_name = name or self._dag.get_unclaimed_op_name(prefix="%s query" % self._metadata.name)
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
        param_names = find_parameter_names(serialized_args)
        param_artifacts = find_parameter_artifacts(self._dag, param_names)
        for artf in param_artifacts:
            if artf.type != ArtifactType.STRING:
                raise InvalidUserArgumentException(
                    "The parameter `%s` must be defined as a string. Instead, got type %s"
                    % (artf.name, artf.type)
                )
        param_artf_ids = [artf.id for artf in param_artifacts]
        op_id = generate_uuid()
        output_artf_id = generate_uuid()
        apply_deltas_to_dag(
            self._dag,
            deltas=[
                AddOrReplaceOperatorDelta(
                    op=Operator(
                        id=op_id,
                        name=op_name,
                        description=description,
                        spec=OperatorSpec(
                            extract=ExtractSpec(
                                service=self._metadata.service,
                                integration_id=self._metadata.id,
                                parameters=mongo_extract_params,
                            )
                        ),
                        inputs=param_artf_ids,
                        outputs=[output_artf_id],
                    ),
                    output_artifacts=[
                        ArtifactMetadata(
                            id=output_artf_id,
                            name=artifact_name_from_op_name(op_name),
                            type=ArtifactType.TABLE,
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

    def config(self, update_mode: LoadUpdateMode) -> SaveConfig:
        """TODO(ENG-2035): Deprecated and will be removed."""
        logger().warning(
            "`integration.config()` is deprecated. Please use `integration.save()` directly instead."
        )
        return SaveConfig(
            integration_info=self._metadata,
            parameters=RelationalDBLoadParams(table=self._collection_name, update_mode=update_mode),
        )

    def save(self, artifact: BaseArtifact, update_mode: LoadUpdateMode) -> None:
        """Registers a save operator of the given artifact, to be executed when it's computed in a published flow.

        Args:
            artifact:
                The artifact to save into this collection.
            update_mode:
                Defines the semantics of the save if a table already exists.
                Options are "replace", "append" (row-wise), or "fail" (if table already exists).
        """
        save_artifact(
            artifact.id(),
            artifact.type(),
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

    def config(self, collection: str, update_mode: LoadUpdateMode) -> SaveConfig:
        """TODO(ENG-2035): Deprecated and will be removed."""
        logger().warning(
            "`integration.config()` is deprecated. Please use `integration.save()` directly instead."
        )
        return SaveConfig(
            integration_info=self._metadata,
            parameters=RelationalDBLoadParams(table=collection, update_mode=update_mode),
        )

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
        save_artifact(
            artifact.id(),
            artifact.type(),
            self._dag,
            self._metadata,
            save_params=RelationalDBLoadParams(table=collection, update_mode=update_mode),
        )

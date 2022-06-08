from __future__ import annotations
import json

import uuid
from aqueduct.utils import get_description_for_check

from aqueduct.api_client import APIClient
from aqueduct.dag import DAG, apply_deltas_to_dag, SubgraphDAGDelta
from aqueduct.error import AqueductError

from aqueduct.generic_artifact import Artifact


class CheckArtifact(Artifact):
    """This class represents a check within the flow's DAG.

    Any `@check`-annotated python function that returns a boolean will
    return this class when that function is called called. This also
    is returned from pre-defined functions like metric.bound(...).

    Examples:
        >>> @check
        >>> def check_something(df1, metric) -> bool:
        >>> return check_result
        >>>
        >>> check_artifact = check_something(table_artifact, metric_artifact)

        The contents of the check artifact can be manifested locally:

        >>> assert check_artifact.get()
    """

    def __init__(
        self,
        api_client: APIClient,
        dag: DAG,
        artifact_id: uuid.UUID,
    ):
        self._api_client = api_client
        self._dag = dag
        self._artifact_id = artifact_id

    def get(self) -> bool:
        """Materializes a CheckArtifact into a boolean.

        Returns:
            A boolean representing whether the check passed or not.

        Raises:
            InvalidRequestError:
                An error occurred because of an issue with the user's code or inputs.
            InternalServerError:
                An unexpected error occurred in the server.
        """
        dag = apply_deltas_to_dag(
            self._dag,
            deltas=[
                SubgraphDAGDelta(
                    artifact_ids=[self._artifact_id],
                    include_load_operators=False,
                )
            ],
            make_copy=True,
        )
        preview_resp = self._api_client.preview(dag=dag)
        artifact_result = preview_resp.artifact_results[self._artifact_id]

        if artifact_result.check:
            return artifact_result.check.passed
        else:
            raise AqueductError("Unable to parse execution results.")

    def describe(self) -> None:
        """Prints out a human-readable description of the check artifact."""
        print("==================== CHECK ARTIFACT =============================")
        input_operator = self._dag.must_get_operator(with_output_artifact_id=self._artifact_id)

        general_dict = get_description_for_check(input_operator)

        # Remove because values already in `readable_dict`
        general_dict.pop("Label")
        general_dict.pop("Level")

        readable_dict = super()._describe()
        readable_dict.update(general_dict)
        readable_dict["Inputs"] = [
            self._dag.must_get_artifact(artf).name for artf in input_operator.inputs
        ]

        print(json.dumps(readable_dict, sort_keys=False, indent=4))

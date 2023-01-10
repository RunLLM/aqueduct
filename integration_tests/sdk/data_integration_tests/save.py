from aqueduct.artifacts.base_artifact import BaseArtifact

from ..shared.globals import artifact_id_to_saved_identifier


def save(integration, artifact: BaseArtifact, *args):
    """Wrapper around integration.save() that also register's the save with the test suite,
    so that `validator.check_saved_artifact()` can be performed later.
    """
    integration.save(artifact, *args)

    # The assumption across all our integration.save() methods is that the identifier
    # is always the argument immediately following the artifact.
    artifact_id_to_saved_identifier[str(artifact.id())] = args[0]

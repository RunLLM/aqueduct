from aqueduct.artifacts.base_artifact import BaseArtifact

from ..shared.globals import artifact_id_to_saved_identifier


def save(integration, artifact: BaseArtifact, *args, **kwargs):
    """Wrapper around integration.save() that also register's the save with the test suite,
    so that `validator.check_saved_artifact()` can be performed later.
    """
    assert (
        len(args) > 0
    ), "We assume the first non-keyword argument is the object identifier, so one must be supplied."
    integration.save(artifact, *args, **kwargs)

    # The assumption across all our integration.save() methods is that the identifier
    # is always the argument immediately following the artifact.
    artifact_id_to_saved_identifier[str(artifact.id())] = args[0]

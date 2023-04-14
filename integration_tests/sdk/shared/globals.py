from typing import Dict

# Toggles whether we should test deprecated code paths. The is useful for ensuring both the new and
# old code paths continue to work when the API changes, but we want to continue to ensure backwards
# compatibility for a while.
use_deprecated_code_paths = False

# Global map tracking all the artifacts we've saved in the test suite and the path that they were saved to.
artifact_id_to_saved_identifier: Dict[str, str] = {}

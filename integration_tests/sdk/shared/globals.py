import uuid
from typing import Dict

# TODO(ENG-1738): this global dictionary is only maintained because we don't have a way
#  of deleting flows by name yet. The teardown code has the flow name, but not the flow id,
#  since that is generated in the test by `publish_flow()`. Therefore, we must register every
#  flow we publish in `publish_flow_test` in this dictionary.
flow_name_to_id: Dict[str, uuid.UUID] = {}

# Toggles whether we should test deprecated code paths. The is useful for ensuring both the new and
# old code paths continue to work when the API changes, but we want to continue to ensure backwards
# compatibility for a while.
use_deprecated_code_paths = False

# Global map tracking all the artifacts we've saved in the test suite and the path that they were saved to.
artifact_id_to_saved_identifier: Dict[str, str] = {}

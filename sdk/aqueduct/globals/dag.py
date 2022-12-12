from aqueduct.models.dag import DAG, Metadata

# Initialize a module-level dag object, to be accessed and modified when the user constructs the flow.
__GLOBAL_DAG__ = DAG(metadata=Metadata())

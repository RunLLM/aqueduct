from aqueduct.constants.enums import RuntimeType, SparkRuntimeType


def test_runtimetype():
    """Make sure SparkRuntimeType is a subset of RuntimeType."""
    assert set(SparkRuntimeType).issubset(set(RuntimeType))

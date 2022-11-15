from aqueduct.config import EngineConfig
from aqueduct.enums import OperatorType, RuntimeType
from aqueduct.error import InvalidUserArgumentException
from aqueduct.operators import ResourceConfig
from aqueduct.tests.utils import (
    _construct_dag,
    _construct_operator,
    default_artifact,
    default_function_spec,
    generate_uuids,
)


def test_list_downstream_operators():
    extract_op_ids = generate_uuids(1)
    extract_artifact_ids = generate_uuids(1)
    fn_op_ids = generate_uuids(3)
    fn_artifact_ids = generate_uuids(3)
    load_op_ids = generate_uuids(1)

    # Basic DAG with three functions operating on the same sql artifact.
    dag = _construct_dag(
        operators=[
            _construct_operator(
                id=extract_op_ids[0],
                name="Extract 0",
                operator_type=OperatorType.EXTRACT,
                inputs=[],
                outputs=[extract_artifact_ids[0]],
            ),
            _construct_operator(
                id=fn_op_ids[0],
                name="Function 0",
                operator_type=OperatorType.FUNCTION,
                inputs=[extract_artifact_ids[0]],
                outputs=[fn_artifact_ids[0]],
            ),
            _construct_operator(
                id=fn_op_ids[1],
                name="Function 1",
                operator_type=OperatorType.FUNCTION,
                inputs=[extract_artifact_ids[0]],
                outputs=[fn_artifact_ids[1]],
            ),
            _construct_operator(
                id=fn_op_ids[2],
                name="Function 2",
                operator_type=OperatorType.FUNCTION,
                inputs=[extract_artifact_ids[0]],
                outputs=[fn_artifact_ids[2]],
            ),
            _construct_operator(
                id=load_op_ids[0],
                name="Load 0",
                operator_type=OperatorType.LOAD,
                inputs=[fn_artifact_ids[0]],
                outputs=[],
            ),
        ],
        artifacts=[
            default_artifact(id=extract_artifact_ids[0], name="Extract Artifact"),
            default_artifact(id=fn_artifact_ids[0], name="Function Artifact"),
            default_artifact(id=fn_artifact_ids[1], name="Function Artifact"),
            default_artifact(id=fn_artifact_ids[2], name="Function Artifact"),
        ],
    )

    assert set(dag.list_downstream_operators(extract_op_ids[0])) == set(
        extract_op_ids + fn_op_ids + load_op_ids
    )

    assert set(dag.list_downstream_operators(fn_op_ids[0])) == {
        fn_op_ids[0],
        load_op_ids[0],
    }

    assert set(dag.list_downstream_operators(fn_op_ids[1])) == {
        fn_op_ids[1],
    }

    assert set(dag.list_downstream_operators(load_op_ids[0])) == set(load_op_ids)


def test_set_engine_config():
    """Check that certain resource configurations are not compatible with certain engines."""
    fn_op_ids = generate_uuids(1)
    fn_artifact_ids = generate_uuids(1)

    fn_spec = default_function_spec()
    fn_spec.resources = ResourceConfig(num_cpus=10, memory_mb=200)

    dag = _construct_dag(
        operators=[
            _construct_operator(
                id=fn_op_ids[0],
                name="Function",
                operator_type=OperatorType.FUNCTION,
                inputs=[],
                outputs=[fn_artifact_ids[0]],
                spec=fn_spec,
            ),
        ],
        artifacts=[
            default_artifact(id=fn_artifact_ids[0], name="Function Artifact"),
        ],
    )

    # Can only set to K8s runtime.
    dag.set_engine_config(EngineConfig(type=RuntimeType.K8S))

    try:
        dag.set_engine_config(EngineConfig())
    except InvalidUserArgumentException as e:
        assert "not supported" in str(e)
    else:
        assert False, "Expected failure"

    try:
        dag.set_engine_config(EngineConfig(type=RuntimeType.AIRFLOW))
    except InvalidUserArgumentException as e:
        assert "not supported" in str(e)
    else:
        assert False, "Expected failure"

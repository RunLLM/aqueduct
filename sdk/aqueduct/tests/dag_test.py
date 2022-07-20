from aqueduct.enums import OperatorType
from aqueduct.tests.utils import (
    _construct_dag,
    _construct_operator,
    default_artifact,
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

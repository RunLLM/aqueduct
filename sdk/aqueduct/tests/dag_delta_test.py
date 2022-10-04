import copy
import uuid
from typing import List

from aqueduct.dag import DAG, Metadata
from aqueduct.dag_deltas import AddOrReplaceOperatorDelta, SubgraphDAGDelta, apply_deltas_to_dag
from aqueduct.enums import OperatorType
from aqueduct.tests.utils import (
    _construct_dag,
    _construct_operator,
    default_artifact,
    generate_uuids,
)
from aqueduct.utils import generate_uuid


def test_add_and_replace_operator_delta():
    extract_op_ids = generate_uuids(1)
    extract_artifact_ids = generate_uuids(1)
    fn_op_ids = generate_uuids(3)
    fn_artifact_ids = generate_uuids(3)
    load_op_ids = generate_uuids(2)

    extract_op = _construct_operator(
        id=extract_op_ids[0],
        name="Extract",
        operator_type=OperatorType.EXTRACT,
        inputs=[],
        outputs=[extract_artifact_ids[0]],
    )
    extract_artifact = default_artifact(id=extract_artifact_ids[0], name="Extract Artifact")

    fn_op_0 = _construct_operator(
        id=fn_op_ids[0],
        name="Function 0",
        operator_type=OperatorType.FUNCTION,
        inputs=[extract_artifact_ids[0]],
        outputs=[fn_artifact_ids[0]],
    )
    fn_artifact_0 = default_artifact(id=fn_artifact_ids[0], name="Function 0 Artifact")

    fn_op_1 = _construct_operator(
        id=fn_op_ids[1],
        name="Function 1",
        operator_type=OperatorType.FUNCTION,
        inputs=[fn_artifact_ids[0]],
        outputs=[fn_artifact_ids[1]],
    )
    fn_artifact_1 = default_artifact(id=fn_artifact_ids[1], name="Function 1 Artifact")

    fn_op_2 = _construct_operator(
        id=fn_op_ids[2],
        name="Function 2",
        operator_type=OperatorType.FUNCTION,
        inputs=[fn_artifact_ids[0]],
        outputs=[fn_artifact_ids[2]],
    )
    fn_artifact_2 = default_artifact(id=fn_artifact_ids[2], name="Function 2 Artifact")

    load_fn_1 = _construct_operator(
        id=load_op_ids[0],
        name="Load Function 1",
        operator_type=OperatorType.LOAD,
        inputs=[fn_artifact_ids[1]],
        outputs=[],
    )
    load_fn_2 = _construct_operator(
        id=load_op_ids[1],
        name="Load Function 2",
        operator_type=OperatorType.LOAD,
        inputs=[fn_artifact_ids[2]],
        outputs=[],
    )

    # Construct a multi-fn, branching DAG from nothing using this DAGDelta.
    dag = DAG(metadata=Metadata())
    apply_deltas_to_dag(
        dag,
        deltas=[
            AddOrReplaceOperatorDelta(extract_op, output_artifacts=[extract_artifact]),
            AddOrReplaceOperatorDelta(fn_op_0, output_artifacts=[fn_artifact_0]),
            AddOrReplaceOperatorDelta(fn_op_1, output_artifacts=[fn_artifact_1]),
            AddOrReplaceOperatorDelta(fn_op_2, output_artifacts=[fn_artifact_2]),
            AddOrReplaceOperatorDelta(load_fn_1, output_artifacts=[]),
            AddOrReplaceOperatorDelta(load_fn_2, output_artifacts=[]),
        ],
    )

    assert dag == _construct_dag(
        operators=[extract_op, fn_op_0, fn_op_1, fn_op_2, load_fn_1, load_fn_2],
        artifacts=[extract_artifact, fn_artifact_0, fn_artifact_1, fn_artifact_2],
    )

    # Try replacing Function 2.
    fn_op_2_replacement_artifact_id = generate_uuid()
    fn_op_2_replacement = _construct_operator(
        id=generate_uuid(),
        name="Function 2",
        operator_type=OperatorType.FUNCTION,
        inputs=[fn_artifact_ids[0]],
        outputs=[fn_op_2_replacement_artifact_id],
    )
    fn_artifact_3 = default_artifact(
        id=fn_op_2_replacement_artifact_id, name="Function 2 Replacement Artifact"
    )

    apply_deltas_to_dag(
        dag,
        deltas=[AddOrReplaceOperatorDelta(fn_op_2_replacement, output_artifacts=[fn_artifact_3])],
    )
    assert dag == _construct_dag(
        operators=[extract_op, fn_op_0, fn_op_1, fn_op_2_replacement, load_fn_1],
        artifacts=[extract_artifact, fn_artifact_0, fn_artifact_1, fn_artifact_3],
    )


def _check_subgraph_test_case(
    dag: DAG,
    expected_dag: DAG,
    artifact_ids: List[uuid.UUID],
    include_load_operators: bool = False,
    include_checks: bool = False,
):
    """Apply the subgraph delta onto `dag` and expect `expected_dag`."""
    computed_dag = apply_deltas_to_dag(
        dag,
        deltas=[
            SubgraphDAGDelta(
                artifact_ids=artifact_ids,
                include_saves=include_load_operators,
                include_checks=include_checks,
            ),
        ],
        make_copy=True,
    )
    assert computed_dag == expected_dag


def test_subgraph_dag_delta():
    extract_op_ids = generate_uuids(3)
    extract_artifact_ids = generate_uuids(3)
    fn_op_ids = generate_uuids(3)
    fn_artifact_ids = generate_uuids(3)
    load_op_ids = generate_uuids(3)

    # Basic DAG with two extract operators, feeding into the same single function.
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
                id=extract_op_ids[1],
                name="Extract 1",
                operator_type=OperatorType.EXTRACT,
                inputs=[],
                outputs=[extract_artifact_ids[1]],
            ),
            _construct_operator(
                id=fn_op_ids[0],
                name="Function 0",
                operator_type=OperatorType.FUNCTION,
                inputs=[extract_artifact_ids[0], extract_artifact_ids[1]],
                outputs=[fn_artifact_ids[0]],
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
            default_artifact(id=extract_artifact_ids[0], name="Extract 0 Artifact"),
            default_artifact(id=extract_artifact_ids[1], name="Extract 1 Artifact"),
            default_artifact(id=fn_artifact_ids[0], name="Function 0 Artifact"),
        ],
    )

    # Anchoring on the function artifact should return the entire dag.
    expected_dag = copy.deepcopy(dag)
    _check_subgraph_test_case(
        dag, expected_dag, artifact_ids=[fn_artifact_ids[0]], include_load_operators=True
    )

    # Check that excluding load operators works as expected.
    expected_dag.remove_operator(load_op_ids[0])
    _check_subgraph_test_case(
        dag, expected_dag, artifact_ids=[fn_artifact_ids[0]], include_load_operators=False
    )

    # Anchoring on one of the extract artifacts should only return that extract artifact.
    expected_dag = copy.deepcopy(dag)
    expected_dag.remove_operators(
        [
            load_op_ids[0],
            extract_op_ids[0],
            fn_op_ids[0],
        ]
    )
    _check_subgraph_test_case(
        dag,
        expected_dag,
        artifact_ids=[extract_artifact_ids[1]],
        include_load_operators=True,
    )

    # Add an additional distinct path through the DAG.
    dag.add_operators(
        [
            _construct_operator(
                id=extract_op_ids[2],
                name="Extract 2",
                operator_type=OperatorType.EXTRACT,
                inputs=[],
                outputs=[extract_artifact_ids[2]],
            ),
            _construct_operator(
                id=fn_op_ids[1],
                name="Function 1",
                operator_type=OperatorType.FUNCTION,
                inputs=[extract_artifact_ids[2]],
                outputs=[fn_artifact_ids[1]],
            ),
            _construct_operator(
                id=load_op_ids[1],
                name="Load 1",
                operator_type=OperatorType.LOAD,
                inputs=[fn_artifact_ids[1]],
                outputs=[],
            ),
        ]
    )
    dag.add_artifacts(
        [
            default_artifact(extract_artifact_ids[2], name="Extract 2 Artifact"),
            default_artifact(fn_artifact_ids[1], name="Function 1 Artifact"),
        ]
    )

    # Anchoring on the one of the branches will delete the other one completely.
    expected_dag = copy.deepcopy(dag)
    expected_dag.remove_operators(
        [
            extract_op_ids[2],
            fn_op_ids[1],
            load_op_ids[1],
        ]
    )
    _check_subgraph_test_case(
        dag, expected_dag, artifact_ids=[fn_artifact_ids[0]], include_load_operators=True
    )

    expected_dag.remove_operator(load_op_ids[0])
    _check_subgraph_test_case(
        dag, expected_dag, artifact_ids=[fn_artifact_ids[0]], include_load_operators=False
    )

    # Anchoring at the end of both branches will preserve the entire DAG.
    expected_dag = copy.deepcopy(dag)
    _check_subgraph_test_case(
        dag,
        expected_dag,
        artifact_ids=[fn_artifact_ids[0], fn_artifact_ids[1]],
        include_load_operators=True,
    )

    expected_dag.remove_operators(load_op_ids[:2])
    _check_subgraph_test_case(
        dag,
        expected_dag,
        artifact_ids=[fn_artifact_ids[0], fn_artifact_ids[1]],
        include_load_operators=False,
    )

    # Check that any non-terminal artifacts in the subgraph are handled
    _check_subgraph_test_case(
        dag,
        expected_dag,
        artifact_ids=[
            fn_artifact_ids[0],
            fn_artifact_ids[1],
            extract_artifact_ids[0],
            extract_artifact_ids[2],
        ],
        include_load_operators=False,
    )


def test_subgraph_delta_with_checks():
    """Tests the `include_check_operators` case."""
    extract_op_ids = generate_uuids(2)
    extract_artifact_ids = generate_uuids(2)
    fn_op_ids = generate_uuids(2)
    fn_artifact_ids = generate_uuids(2)

    # There are three checks, one on a terminal, one on an intermediate node, and one with dependencies
    # outside of the originally specified subgraph.
    check_op_ids = generate_uuids(3)
    check_artifact_ids = generate_uuids(3)

    # Basic DAG with two parallel extract and function patterns, with the various checks attached.
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
                id=extract_op_ids[1],
                name="Extract 1",
                operator_type=OperatorType.EXTRACT,
                inputs=[],
                outputs=[extract_artifact_ids[1]],
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
                inputs=[extract_artifact_ids[1]],
                outputs=[fn_artifact_ids[1]],
            ),
            _construct_operator(
                id=check_op_ids[0],
                name="Check on terminal node",
                operator_type=OperatorType.CHECK,
                inputs=[fn_artifact_ids[0]],
                outputs=[check_artifact_ids[0]],
            ),
            _construct_operator(
                id=check_op_ids[1],
                name="Check on intermediate node",
                operator_type=OperatorType.CHECK,
                inputs=[extract_artifact_ids[0]],
                outputs=[check_artifact_ids[1]],
            ),
            _construct_operator(
                id=check_op_ids[2],
                name="Check with multiple dependencies",
                operator_type=OperatorType.CHECK,
                inputs=[extract_artifact_ids[0], fn_artifact_ids[1]],
                outputs=[check_artifact_ids[2]],
            ),
        ],
        artifacts=[
            default_artifact(id=extract_artifact_ids[0], name="Extract 0 Artifact"),
            default_artifact(id=extract_artifact_ids[1], name="Extract 1 Artifact"),
            default_artifact(id=fn_artifact_ids[0], name="Function 0 Artifact"),
            default_artifact(id=fn_artifact_ids[1], name="Function 1 Artifact"),
            default_artifact(id=check_artifact_ids[0], name="Check on terminal Artifact"),
            default_artifact(id=check_artifact_ids[1], name="Check on intermediate Artifact"),
            default_artifact(id=check_artifact_ids[2], name="Check with multiple deps Artifact"),
        ],
    )

    # The entire tree will be subgraphed because the check operator with multiple dependencies
    # pulls in the other half of the dag.
    expected_dag = copy.deepcopy(dag)
    _check_subgraph_test_case(
        dag,
        expected_dag,
        artifact_ids=[fn_artifact_ids[0]],
        include_checks=True,
    )

    # If the multiple dependency check is requested explicitly, there is one upstream function operator
    # and its check that won't be included.
    expected_dag.remove_operator(fn_op_ids[0])
    expected_dag.remove_operator(check_op_ids[0])
    _check_subgraph_test_case(
        dag,
        expected_dag,
        artifact_ids=[check_artifact_ids[2]],
        include_checks=True,
    )

    # None of the copies are included if include_check_artifact=False.
    expected_dag = copy.deepcopy(dag)
    for check_op_id in check_op_ids:
        expected_dag.remove_operator(check_op_id)
    expected_dag.remove_operator(extract_op_ids[1])
    expected_dag.remove_operator(fn_op_ids[1])
    _check_subgraph_test_case(
        dag,
        expected_dag,
        artifact_ids=[fn_artifact_ids[0]],
        include_checks=False,
    )


def test_apply_deltas_make_copy():
    extract_op_id = generate_uuid()
    extract_artifact_id = generate_uuid()
    fn_op_id = generate_uuid()
    fn_artifact_id = generate_uuid()
    load_op_id = generate_uuid()

    dag = _construct_dag(
        operators=[
            _construct_operator(
                id=extract_op_id,
                name="Extract",
                operator_type=OperatorType.EXTRACT,
                inputs=[],
                outputs=[extract_artifact_id],
            ),
            _construct_operator(
                id=fn_op_id,
                name="Function",
                operator_type=OperatorType.FUNCTION,
                inputs=[extract_artifact_id],
                outputs=[fn_artifact_id],
            ),
            _construct_operator(
                id=load_op_id,
                name="Load",
                operator_type=OperatorType.LOAD,
                inputs=[fn_artifact_id],
                outputs=[],
            ),
        ],
        artifacts=[
            default_artifact(id=extract_artifact_id, name="Extract Artifact"),
            default_artifact(id=fn_artifact_id, name="Function Artifact"),
        ],
    )

    # Check that the original DAG is not modified when make_copy=True
    computed_dag = apply_deltas_to_dag(
        dag,
        deltas=[
            SubgraphDAGDelta(
                artifact_ids=[fn_artifact_id],
                include_saves=False,
            ),
        ],
        make_copy=True,
    )
    assert computed_dag != dag

    # Check that the original DAG is modified when make_copy=False
    computed_dag = apply_deltas_to_dag(
        dag,
        deltas=[
            SubgraphDAGDelta(
                artifact_ids=[fn_artifact_id],
                include_saves=False,
            ),
        ],
        make_copy=False,
    )
    assert computed_dag == dag


def test_metrics_subgraph_dag_delta():
    extract_n = 2
    extract_op_ids = generate_uuids(extract_n)
    extract_artifact_ids = generate_uuids(extract_n)
    fn_n = 10
    fn_op_ids = generate_uuids(fn_n)
    fn_artifact_ids = generate_uuids(fn_n)
    metric_n = 2
    metric_op_ids = generate_uuids(metric_n)
    metric_artifact_ids = generate_uuids(metric_n)

    extract_artifacts = [
        default_artifact(id=extract_artifact_ids[i], name=f"Extract {i} Artifact")
        for i in range(extract_n)
    ]
    fn_artifacts = [
        default_artifact(id=fn_artifact_ids[i], name=f"Function {i} Artifact") for i in range(fn_n)
    ]
    metric_artifacts = [
        default_artifact(id=metric_artifact_ids[i], name=f"Metric {i} Artifact")
        for i in range(metric_n)
    ]

    extract_ops = [
        _construct_operator(
            id=extract_op_ids[i],
            name=f"Extract {i}",
            operator_type=OperatorType.EXTRACT,
            inputs=[],
            outputs=[extract_artifact_ids[i]],
        )
        for i in range(extract_n)
    ]
    fn_ops = [
        _construct_operator(
            id=fn_op_ids[i],
            name=f"Function {i}",
            operator_type=OperatorType.FUNCTION,
            inputs=[],
            outputs=[fn_artifact_ids[i]],
        )
        for i in range(fn_n)
    ]
    metric_ops = [
        _construct_operator(
            id=metric_op_ids[i],
            name=f"Metric {i}",
            operator_type=OperatorType.METRIC,
            inputs=[],
            outputs=[metric_artifact_ids[i]],
        )
        for i in range(metric_n)
    ]

    # Fill out the inputs
    fn_ops[0].inputs.append(extract_artifact_ids[0])
    fn_ops[3].inputs.append(extract_artifact_ids[0])
    fn_ops[5].inputs.append(extract_artifact_ids[0])
    fn_ops[8].inputs.append(extract_artifact_ids[1])
    fn_ops[1].inputs.append(fn_artifact_ids[0])
    fn_ops[2].inputs.append(fn_artifact_ids[1])
    fn_ops[4].inputs.append(fn_artifact_ids[3])
    fn_ops[7].inputs.append(fn_artifact_ids[4])
    fn_ops[6].inputs.append(fn_artifact_ids[5])
    fn_ops[7].inputs.append(fn_artifact_ids[6])
    metric_ops[0].inputs.append(fn_artifact_ids[7])
    fn_ops[9].inputs.append(fn_artifact_ids[8])
    metric_ops[1].inputs.append(fn_artifact_ids[8])

    # DAG
    # - 1 subgraph with metric at a terminal node
    #   with 2 branch feeding into it
    #   and 1 branch unrelated to it
    # - 1 linear subgraph with metric not at the terminal node
    #
    # e0 - f0 - f1 - f2
    #    \ f3 - f4 \
    #    \ f5 - f6 - f7 - m0
    #
    # e1 - f8 - f9
    #         \ m1
    dag = _construct_dag(
        operators=extract_ops + fn_ops + metric_ops,
        artifacts=extract_artifacts + fn_artifacts + metric_artifacts,
    )

    metric_check = {
        0: {
            "extract": [0],
            "fn": [3, 4, 5, 6, 7],
        },
        1: {
            "extract": [1],
            "fn": [8],
        },
    }

    check_artifacts = {
        str(metric_artifact_ids[i]): set(
            [str(metric_artifact_ids[i])]
            + [str(extract_artifact_ids[j]) for j in metric_check[i]["extract"]]
            + [str(fn_artifact_ids[j]) for j in metric_check[i]["fn"]]
        )
        for i in metric_check.keys()
    }

    check_ops = {
        str(metric_artifact_ids[i]): set(
            [str(metric_op_ids[i])]
            + [str(extract_op_ids[j]) for j in metric_check[i]["extract"]]
            + [str(fn_op_ids[j]) for j in metric_check[i]["fn"]]
        )
        for i in metric_check.keys()
    }

    for i in range(metric_n):
        sub_dag = apply_deltas_to_dag(
            dag,
            [SubgraphDAGDelta(artifact_ids=[metric_artifact_ids[i]])],
            make_copy=True,
        )
        correct_artifacts = check_artifacts[str(metric_artifact_ids[i])]
        actual_artifacts = set([str(artifact.id) for artifact in sub_dag.list_artifacts()])
        assert actual_artifacts == correct_artifacts

        correct_ops = check_ops[str(metric_artifact_ids[i])]
        actual_ops = set([str(op.id) for op in sub_dag.list_operators()])
        assert actual_ops == correct_ops

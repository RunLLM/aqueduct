package dag

import (
	"testing"

	"github.com/aqueducthq/aqueduct/lib/collections/artifact"
	"github.com/aqueducthq/aqueduct/lib/collections/operator"
	"github.com/aqueducthq/aqueduct/lib/collections/workflow_dag"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// This manually creates a DAG as follows:
// extract_0 -> artifact_0 --
// 						    |
//							v
//						    |--> func_0 -> artifact_2 -> load_0
//						    ^
//						    |
// extract_1 -> artifact_1 --
func generateBasicDag(t *testing.T) *workflow_dag.WorkflowDag {
	artifactZero := artifact.DBArtifact{
		Id: uuid.New(),
	}

	artifactOne := artifact.DBArtifact{
		Id: uuid.New(),
	}

	artifactTwo := artifact.DBArtifact{
		Id: uuid.New(),
	}

	extractZero := operator.Operator{
		Id:      uuid.New(),
		Outputs: []uuid.UUID{artifactZero.Id},
	}

	extractOne := operator.Operator{
		Id:      uuid.New(),
		Outputs: []uuid.UUID{artifactOne.Id},
	}

	functionZero := operator.Operator{
		Id:      uuid.New(),
		Inputs:  []uuid.UUID{artifactZero.Id, artifactOne.Id},
		Outputs: []uuid.UUID{artifactTwo.Id},
	}

	loadZero := operator.Operator{
		Id:     uuid.New(),
		Inputs: []uuid.UUID{artifactTwo.Id},
	}

	return &workflow_dag.WorkflowDag{
		Operators: map[uuid.UUID]operator.Operator{
			extractZero.Id:  extractZero,
			extractOne.Id:   extractOne,
			functionZero.Id: functionZero,
			loadZero.Id:     loadZero,
		},
		Artifacts: map[uuid.UUID]artifact.DBArtifact{
			artifactZero.Id: artifactZero,
			artifactOne.Id:  artifactOne,
			artifactTwo.Id:  artifactTwo,
		},
	}
}

// This manually creates a cyclic DAG as follows:
// extract_0 -> artifact_0 --
// 						    |
//							v
//							|--> func_0 -> artifact_2 -> load_0
//							^				|
//							|				|-> extract_0 // cyclic
//							|
// extract_1 -> artifact_1 --
func generateCyclicDag(t *testing.T) *workflow_dag.WorkflowDag {
	artifactZero := artifact.DBArtifact{
		Id: uuid.New(),
	}

	artifactOne := artifact.DBArtifact{
		Id: uuid.New(),
	}

	artifactTwo := artifact.DBArtifact{
		Id: uuid.New(),
	}

	extractZero := operator.Operator{
		Id:      uuid.New(),
		Inputs:  []uuid.UUID{artifactTwo.Id},
		Outputs: []uuid.UUID{artifactZero.Id},
	}

	extractOne := operator.Operator{
		Id:      uuid.New(),
		Outputs: []uuid.UUID{artifactOne.Id},
	}

	functionZero := operator.Operator{
		Id:      uuid.New(),
		Inputs:  []uuid.UUID{artifactZero.Id, artifactOne.Id},
		Outputs: []uuid.UUID{artifactTwo.Id},
	}

	loadZero := operator.Operator{
		Id:     uuid.New(),
		Inputs: []uuid.UUID{artifactTwo.Id},
	}

	return &workflow_dag.WorkflowDag{
		Operators: map[uuid.UUID]operator.Operator{
			extractZero.Id:  extractZero,
			extractOne.Id:   extractOne,
			functionZero.Id: functionZero,
			loadZero.Id:     loadZero,
		},
		Artifacts: map[uuid.UUID]artifact.DBArtifact{
			artifactZero.Id: artifactZero,
			artifactOne.Id:  artifactOne,
			artifactTwo.Id:  artifactTwo,
		},
	}
}

// This manually creates a DAG with an operator whose dependency is never going to be met:
// artifact_0 -> validation_0
func generateUnexecutableOperatorDag(t *testing.T) *workflow_dag.WorkflowDag {
	validationOpId := uuid.New()
	artifactId := uuid.New()

	artifactObject := artifact.DBArtifact{
		Id: artifactId,
	}

	validationOperator := operator.Operator{
		Id:     validationOpId,
		Inputs: []uuid.UUID{artifactObject.Id},
	}

	return &workflow_dag.WorkflowDag{
		Operators: map[uuid.UUID]operator.Operator{validationOpId: validationOperator},
		Artifacts: map[uuid.UUID]artifact.DBArtifact{artifactId: artifactObject},
	}
}

// This manually creates a DAG with no operator.
func generateEmptyDag(t *testing.T) *workflow_dag.WorkflowDag {
	return &workflow_dag.WorkflowDag{
		Operators: map[uuid.UUID]operator.Operator{},
		Artifacts: map[uuid.UUID]artifact.DBArtifact{},
	}
}

// This manually creates a DAG with an unreachable artifract:
// operator_0 -> artifact_0, artifact_1
func generateUnreachableArtifactDag(t *testing.T) *workflow_dag.WorkflowDag {
	artifactZero := artifact.DBArtifact{
		Id: uuid.New(),
	}

	artifactOne := artifact.DBArtifact{
		Id: uuid.New(),
	}

	operatorZero := operator.Operator{
		Id:      uuid.New(),
		Outputs: []uuid.UUID{artifactZero.Id},
	}

	return &workflow_dag.WorkflowDag{
		Operators: map[uuid.UUID]operator.Operator{operatorZero.Id: operatorZero},
		Artifacts: map[uuid.UUID]artifact.DBArtifact{artifactZero.Id: artifactZero, artifactOne.Id: artifactOne},
	}
}

// This manually creates a DAG with an edge that contains an undefined artifact:
// operator_0 -> artifact_0, artifact_0 not included in `dags.Artifacts`
func generateUndefinedArtifactDag(t *testing.T) *workflow_dag.WorkflowDag {
	artifactId := uuid.New()

	operatorZero := operator.Operator{
		Id:      uuid.New(),
		Outputs: []uuid.UUID{artifactId},
	}

	return &workflow_dag.WorkflowDag{
		Operators: map[uuid.UUID]operator.Operator{operatorZero.Id: operatorZero},
		Artifacts: map[uuid.UUID]artifact.DBArtifact{},
	}
}

func TestValidate(t *testing.T) {
	basicDag := generateBasicDag(t)
	err := Validate(
		basicDag,
	)
	require.Nil(t, err)

	cyclicDag := generateCyclicDag(t)
	err = Validate(
		cyclicDag,
	)
	require.Equal(t, err, ErrUnexecutableOperator)

	unExecutableOperatorDag := generateUnexecutableOperatorDag(t)
	err = Validate(
		unExecutableOperatorDag,
	)
	require.Equal(t, err, ErrUnexecutableOperator)

	emptyDag := generateEmptyDag(t)
	err = Validate(
		emptyDag,
	)
	require.Equal(t, err, ErrNoOperator)

	unreachableArtifactDag := generateUnreachableArtifactDag(t)
	err = Validate(
		unreachableArtifactDag,
	)
	require.Equal(t, err, ErrUnreachableArtifact)

	undefinedArtifactDag := generateUndefinedArtifactDag(t)
	err = Validate(
		undefinedArtifactDag,
	)
	require.Equal(t, err, ErrUnDefinedArtifact)
}

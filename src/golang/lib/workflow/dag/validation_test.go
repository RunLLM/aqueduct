package dag

import (
	"testing"

	"github.com/aqueducthq/aqueduct/lib/models"
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
func generateBasicDag(t *testing.T) *models.DAG {
	artifactZero := models.Artifact{
		ID: uuid.New(),
	}

	artifactOne := models.Artifact{
		ID: uuid.New(),
	}

	artifactTwo := models.Artifact{
		ID: uuid.New(),
	}

	extractZero := models.Operator{
		ID:      uuid.New(),
		Outputs: []uuid.UUID{artifactZero.ID},
	}

	extractOne := models.Operator{
		ID:      uuid.New(),
		Outputs: []uuid.UUID{artifactOne.ID},
	}

	functionZero := models.Operator{
		ID:      uuid.New(),
		Inputs:  []uuid.UUID{artifactZero.ID, artifactOne.ID},
		Outputs: []uuid.UUID{artifactTwo.ID},
	}

	loadZero := models.Operator{
		ID:     uuid.New(),
		Inputs: []uuid.UUID{artifactTwo.ID},
	}

	return &models.DAG{
		Operators: map[uuid.UUID]models.Operator{
			extractZero.ID:  extractZero,
			extractOne.ID:   extractOne,
			functionZero.ID: functionZero,
			loadZero.ID:     loadZero,
		},
		Artifacts: map[uuid.UUID]models.Artifact{
			artifactZero.ID: artifactZero,
			artifactOne.ID:  artifactOne,
			artifactTwo.ID:  artifactTwo,
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
func generateCyclicDag(t *testing.T) *models.DAG {
	artifactZero := models.Artifact{
		ID: uuid.New(),
	}

	artifactOne := models.Artifact{
		ID: uuid.New(),
	}

	artifactTwo := models.Artifact{
		ID: uuid.New(),
	}

	extractZero := models.Operator{
		ID:      uuid.New(),
		Inputs:  []uuid.UUID{artifactTwo.ID},
		Outputs: []uuid.UUID{artifactZero.ID},
	}

	extractOne := models.Operator{
		ID:      uuid.New(),
		Outputs: []uuid.UUID{artifactOne.ID},
	}

	functionZero := models.Operator{
		ID:      uuid.New(),
		Inputs:  []uuid.UUID{artifactZero.ID, artifactOne.ID},
		Outputs: []uuid.UUID{artifactTwo.ID},
	}

	loadZero := models.Operator{
		ID:     uuid.New(),
		Inputs: []uuid.UUID{artifactTwo.ID},
	}

	return &models.DAG{
		Operators: map[uuid.UUID]models.Operator{
			extractZero.ID:  extractZero,
			extractOne.ID:   extractOne,
			functionZero.ID: functionZero,
			loadZero.ID:     loadZero,
		},
		Artifacts: map[uuid.UUID]models.Artifact{
			artifactZero.ID: artifactZero,
			artifactOne.ID:  artifactOne,
			artifactTwo.ID:  artifactTwo,
		},
	}
}

// This manually creates a DAG with an operator whose dependency is never going to be met:
// artifact_0 -> validation_0
func generateUnexecutableOperatorDag(t *testing.T) *models.DAG {
	validationOpId := uuid.New()
	artifactID := uuid.New()

	artifactObject := models.Artifact{
		ID: artifactID,
	}

	validationOperator := models.Operator{
		ID:     validationOpId,
		Inputs: []uuid.UUID{artifactObject.ID},
	}

	return &models.DAG{
		Operators: map[uuid.UUID]models.Operator{validationOpId: validationOperator},
		Artifacts: map[uuid.UUID]models.Artifact{artifactID: artifactObject},
	}
}

// This manually creates a DAG with no operator.
func generateEmptyDag(t *testing.T) *models.DAG {
	return &models.DAG{
		Operators: map[uuid.UUID]models.Operator{},
		Artifacts: map[uuid.UUID]models.Artifact{},
	}
}

// This manually creates a DAG with an unreachable artifract:
// operator_0 -> artifact_0, artifact_1
func generateUnreachableArtifactDag(t *testing.T) *models.DAG {
	artifactZero := models.Artifact{
		ID: uuid.New(),
	}

	artifactOne := models.Artifact{
		ID: uuid.New(),
	}

	operatorZero := models.Operator{
		ID:      uuid.New(),
		Outputs: []uuid.UUID{artifactZero.ID},
	}

	return &models.DAG{
		Operators: map[uuid.UUID]models.Operator{operatorZero.ID: operatorZero},
		Artifacts: map[uuid.UUID]models.Artifact{artifactZero.ID: artifactZero, artifactOne.ID: artifactOne},
	}
}

// This manually creates a DAG with an edge that contains an undefined artifact:
// operator_0 -> artifact_0, artifact_0 not included in `dags.Artifacts`
func generateUndefinedArtifactDag(t *testing.T) *models.DAG {
	artifactID := uuid.New()

	operatorZero := models.Operator{
		ID:      uuid.New(),
		Outputs: []uuid.UUID{artifactID},
	}

	return &models.DAG{
		Operators: map[uuid.UUID]models.Operator{operatorZero.ID: operatorZero},
		Artifacts: map[uuid.UUID]models.Artifact{},
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

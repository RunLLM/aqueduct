package dag

import (
	"context"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

var (
	ErrNoOperator              = errors.New("The DAG does not contain any operator.")
	ErrMultipleArtifactParents = errors.New("The DAG contains an artifact that's generated by multiple operators.")
	ErrUnreachableArtifact     = errors.New("The DAG has an unreachable artifact")
	ErrUnDefinedArtifact       = errors.New("The DAG's operator edge contains an undefined artifact.")
	ErrUnexecutableOperator    = errors.New("The DAG contains an operator whose dependencies will never be met.")

	ValidationErrors = map[error]bool{
		ErrNoOperator:              true,
		ErrMultipleArtifactParents: true,
		ErrUnreachableArtifact:     true,
		ErrUnDefinedArtifact:       true,
		ErrUnexecutableOperator:    true,
	}
)

func Validate(
	dag *models.DAG,
) error {
	if len(dag.Operators) == 0 {
		return ErrNoOperator
	}

	// In this map, the keys are all artifact IDs that appear in the
	// dag's edge definition, and the value is a boolean indicating
	// whether each artifact is defined in `dag.Artifacts`.
	artifactIdsInEdges := make(map[uuid.UUID]bool)

	// In this map, the key is an artifact id and the value is a
	// boolean indicating whether this artifact already has a
	// parent operator pointing to it.
	artifactParents := make(map[uuid.UUID]bool)

	for _, op := range dag.Operators {
		for _, inputArtifactId := range op.Inputs {
			artifactIdsInEdges[inputArtifactId] = false
		}
		for _, outputArtifactId := range op.Outputs {
			artifactIdsInEdges[outputArtifactId] = false

			if _, ok := artifactParents[outputArtifactId]; !ok {
				artifactParents[outputArtifactId] = true
			} else {
				// If the artifact already has a parent, we return an error as
				// we don't allow multiple operators pointing to a single artifact.
				return ErrMultipleArtifactParents
			}
		}
	}

	for artifactId := range dag.Artifacts {
		if _, ok := artifactIdsInEdges[artifactId]; !ok {
			// This means the dag contains an artifact that's
			// not included in any operator edges.
			return ErrUnreachableArtifact
		}

		artifactIdsInEdges[artifactId] = true
	}

	for _, appeared := range artifactIdsInEdges {
		if !appeared {
			// This means the dag's operator edges contain
			// an artifact that's not defined in `dags.Artifacts`.
			return ErrUnDefinedArtifact
		}
	}

	return checkUnexecutableOperator(dag)
}

func ValidateDagOperatorIntegrationOwnership(
	ctx context.Context,
	operators map[uuid.UUID]models.Operator,
	orgID string,
	userID uuid.UUID,
	integrationRepo repos.Integration,
	DB database.Database,
) (bool, error) {
	for _, operator := range operators {
		var integrationID uuid.UUID
		if operator.Spec.IsExtract() {
			integrationID = operator.Spec.Extract().IntegrationId
		} else if operator.Spec.IsLoad() {
			integrationID = operator.Spec.Load().IntegrationId
		} else {
			continue
		}

		ok, err := integrationRepo.ValidateOwnership(
			ctx,
			integrationID,
			orgID,
			userID,
			DB,
		)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}

	return true, nil
}

func checkUnexecutableOperator(dag *models.DAG) error {
	numOperators := len(dag.Operators)
	operatorsExecuted := make(map[uuid.UUID]bool, numOperators)
	for operatorId := range dag.Operators {
		operatorsExecuted[operatorId] = false
	}

	artifactToDownstreamOperatorIds := make(map[uuid.UUID][]uuid.UUID, len(dag.Artifacts))
	operatorDependencies := make(map[uuid.UUID]map[uuid.UUID]bool, numOperators)
	ready := make(map[uuid.UUID]bool, numOperators)
	active := make(map[uuid.UUID]bool, numOperators)

	for id, operator := range dag.Operators {
		operatorDependencies[id] = make(map[uuid.UUID]bool, len(operator.Inputs))
		for _, artifactId := range operator.Inputs {
			operatorDependencies[id][artifactId] = true
		}

		if len(operator.Inputs) == 0 {
			ready[id] = true
		}

		for _, artifactId := range operator.Inputs {
			downstreamOps, ok := artifactToDownstreamOperatorIds[artifactId]
			if !ok {
				downstreamOps = make([]uuid.UUID, 0, len(dag.Operators))
				artifactToDownstreamOperatorIds[artifactId] = downstreamOps
			}

			artifactToDownstreamOperatorIds[artifactId] = append(downstreamOps, id)
		}
	}

	for len(ready) > 0 || len(active) > 0 {
		completedIds := make([]uuid.UUID, 0, len(active))
		for id := range active {
			completedIds = append(completedIds, id)
			for _, artifactId := range dag.Operators[id].Outputs {
				if downstreampOps, ok := artifactToDownstreamOperatorIds[artifactId]; ok {
					for _, downstreamOpId := range downstreampOps {
						delete(operatorDependencies[downstreamOpId], artifactId)
						if len(operatorDependencies[downstreamOpId]) == 0 {
							ready[downstreamOpId] = true
						}
					}
				}
			}
		}

		for _, id := range completedIds {
			delete(active, id)
			operatorsExecuted[id] = true
		}

		for id := range ready {
			active[id] = true
		}

		ready = map[uuid.UUID]bool{}
	}

	for _, reachable := range operatorsExecuted {
		if !reachable {
			return ErrUnexecutableOperator
		}
	}

	return nil
}

package response

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

// This file should map exactly to
// `src/ui/common/src/handlers/responses/workflow.ts`
type Workflow struct {
	ID                   uuid.UUID                   `json:"id"`
	UserID               uuid.UUID                   `json:"user_id"`
	Name                 string                      `json:"name"`
	Description          string                      `json:"description"`
	Schedule             shared.Schedule             `json:"schedule"`
	CreatedAt            time.Time                   `json:"created_at"`
	RetentionPolicy      shared.RetentionPolicy      `json:"retention_policy"`
	NotificationSettings shared.NotificationSettings `json:"notification_settings"`
}

func NewWorkflowFromDBObject(dbWorkflow *models.Workflow) *Workflow {
	return &Workflow{
		ID:                   dbWorkflow.ID,
		UserID:               dbWorkflow.UserID,
		Name:                 dbWorkflow.Name,
		Description:          dbWorkflow.Description,
		Schedule:             dbWorkflow.Schedule,
		CreatedAt:            dbWorkflow.CreatedAt,
		RetentionPolicy:      dbWorkflow.RetentionPolicy,
		NotificationSettings: dbWorkflow.NotificationSettings,
	}
}

type DAG struct {
	ID            uuid.UUID            `json:"id"`
	WorkflowID    uuid.UUID            `json:"workflow_id"`
	CreatedAt     time.Time            `json:"created_at"`
	StorageConfig shared.StorageConfig `json:"storage_config"`
	EngineConfig  shared.EngineConfig  `json:"engine_config"`
}

func NewDAGFromDBObject(dbDAG *models.DAG) *DAG {
	return &DAG{
		ID:            dbDAG.ID,
		WorkflowID:    dbDAG.WorkflowID,
		CreatedAt:     dbDAG.CreatedAt,
		StorageConfig: dbDAG.StorageConfig,
		EngineConfig:  dbDAG.EngineConfig,
	}
}

type DAGResult struct {
	ID        uuid.UUID              `json:"id"`
	DagID     uuid.UUID              `json:"dag_id"`
	ExecState *shared.ExecutionState `json:"exec_state"`
}

func NewDAGResultFromDBObject(dbDAGResult *models.DAGResult) *DAGResult {
	var execStatePtr *shared.ExecutionState
	if !dbDAGResult.ExecState.IsNull {
		execStatePtr = &dbDAGResult.ExecState.ExecutionState
	}

	return &DAGResult{
		ID:        dbDAGResult.ID,
		DagID:     dbDAGResult.DagID,
		ExecState: execStatePtr,
	}
}

type WorkflowAndDagIDs struct {
	WorkflowID uuid.UUID `json:"id"`
	DagID      uuid.UUID `json:"dag_id"`
}

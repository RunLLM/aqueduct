package sqlite

import (
	"context"
	"fmt"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/views"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type workflowRepo struct {
	workflowReader
	workflowWriter
}

type workflowReader struct{}

type workflowWriter struct{}

func NewWorklowRepo() repos.Workflow {
	return &workflowRepo{
		workflowReader: workflowReader{},
		workflowWriter: workflowWriter{},
	}
}

func (*workflowReader) Exists(ctx context.Context, id uuid.UUID, db database.Database) (bool, error) {
	return utils.IdExistsInTable(ctx, id, models.WorkflowTable, db)
}

func (r *workflowReader) Get(ctx context.Context, id uuid.UUID, db database.Database) (*models.Workflow, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM workflow WHERE id = $1;`,
		models.WorkflowCols(),
	)
	args := []interface{}{id}

	return r.getOne(ctx, db, query, args...)
}

func (r *workflowReader) GetByOwnerAndName(ctx context.Context, ownerID uuid.UUID, name string, db database.Database) (*models.Workflow, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM workflow WHERE user_id = $1 and name = $2;`,
		models.WorkflowCols(),
	)
	args := []interface{}{ownerID, name}

	return r.getOne(ctx, db, query, args...)
}

func (r *workflowReader) GetLatestStatusesByOrg(ctx context.Context, orgID uuid.UUID, db database.Database) ([]views.LatestWorkflowStatus, error)

func (r *workflowReader) List(ctx context.Context, db database.Database) ([]models.Workflow, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM workflow;`,
		models.WorkflowCols(),
	)

	return r.get(ctx, db, query)
}

func (r *workflowReader) ValidateOrg(ctx context.Context, id uuid.UUID, orgID uuid.UUID, db database.Database) (bool, error) {
	query := `
	SELECT 
		COUNT(*) AS count 
	FROM 
		workflow INNER JOIN app_user ON workflow.user_id = app_user.id
	WHERE
		workflow.id = $1
		AND app_user.organization_id = $2;`
	args := []interface{}{id, orgID}

	var count utils.CountResult
	err := db.Query(ctx, &count, query, args...)
	if err != nil {
		return false, err
	}

	return count.Count == 1, nil
}

func (*workflowReader) get(ctx context.Context, db database.Database, query string, args ...interface{}) ([]models.Workflow, error) {
	var workflows []models.Workflow
	err := db.Query(ctx, &workflows, query, args...)
	return workflows, err
}

func (r *workflowReader) getOne(ctx context.Context, db database.Database, query string, args ...interface{}) (*models.Workflow, error) {
	workflows, err := r.get(ctx, db, query, args...)
	if err != nil {
		return nil, nil
	}

	if len(workflows) != 1 {
		return nil, errors.Newf("Expected 1 workflow but got %v", len(workflows))
	}

	return &workflows[0], nil
}

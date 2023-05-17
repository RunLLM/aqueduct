package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/aqueducthq/aqueduct/lib/database"
	"github.com/aqueducthq/aqueduct/lib/database/stmt_preparers"
	"github.com/aqueducthq/aqueduct/lib/models"
	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/repos"
	"github.com/dropbox/godropbox/errors"
	"github.com/google/uuid"
)

type resourceRepo struct {
	resourceReader
	resourceWriter
}

type resourceReader struct{}

type resourceWriter struct{}

func NewResourceRepo() repos.Resource {
	return &resourceRepo{
		resourceReader: resourceReader{},
		resourceWriter: resourceWriter{},
	}
}

func (*resourceReader) Get(ctx context.Context, ID uuid.UUID, DB database.Database) (*models.Resource, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM %s WHERE id = $1;`,
		models.ResourceCols(),
		models.ResourceTable,
	)
	args := []interface{}{ID}

	return getResource(ctx, DB, query, args...)
}

func (*resourceReader) GetBatch(ctx context.Context, IDs []uuid.UUID, DB database.Database) ([]models.Resource, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM %s WHERE id IN (%s);`,
		models.ResourceCols(),
		models.ResourceTable,
		stmt_preparers.GenerateArgsList(len(IDs), 1),
	)
	args := stmt_preparers.CastIdsListToInterfaceList(IDs)

	return getResources(ctx, DB, query, args...)
}

func (*resourceReader) GetByConfigField(ctx context.Context, fieldName string, fieldValue string, DB database.Database) ([]models.Resource, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE json_extract(config, $1) = $2;",
		models.ResourceCols(),
		models.ResourceTable,
	)

	// The full 'where' condition becomes
	// `json_extract(config, '$.field_name') = 'field_value'`
	// which matches https://www.sqlite.org/json1.html .
	// We parametrize the extracted field_name and field_value
	// to prevent injection.
	args := []interface{}{"$." + fieldName, fieldValue}

	return getResources(ctx, DB, query, args...)
}

func (*resourceReader) GetByNameAndUser(
	ctx context.Context,
	resourceName string,
	userID uuid.UUID,
	orgID string,
	DB database.Database,
) (*models.Resource, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM %s WHERE name = $1 AND organization_id = $2 AND (user_id IS NULL OR user_id = $3);`,
		models.ResourceCols(),
		models.ResourceTable,
	)
	args := []interface{}{resourceName, orgID, userID}
	return getResource(ctx, DB, query, args...)
}

func (*resourceReader) GetByOrg(ctx context.Context, orgId string, DB database.Database) ([]models.Resource, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM %s WHERE organization_id = $1 AND user_id IS NULL;`,
		models.ResourceCols(),
		models.ResourceTable,
	)
	args := []interface{}{orgId}
	return getResources(ctx, DB, query, args...)
}

func (*resourceReader) GetByServiceAndUser(ctx context.Context, service shared.Service, userID uuid.UUID, DB database.Database) ([]models.Resource, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM %s WHERE service = $1 AND user_id = $2;`,
		models.ResourceCols(),
		models.ResourceTable,
	)
	args := []interface{}{service, userID}
	return getResources(ctx, DB, query, args...)
}

func (*resourceReader) GetByUser(ctx context.Context, orgID string, userID uuid.UUID, DB database.Database) ([]models.Resource, error) {
	query := fmt.Sprintf(
		`SELECT %s FROM %s WHERE organization_id = $1 AND (user_id IS NULL OR user_id = $2);`,
		models.ResourceCols(),
		models.ResourceTable,
	)
	args := []interface{}{orgID, userID}
	return getResources(ctx, DB, query, args...)
}

func (*resourceReader) ValidateOwnership(ctx context.Context, resourceID uuid.UUID, orgID string, userID uuid.UUID, DB database.Database) (bool, error) {
	var count countResult

	query := fmt.Sprintf(
		`SELECT %s FROM %s WHERE id = $1;`,
		models.ResourceCols(),
		models.ResourceTable,
	)
	args := []interface{}{resourceID}

	resourceObject, err := getResource(ctx, DB, query, args...)
	if err != nil {
		return false, err
	}
	userOnly := shared.IsUserOnlyResource(resourceObject.Service)

	if userOnly {
		query := fmt.Sprintf(`SELECT COUNT(*) AS count FROM %s WHERE id = $1 AND user_id = $2;`, models.ResourceTable)
		err := DB.Query(ctx, &count, query, resourceID, userID)
		if err != nil {
			return false, err
		}
	} else {
		query := fmt.Sprintf(`SELECT COUNT(*) AS count FROM %s WHERE id = $1 AND organization_id = $2;`, models.ResourceTable)
		err := DB.Query(ctx, &count, query, resourceID, orgID)
		if err != nil {
			return false, err
		}
	}

	return count.Count == 1, nil
}

func (*resourceWriter) Create(
	ctx context.Context,
	orgID string,
	service shared.Service,
	name string,
	config *shared.ResourceConfig,
	DB database.Database,
) (*models.Resource, error) {
	cols := []string{
		models.ResourceID,
		models.ResourceOrgID,
		models.ResourceService,
		models.ResourceName,
		models.ResourceConfig,
		models.ResourceCreatedAt,
	}
	query := DB.PrepareInsertWithReturnAllStmt(models.ResourceTable, cols, models.ResourceCols())

	ID, err := GenerateUniqueUUID(ctx, models.ResourceTable, DB)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		ID,
		orgID,
		service,
		name,
		config,
		time.Now(),
	}
	return getResource(ctx, DB, query, args...)
}

func (*resourceWriter) CreateForUser(
	ctx context.Context,
	orgID string,
	userID uuid.UUID,
	service shared.Service,
	name string,
	config *shared.ResourceConfig,
	DB database.Database,
) (*models.Resource, error) {
	cols := []string{
		models.ResourceID,
		models.ResourceUserID,
		models.ResourceOrgID,
		models.ResourceService,
		models.ResourceName,
		models.ResourceConfig,
		models.ResourceCreatedAt,
	}
	query := DB.PrepareInsertWithReturnAllStmt(models.ResourceTable, cols, models.ResourceCols())

	ID, err := GenerateUniqueUUID(ctx, models.ResourceTable, DB)
	if err != nil {
		return nil, err
	}

	args := []interface{}{
		ID,
		userID,
		orgID,
		service,
		name,
		config,
		time.Now(),
	}
	return getResource(ctx, DB, query, args...)
}

func (*resourceWriter) Delete(ctx context.Context, ID uuid.UUID, DB database.Database) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1;`, models.ResourceTable)
	return DB.Execute(ctx, query, ID)
}

func (*resourceWriter) Update(ctx context.Context, ID uuid.UUID, changes map[string]interface{}, DB database.Database) (*models.Resource, error) {
	var resource models.Resource
	err := repos.UpdateRecordToDest(ctx, &resource, changes, models.ResourceTable, models.ResourceID, ID, models.ResourceCols(), DB)
	return &resource, err
}

func getResources(ctx context.Context, DB database.Database, query string, args ...interface{}) ([]models.Resource, error) {
	var resources []models.Resource
	err := DB.Query(ctx, &resources, query, args...)
	return resources, err
}

func getResource(ctx context.Context, DB database.Database, query string, args ...interface{}) (*models.Resource, error) {
	resources, err := getResources(ctx, DB, query, args...)
	if err != nil {
		return nil, err
	}

	if len(resources) == 0 {
		return nil, database.ErrNoRows()
	}

	if len(resources) != 1 {
		return nil, errors.Newf("Expected 1 resource but got %v", len(resources))
	}

	return &resources[0], nil
}

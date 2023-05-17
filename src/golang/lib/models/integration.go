package models

import (
	"strings"
	"time"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/aqueducthq/aqueduct/lib/models/utils"
	"github.com/google/uuid"
)

const (
	ResourceTable = "resource"

	// Resource table column names
	ResourceID        = "id"
	ResourceOrgID     = "organization_id"
	ResourceUserID    = "user_id"
	ResourceService   = "service"
	ResourceName      = "name"
	ResourceConfig    = "config"
	ResourceCreatedAt = "created_at"
)

// A Resource maps to the resource table.
type Resource struct {
	ID        uuid.UUID             `db:"id" json:"id"`
	UserID    utils.NullUUID        `db:"user_id" json:"user_id"`
	OrgID     string                `db:"organization_id"`
	Service   shared.Service        `db:"service"`
	Name      string                `db:"name"`
	Config    shared.ResourceConfig `db:"config"`
	CreatedAt time.Time             `db:"created_at"`
}

// ResourceCols returns a comma-separated string of all Resource columns.
func ResourceCols() string {
	return strings.Join(allResourceCols(), ",")
}

func allResourceCols() []string {
	return []string{
		ResourceID,
		ResourceOrgID,
		ResourceUserID,
		ResourceService,
		ResourceName,
		ResourceConfig,
		ResourceCreatedAt,
	}
}

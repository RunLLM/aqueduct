package models

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

// A Integration maps to the integration table.
type Integration struct {
	ID              uuid.UUID              `db:"id" json:"id"`
	UserId          uuid.NullUUID              `db:"user_id" json:"user_id"`
	OrganizationId string         `db:"organization_id"`
	Service        Service        `db:"service"`
	Name           string         `db:"name"`
	Config         utils.Config   `db:"config"`
	CreatedAt      time.Time      `db:"created_at"`
	Validated      bool           `db:"validated"`
}
package models

import (
	"time"

	"github.com/aqueducthq/aqueduct/lib/models/shared"
	"github.com/google/uuid"
)

// A Workflow maps to the workflow table.
type Workflow struct {
	ID              uuid.UUID              `db:"id" json:"id"`
	UserId          uuid.UUID              `db:"user_id" json:"user_id"`
	Name            string                 `db:"name" json:"name"`
	Description     string                 `db:"description" json:"description"`
	Schedule        shared.Schedule        `db:"schedule" json:"schedule"`
	CreatedAt       time.Time              `db:"created_at" json:"created_at"`
	RetentionPolicy shared.RetentionPolicy `db:"retention_policy" json:"retention_policy"`
}

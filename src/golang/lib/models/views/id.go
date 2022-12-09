package views

import "github.com/google/uuid"

// ObjectID is a wrapper around any ID
type ObjectID struct {
	ID uuid.UUID `db:"id" json:"id"`
}

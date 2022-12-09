package utils

import (
	"github.com/google/uuid"
)

// NullUUID represents a uuid.UUID that may be NULL.
type NullUUID struct {
	UUID   uuid.UUID
	IsNull bool
}

func (n *NullUUID) Scan(value interface{}) error {
	if value == nil {
		// UUID is NULL
		n.IsNull = true
		return nil
	}

	id := &uuid.UUID{}
	if err := id.Scan(value); err != nil {
		return err
	}

	n.UUID, n.IsNull = *id, false
	return nil
}

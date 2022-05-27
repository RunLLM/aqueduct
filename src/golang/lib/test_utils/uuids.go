package test_utils

import (
	"testing"

	"github.com/aqueducthq/aqueduct/lib/collections/utils"
	"github.com/google/uuid"
)

func GenerateUUIDSlice(t *testing.T, length int) []uuid.UUID {
	uuids := make([]uuid.UUID, 0, length)
	for i := 0; i < length; i++ {
		uuids = append(uuids, uuid.New())
	}
	return uuids
}

func ConvertUUIDsToNullUUIDSlice(uuids []uuid.UUID) utils.NullUUIDSlice {
	if uuids == nil {
		return utils.NullUUIDSlice{
			IsNull: true,
		}
	}
	return utils.NullUUIDSlice{
		UUIDSlice: uuids,
		IsNull:    false,
	}
}

func ConvertUUIDToNullUUID(id uuid.UUID) utils.NullUUID {
	if id == uuid.Nil {
		return utils.NullUUID{
			IsNull: true,
		}
	}
	return utils.NullUUID{
		UUID:   id,
		IsNull: false,
	}
}

package graph

import "github.com/google/uuid"

type node struct {
	id    uuid.UUID
	edges []uuid.UUID
}

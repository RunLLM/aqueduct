package graph

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestHasPath(t *testing.T) {
	type test struct {
		root    uuid.UUID
		dest    uuid.UUID
		hasPath bool
	}

	a, b, c, d := uuid.New(), uuid.New(), uuid.New(), uuid.New()

	gph := NewDirected()
	gph.AddNode(a)
	gph.AddNode(b)
	gph.AddNode(c)
	gph.AddNode(d)

	gph.AddEdge(a, b)
	gph.AddEdge(a, c)
	gph.AddEdge(b, d)

	tests := []test{
		{root: a, dest: b, hasPath: true},
		{root: a, dest: c, hasPath: true},
		{root: a, dest: d, hasPath: true},
		{root: c, dest: b, hasPath: false},
		{root: c, dest: d, hasPath: false},
		{root: c, dest: a, hasPath: false},
		{root: b, dest: d, hasPath: true},
		{root: b, dest: c, hasPath: false},
	}

	for _, tc := range tests {
		require.Equal(t, tc.hasPath, gph.HasPath(tc.root, tc.dest))
	}
}

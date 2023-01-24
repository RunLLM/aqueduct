package graph

import "github.com/google/uuid"

type Directed struct {
	nodes map[uuid.UUID]*node
}

// NewDirected returns an empty directed graph.
func NewDirected() *Directed {
	return &Directed{
		nodes: make(map[uuid.UUID]*node),
	}
}

// AddNode adds a new node. It does nothing if the node already exists.
func (g *Directed) AddNode(nodeID uuid.UUID) {
	if _, ok := g.nodes[nodeID]; !ok {
		g.nodes[nodeID] = &node{id: nodeID}
	}
}

// AddEdge adds a directed edge from node fromID to node toID.
func (g *Directed) AddEdge(fromID uuid.UUID, toID uuid.UUID) {
	g.nodes[fromID].edges = append(g.nodes[fromID].edges, toID)
}

// HasPath returns whether there is a path from root to dest.
func (g *Directed) HasPath(root uuid.UUID, dest uuid.UUID) bool {
	if root == dest {
		return true
	}

	var queue []uuid.UUID
	queue = append(queue, g.nodes[root].edges...)

	for len(queue) > 0 {
		root = queue[0]
		queue = queue[1:]

		if root == dest {
			return true
		}

		queue = append(queue, g.nodes[root].edges...)
	}

	return false
}

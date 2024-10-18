package Onyx

import (
	"github.com/dgraph-io/badger/v4"
)

// ContainsEdge checks whether an edge exists between two nodes
func (g *Graph) ContainsEdge(from string, to string, txn *badger.Txn) (bool, error) {
	// Get the edges of the 'from' node
	dstNodes, err := g.GetEdges(from, txn)
	if err != nil {
		// If the 'from' node doesn't exist, return false
		if err == badger.ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}

	// Check if the 'to' node is in the destination nodes list
	_, exists := dstNodes[to]

	return exists, nil
}

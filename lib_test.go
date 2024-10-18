package Onyx

import (
    "github.com/dgraph-io/badger/v4"
    "testing"
)

func TestContainsEdge(t *testing.T) {
    graph, err := NewGraph("", true) // In-memory graph for testing
    if err != nil {
        t.Fatalf("failed to create graph: %v", err)
    }
    defer graph.Close()

    txn := graph.DB.NewTransaction(true)
    defer txn.Discard()

    // Test when no edge exists
    exists, err := graph.ContainsEdge("A", "B", txn)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if exists {
        t.Fatalf("expected false, got true for non-existing edge")
    }

    // Add an edge from A to B
    err = graph.AddEdge("A", "B", txn)
    if err != nil {
        t.Fatalf("failed to add edge: %v", err)
    }

    // Test when edge exists
    exists, err = graph.ContainsEdge("A", "B", txn)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !exists {
        t.Fatalf("expected true, got false for existing edge")
    }

    // Test non-existent edge from A to C
    exists, err = graph.ContainsEdge("A", "C", txn)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if exists {
        t.Fatalf("expected false, got true for non-existing edge A to C")
    }

    // Commit transaction after tests
    err = txn.Commit()
    if err != nil {
        t.Fatalf("failed to commit transaction: %v", err)
    }
}

package Onyx

import (
    "os"
    "testing"
)

func TestPickRandomVertex(t *testing.T) {
    // Create graph in /tmp for testing and ensure errors are handled
    graph, err := NewGraph("/tmp/onyxsdlkjf", false)
    if err != nil {
        t.Fatalf("failed to create graph: %v", err)
    }
    defer graph.Close()
    defer os.RemoveAll("/tmp/onyxsdlkjf") // Cleanup after test

    // Add edges
    _ = graph.AddEdge("a", "b", nil)
    _ = graph.AddEdge("b", "c", nil)
    _ = graph.AddEdge("c", "d", nil)
    _ = graph.AddEdge("d", "e", nil)

    // Test PickRandomVertex
    v, err := graph.PickRandomVertex(nil)
    if err != nil {
        t.Fatalf("failed to pick random vertex: %v", err)
    }
    t.Logf("Picked random vertex: %s", v)

    // Ensure the picked vertex is valid
    validVertices := map[string]bool{"a": true, "b": true, "c": true, "d": true}
    if !validVertices[v] {
        t.Fatalf("picked vertex %s is not valid", v)
    }
}

func TestInsertAndRead(t *testing.T) {
    // Create graph in /tmp for testing and ensure errors are handled
    graph, err := NewGraph("/tmp/onyxsdlkjf", false)
    if err != nil {
        t.Fatalf("failed to create graph: %v", err)
    }
    defer graph.Close()
    defer os.RemoveAll("/tmp/onyxsdlkjf") // Cleanup after test

    // Add edges
    err = graph.AddEdge("a", "b", nil)
    if err != nil {
        t.Fatal(err)
    }
    err = graph.AddEdge("a", "c", nil)
    if err != nil {
        t.Fatal(err)
    }
    err = graph.AddEdge("a", "d", nil)
    if err != nil {
        t.Fatal(err)
    }

    // Get edges for node "a"
    dstNodes, err := graph.GetEdges("a", nil)
    if err != nil {
        t.Fatal(err)
    }

    // Check that all expected destination nodes exist
    for _, node := range []string{"b", "c", "d"} {
        if _, ok := dstNodes[node]; !ok {
            t.Fatalf("%s not in edgelist", node)
        }
    }
}

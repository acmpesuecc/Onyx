package Onyx

import "testing"

func TestPickRandomVertext(T *testing.T) {
	graph, _ := NewGraph("/tmp/onyxsdlkjf", false)
	defer graph.Close()
	_ = graph.AddEdge("a", "b", nil)
	_ = graph.AddEdge("b", "c", nil)
	_ = graph.AddEdge("c", "d", nil)
	_ = graph.AddEdge("d", "e", nil)

	v, _ := graph.PickRandomVertex()
	T.Log("==", string(v), "==")
}

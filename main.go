package Onyx

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"sync"
)

// TODO: Add Label support for edgess
type Graph struct {
	DB *badger.DB
}

func NewGraph(path string, inMemory bool) (*Graph, error) {
	var db *badger.DB
	var err error

	if inMemory {
		db, err = badger.Open(badger.DefaultOptions("").WithInMemory(true))
	} else {
		db, err = badger.Open(badger.DefaultOptions(path))
	}

	return &Graph{db}, err
}

func (g *Graph) Close() {
	g.DB.Close()
}

func (g *Graph) AddEdge(from string, to string, txn *badger.Txn) error {
	localTxn := txn == nil
	if localTxn {
		txn = g.DB.NewTransaction(true)
		defer txn.Discard()
	}

	var dstNodes map[string]bool

	item, err := txn.Get([]byte(from))
	if err != nil {
		if err == badger.ErrKeyNotFound {
			dstNodes = make(map[string]bool)
		} else {
			return err
		}
	} else { //if err==nil
		valCopy, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		dstNodes, err = deserializeEdgeMap(valCopy)
		if err != nil {
			return err
		}
	}

	dstNodes[to] = true

	serializedEdgeMap, err := serializeEdgeMap(dstNodes)
	if err != nil {
		return err
	}
	err = txn.Set([]byte(from), serializedEdgeMap)
	if err != nil {
		return err
	}

	if localTxn {
		err = txn.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Graph) RemoveEdge(from string, to string, txn *badger.Txn) error {
	localTxn := txn == nil
	if localTxn {
		txn = g.DB.NewTransaction(true)
		defer txn.Discard()
	}

	item, err := txn.Get([]byte(from))
	if err != nil {
		return err
	}

	valCopy, err := item.ValueCopy(nil)
	if err != nil {
		return err
	}

	dstNodes, err := deserializeEdgeMap(valCopy)
	if err != nil {
		return err
	}
	delete(dstNodes, to)

	serializedEdgeMap, err := serializeEdgeMap(dstNodes)
	if err != nil {
		return err
	}
	err = txn.Set([]byte(from), serializedEdgeMap)
	if err != nil {
		return err
	}

	if localTxn {
		err = txn.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Graph) GetEdges(from string, txn *badger.Txn) (map[string]bool, error) {
	localTxn := txn == nil
	if localTxn {
		txn = g.DB.NewTransaction(false)
		defer txn.Discard()
	}

	item, err := txn.Get([]byte(from))
	if err != nil {
		return nil, err
	}

	if localTxn {
		err = txn.Commit()
		if err != nil {
			return nil, err
		}
	}

	valCopy, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}

	neighbors, err := deserializeEdgeMap(valCopy)
	return neighbors, err
}

func (g *Graph) IterAllEdges(f func(iterator *badger.Item) error, prefetchSize int, txn *badger.Txn) error {
	localTxn := txn == nil
	if localTxn {
		txn = g.DB.NewTransaction(false)
		defer txn.Discard()
	}

	opts := badger.DefaultIteratorOptions
	opts.PrefetchSize = prefetchSize
	it := txn.NewIterator(opts)
	defer it.Close()
	for it.Rewind(); it.Valid(); it.Next() {
		item := it.Item()
		err := f(item)
		if err != nil {
			return err
		}
	}

	if localTxn {
		err := txn.Commit()
		if err != nil {
			return nil
		}
	}
	return nil
}

func serializeEdgeMap(m map[string]bool) ([]byte, error) {
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	err := e.Encode(m)
	return b.Bytes(), err
}

func deserializeEdgeMap(serializedMap []byte) (map[string]bool, error) {
	b := bytes.NewBuffer(serializedMap)
	d := gob.NewDecoder(b)

	deserializedMap := make(map[string]bool)
	// Decoding the serialized data
	err := d.Decode(deserializedMap)
	return deserializedMap, err
}

func main() {
	graph, err := NewGraph("", true)
	if err != nil {
		panic(err)
	}

	err = graph.AddEdge("a", "b", nil)
	err = graph.AddEdge("a", "c", nil)
	err = graph.AddEdge("c", "d", nil)
	err = graph.AddEdge("c", "e", nil)

	if err != nil {
		panic(err)
	}

	a_n, err := graph.GetEdges("a", nil)
	fmt.Println("Neighbors of a: ", a_n)

	a_n, err = graph.GetEdges("c", nil)
	fmt.Println("Neighbors of c: ", a_n)

	fmt.Println("Checking Concurrency")
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		txn1 := graph.DB.NewTransaction(true)
		defer txn1.Discard()

		graph.RemoveEdge("a", "b", txn1)
		a_n, _ = graph.GetEdges("a", txn1)
		fmt.Println("Neighbors of a: ", a_n)

		i := 1
		err := txn1.Commit()
		for err == badger.ErrConflict && i < 10 {
			fmt.Println("[First] Retry Commit #", i)

			txn1 = graph.DB.NewTransaction(true)
			defer txn1.Discard()

			graph.RemoveEdge("a", "b", txn1)

			err = txn1.Commit()
			i++
		}

		if err != nil {
			fmt.Println(err)
		}

		wg.Done()
	}()

	wg.Add(1)
	go func() {
		txn2 := graph.DB.NewTransaction(true)
		defer txn2.Discard()

		graph.RemoveEdge("c", "e", txn2)
		a_n, _ = graph.GetEdges("c", txn2)
		fmt.Println("Neighbors of c: ", a_n)

		i := 1
		err := txn2.Commit()
		for err == badger.ErrConflict && i < 10 {
			fmt.Println("[Second] Retry Commit #", i)

			txn2 = graph.DB.NewTransaction(true)
			defer txn2.Discard()

			graph.RemoveEdge("c", "e", txn2)

			err = txn2.Commit()
			i++
		}

		if err != nil {
			fmt.Println(err)
		}
		wg.Done()
	}()

	wg.Wait()
	a_n, _ = graph.GetEdges("a", nil)
	fmt.Println("Neighbors of a: ", a_n)
	c_n, _ := graph.GetEdges("c", nil)
	fmt.Println("Neighbors of c: ", c_n)
}

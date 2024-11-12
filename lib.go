package Onyx

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/ristretto/z"
	"math/rand"
	"time"
)

// TODO: Add Label support for edgess
type Graph struct {
	DB *badger.DB
}

const allVerticesKey = "all_vertices"

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

	go g.updateVertexList(nil, []string{from, to})

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
		if err == badger.ErrKeyNotFound {
			return make(map[string]bool), nil
		}
		return nil, err
	}

	valCopy, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}

	if localTxn {
		err = txn.Commit()
		if err != nil {
			return nil, err
		}
	}

	neighbors, err := deserializeEdgeMap(valCopy)
	return neighbors, err
}

func (g *Graph) OutDegree(from string, txn *badger.Txn) (int, error) {
	dstNodes, err := g.GetEdges(from, txn)
	if err != nil {
		return 0, err
	}
	return len(dstNodes), nil
}

func (g *Graph) IterAllEdges(f func(src string, dst string) error, prefetchSize int, txn *badger.Txn) error {
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
		src := string(item.Key())

		serVal, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		neighbors, err := deserializeEdgeMap(serVal)
		if err != nil {
			return err
		}

		for neighbor, _ := range neighbors {
			err = f(src, neighbor)
			if err != nil {
				return err
			}
		}

		return nil
	}

	if localTxn {
		err := txn.Commit()
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Graph) PickRandomVertices(num int, txn *badger.Txn) ([]string, error) {
	localTxn := txn == nil
	if localTxn {
		txn = g.DB.NewTransaction(false)
		defer txn.Discard()
	}

	vertices, err := g.GetEdges(allVerticesKey, txn)
	if err != nil {
		return nil, err
	}

	if localTxn {
		err := txn.Commit()
		if err != nil {
			return nil, err
		}
	}

	randomVertices := PickNRandomKeys(vertices, num)
	return randomVertices, nil
}

func (g *Graph) PickRandomVertexLegacy(txn *badger.Txn) (string, error) {
	localTxn := txn == nil
	if localTxn {
		txn = g.DB.NewTransaction(false)
		defer txn.Discard()
	}

	keys := make([][]byte, 0)
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	it := txn.NewIterator(opts)
	defer it.Close()
	c := 0
	for it.Rewind(); it.Valid() && c < 1000; it.Next() {
		item := it.Item()
		k := item.Key()
		keys = append(keys, k)
		c++
	}
	it.Close()

	if localTxn {
		err := txn.Commit()
		if err != nil {
			return "", err
		}
	}

	return string(keys[rand.Intn(len(keys))]), nil
}

func (g *Graph) PickRandomVertexIncorrectEfficient() (string, error) {
	var keys []string
	count := 0
	stream := g.DB.NewStream()
	stream.NumGo = 16

	// overide stream.KeyToList as we only want keys. Also
	// we can take only first version for the key.
	stream.KeyToList = nil
	//stream.KeyToList = func(key []byte, itr *badger.Iterator) (*pb.KVList, error) {
	//	l := &pb.KVList{}
	//	// Since stream framework copies the item's key while calling
	//	// KeyToList, we can directly append key to list.
	//	l.Kv = append(l.Kv, &pb.KV{Key: key})
	//	return l, nil
	//}

	// The bigger the sample size, the more randomness in the outcome.
	sampleSize := 1000
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream.Send = func(buf *z.Buffer) error {
		if count >= sampleSize {
			cancel()
			return nil
		}

		_ = string(buf.Bytes())
		keys = append(keys, fmt.Sprintf("%d", count))
		count++
		return nil
	}

	if err := stream.Orchestrate(ctx); err != nil && err != context.Canceled {
		return "", err
	}

	fmt.Print(keys)
	return keys[rand.Intn(len(keys))], nil
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
	err := d.Decode(&deserializedMap)
	return deserializedMap, err
}

// Add or update the vertex in the all-vertices list
func (g *Graph) updateVertexList(txn *badger.Txn, verticesToAdd []string) error {
	localTxn := txn == nil
	if localTxn {
		txn = g.DB.NewTransaction(true)
		defer txn.Discard()
	}

	item, err := txn.Get([]byte(allVerticesKey))
	var vertices map[string]bool

	if err == badger.ErrKeyNotFound {
		vertices = make(map[string]bool)
	} else if err == nil {
		valCopy, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		vertices, err = deserializeEdgeMap(valCopy)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	vListChanged := false
	for _, vertex := range verticesToAdd {
		// Add the vertex to the map if it doesn’t already exist
		if _, exists := vertices[vertex]; !exists {
			vertices[vertex] = true
			vListChanged = true
		}
	}
	if !vListChanged {
		return nil //no changes to persist
	}

	serializedVertices, err := serializeEdgeMap(vertices)
	if err != nil {
		return err
	}
	err = txn.Set([]byte(allVerticesKey), serializedVertices)
	if err != nil {
		return err
	}

	if localTxn {
		err := txn.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

// PickNRandomKeys picks n unique random keys from the map using reservoir sampling.
func PickNRandomKeys(m map[string]bool, n int) []string {
	keys := make([]string, 0, n)
	i := 0

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	for key := range m {
		if i < n {
			// Directly add the first n keys
			keys = append(keys, key)
		} else {
			// Randomly replace elements in the reservoir with decreasing probability
			j := rand.Intn(i + 1)
			if j < n {
				keys[j] = key
			}
		}
		i++
	}

	// If there are fewer than n keys in the map, just return them
	return keys
}

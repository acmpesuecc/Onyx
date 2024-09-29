package main

import (
	"fmt"
	"github.com/dgraph-io/badger/v4"
	"log"
)

type Graph struct {
	db *badger.DB
}

func NewGraph(path string, inMemory bool) (*Graph, error) {
	db, err := badger.Open(badger.DefaultOptions(path).WithInMemory(inMemory))
	if err != nil {
		return nil, err
	}
	return &Graph{db}, nil
}

func (g *Graph) Close() {
	g.db.Close()
}

//func (g *Graph) AddEdge(from string, to string) error {
//	g.db.Update(func(txn *badger.Txn) error {
//		neighbors, err := txn.Get([]byte(from))
//		if err != nil {
//			return err
//		}
//
//		neighbors.
//	})
//}

func main() {
	db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//set a KV in a txn
	err = db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte("key"), []byte("value"))
		if err != nil {
			return err
		}
		err = txn.Set([]byte("key"), []byte("value2"))
		return err
	})

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("key"))
		if err != nil {
			return err
		}
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		fmt.Println(string(val))
		return nil
	})
}

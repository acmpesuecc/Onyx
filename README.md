# Onyx
Onyx is a embedded, on-disk, concurrent graph database which is built over [badger](https://github.com/dgraph-io/badger) which is aimed at effiecient edge-list scans. Since its a wrapper around badger, Onyx inherits a lot features provided by badger such as:
- Transaction support
- ACID compliant
- Serializable Snapshot Isolation (SSI) guarentee

**Note:** If you are a contributer, please read [CONTRIBUTING.md](https://github.com/acmpesuecc/Onyx/blob/main/CONTRIBUTING.md). **YOU WILL NOT UNDERSTAND THE PROJECT OTHERWISE, PLEASE ACTUALLY FUCKING READ IT**

**Please contact me at +91 8618950413 for any help/queries/assistance**

# Installation 
To install the latest version:
```bash
go get https://github.com/Dynaclo/Onyx
```

# Usage

```go
import "github.com/Dynaclo/Onyx"

graph, err := Onyx.NewGraph("", true)
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
```

## Using Transactions
You can create a `*badger.Txn` and pass it on as the last arguement of every Onyx graph operation function and the graph operation will be executed in that Onyx transaction. If `nil` is passed, the library will execute the operation is a seperate transaction isolated only to that operation
```go
txn1 := graph.DB.NewTransaction(true)
defer txn1.Discard()

graph.AddEdge("e", "f", txn1)
graph.RemoveEdge("a", "b", txn1)
a_n, _ = graph.GetEdges("a", txn1)
fmt.Println("Neighbors of a: ", a_n)

err := txn1.Commit()
if err!=nil {
  return err
}
```

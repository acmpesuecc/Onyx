# Contributing Guidelines

1. Please fork this repo (acmpesuecc/Onyx) to your own account and work on your fork of this repo.
2. Create a PR from your fork to this repo (remember to reference the correct issue in your PR)

# How to run and test changes locally
- This is a go library and Not a regular go project which means that there isn't a main function to run code. Other projects can import this library directly as shown in the README
- This means that for every feature that you implement, the only way for you to test it is to create a associated unit test in the lib_test.go file and test your code using the unit test
- **You are expected to create unit tests for any library functions that you add**
- You may use `go test` to run the entire test suite or `go test -run <unit_test_fn_name>` to run a specific unit test

# Design and Internals
## Transactions
Every Onyx function takes in a `*badger.Txn` as its final parameter. The graph operation function will be executed in the transaction that was passed in. If nil is passed, the library will execute the operation is a seperate transaction isolated only to that operation. 

Therefore it is **extremely important** that every publically exposed graph op function that you implement to extend Onyx follows the following 3 rules:
1. Must have `txn *badger.Txn` as its final paramater
2. Must Start with:
```go
localTxn := txn == nil
if localTxn {
  txn = g.DB.NewTransaction(<IsRW?>)
  defer txn.Discard()
}
```
3. Must End with:
```go
if localTxn {
  err = txn.Commit()
  if err != nil {
    return <based on return type>..., err
  }
}
```
The above 2 code blocks implement local transactions.  
- `<IsRW?>` is `true` for functions which write to the graph (like `AddEdge`, `RemoveEdge`) and creates a Read-Write transaction.  
- `<IsRW?>` is `false` for functions which only read from the graph (like `GetEdges`, `IterAllEdges`) and creates a Read-Write transaction. 

## Internals
All the data of the graph is stored in
```go
type Graph struct {
	DB *badger.DB
}
```
ie a singular DB attribute which is a pointer to a badgerdb kv store, which as promised in the README is the underlying data structure which stores the graph and gives us all the nice features that Onyx can claim. badger requires the key and value to be of type `[]byte` ie a arbitary byte array with us having to take care of all the serialization and deserialization to whatever types we need

- The graph is stored in the badger key value store as a series of edge lists for every source node.
- An edge list is a list of all the destination nodes `dst` for a particular source node `src` for which the edge `src->dst` exists in the graph.
- In the Onyx codebase, this edgelist is represented as a `map[string]bool` where the `string` is the destination node in the edge list and the `bool` is arbitary. This is because go does not have a native set data structure. A `map` has been used instead of a slice as it allows us to check if a particular dst node exists in the edge list in O(1) tiem instead of traversing a list. If a key with the dest node exists in the map, it exisst in the edgelist and vice versa.

So here is how it works, we store KV pairs in badger where:
- **Key** is the source node of the edge. It is a `string` which is converted to `[]byte` and back using the simple `string()` and `[]byte()`
- **Value** is the edge list for the source node that is the Key. The Edge List is a `map[string]bool` as explained above and is serialized to `[]byte` and back using `gob`

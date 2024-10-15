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

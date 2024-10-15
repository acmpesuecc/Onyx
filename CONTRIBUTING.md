# Contributing Guidelines

1. Please fork this repo (acmpesuecc/Onyx) to your own account and work on your fork of this repo.
2. Create a PR from your fork to this repo (remember to reference the correct issue in your PR)

# How to run and test changes locally
- This is a go library and Not a regular go project which means that there isn't a main function to run code. Other projects can import this library directly as shown in the README
- This means that for every feature that you implement, the only way for you to test it is to create a associated unit test in the lib_test.go file and test your code using the unit test
- **You are expected to create unit tests for any library functions that you add**
- You may use `go test` to run the entire test suite or `go test -run <unit_test_fn_name>` to run a specific unit test

# Onyx
Onyx is a embedded, on-disk, concurrent graph database which is built over [badger](https://github.com/dgraph-io/badger) which is aimed at effiecient edge-list scans. Since its a wrapper around badger, Onyx inherits a lot features provided by badger such as:
- Transaction support
- ACID compliant
- Serializable Snapshot Isolation (SSI) guarentee

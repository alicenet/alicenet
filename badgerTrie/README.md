# Sparse Merkle Tree

A performance oriented implementation of a binary SMT with parallel update, node batching and storage shortcuts.

Details of the SMT implementation : [https://medium.com/@ouvrard.pierre.alain/sparse-merkle-tree-86e6e2fc26da](https://medium.com/@ouvrard.pierre.alain/sparse-merkle-tree-86e6e2fc26da)

### Features

- Efficient Merkle proof verification (**binary** tree structure)
- Compressed Merkle proofs
- Efficient database reads and storage (**node batching**)
- Reduced data storage (**shortcut nodes** for subtrees containing one key)
- Simultaneous update of multiple keys with goroutines

### usage

the following piece of example code shows how to setup an instance of badger db, a new sparse merkle tree with that database, and how to commit new data to it

```
dbPath := path.Join(".db", "db")
if _, err := os.Stat(dbPath); os.IsNotExist(err) {
    _ = os.MkdirAll(dbPath, 0711)
}
st := NewDB(dbPath)

smt := NewSMT(nil, Hasher, st)
keys := getFreshData(32, 32)
values := getFreshData(32, 32)
smt.Update(keys, values)
smt.Commit()
```

in that example, the getFreshData function returns 32 byte slices that are each contain 32 bytes, so 32 keys and values are added to the sparse merkle tree. instead of a badgerdb instance, there is also an in memory options which stores keys and values with a map data structure

```
dbPath := path.Join(".db", "db")
if _, err := os.Stat(dbPath); os.IsNotExist(err) {
    _ = os.MkdirAll(dbPath, 0711)
}
st := NewMemoryDB(dbPath)
```

additionally, there is an option to use an unmanaged version of badgerdb, if you would like to use your own instance without the default option which comes with a predefined setup and garbage collection. the following is an example of how you could create a sparse merkle tree with the unmanaged badgerdb

```
func getPrefix() []byte {
    return []byte {3,5}
}

dbPath := path.Join(".db", "db")
if _, err := os.Stat(dbPath); os.IsNotExist(err) {
    _ = os.MkdirAll(dbPath, 0711)
}

opts := badger.DefaultOptions(dir)
db, err := badger.Open(opts)
if err != nil {
	return nil, err
}

st := NewUnmanagedBadgerDB(dbPath, db, getPrefix)

smt := NewSMT(nil, Hasher, st)
```

in that example, you are free to change how you get your \*badger.DB object and the prefix function. the prefix function will allow you to segregate the data in this tree from other data in the database. once you have a sparse merkle tree object, you can create merkle proofs and verify them. the following example shows one way you could do this

```
val, err := smt.Get(keys[0])
if err != nil {
    log.Fatal(err)
}

bitset, mp, err := smt.MerkleProofCompressed(keys[0])
if err != nil {
    log.Fatal(err)
}

result := smt.VerifyMerkleProofCompressed(bitset, mp, keys[0], val)
```

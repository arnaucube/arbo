// Package goleveldb implements arbo database interfaces using LevelDB.
package goleveldb

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/arnaucube/arbo/db"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// LevelDB implements db.Database using LevelDB.
type LevelDB struct {
	db *leveldb.DB
}

// Ensure that LevelDB implements the db.Database interface
var _ db.Database = (*LevelDB)(nil)

// New returns a LevelDB which implements the db.Database interface
func New(opts db.Options) (*LevelDB, error) {
	// Open the LevelDB database
	db, err := leveldb.OpenFile(opts.Path, &opt.Options{})
	if err != nil {
		return nil, fmt.Errorf("could not open leveldb: %w", err)
	}
	return &LevelDB{
		db: db,
	}, nil
}

// Close closes the database.
func (d *LevelDB) Close() error {
	return d.db.Close()
}

// WriteTx creates a buffered write transaction.
func (d *LevelDB) WriteTx() db.WriteTx {
	return &WriteTx{
		db:    d.db,
		batch: new(leveldb.Batch),
	}
}

// Get retrieves the value associated with key.
func (d *LevelDB) Get(key []byte) ([]byte, error) {
	val, err := d.db.Get(key, nil)
	if errors.Is(err, leveldb.ErrNotFound) {
		return nil, db.ErrKeyNotFound
	}
	if err != nil {
		return nil, err
	}
	return val, nil
}

// Iterate visits keys with the given prefix in lexicographic order.
func (d *LevelDB) Iterate(prefix []byte, callback func(key, value []byte) bool) error {
	iter := d.db.NewIterator(util.BytesPrefix(prefix), nil)
	defer iter.Release()
	for iter.Next() {
		if !callback(iter.Key(), iter.Value()) {
			break
		}
	}
	return iter.Error()
}

// Set stores a key-value pair.
func (d *LevelDB) Set(key, value []byte) error {
	return d.db.Put(key, value, nil)
}

// Delete removes a key.
func (d *LevelDB) Delete(key []byte) error {
	return d.db.Delete(key, nil)
}

// Commit writes a batch to the database.
func (d *LevelDB) Commit(batch *leveldb.Batch) error {
	return d.db.Write(batch, nil)
}

// Compact implements the db.Database.Compact interface method.
func (d *LevelDB) Compact() error {
	return d.db.CompactRange(util.Range{})
}

// WriteTx implements the interface db.WriteTx for goleveldb
type WriteTx struct {
	batch      *leveldb.Batch
	db         *leveldb.DB
	inMemBatch sync.Map
}

// check that WriteTx implements the db.WriteTx interface
var _ db.WriteTx = (*WriteTx)(nil)

// Get retrieves a value, including pending writes.
func (tx *WriteTx) Get(k []byte) ([]byte, error) {
	val, ok := tx.inMemBatch.Load(string(k))
	if !ok {
		val, err := tx.db.Get(k, nil)
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, db.ErrKeyNotFound
		}
		return val, err
	}
	return val.([]byte), nil
}

// Iterate visits pending writes and stored keys with the given prefix.
func (tx *WriteTx) Iterate(prefix []byte, callback func(k, v []byte) bool) error {
	inMemory := make(map[string]bool)
	tx.inMemBatch.Range(func(k, v any) bool {
		keyBytes := []byte(k.(string))
		if bytes.HasPrefix(keyBytes, prefix) {
			inMemory[string(keyBytes)] = true
			return callback(keyBytes, v.([]byte))
		}
		return true
	})
	iter := tx.db.NewIterator(util.BytesPrefix(prefix), nil)
	defer iter.Release()
	for iter.Next() {
		if inMemory[string(iter.Key())] {
			continue
		}
		if !callback(iter.Key(), iter.Value()) {
			break
		}
	}
	return iter.Error()
}

// Set buffers a key-value pair.
func (tx *WriteTx) Set(k, v []byte) error {
	tx.batch.Put(k, v)
	tx.inMemBatch.Store(string(k), v)
	return nil
}

// Delete buffers the removal of a key.
func (tx *WriteTx) Delete(k []byte) error {
	tx.batch.Delete(k)
	tx.inMemBatch.Delete(string(k))
	return nil
}

// Apply copies the entries from another transaction.
func (tx *WriteTx) Apply(otherTx db.WriteTx) error {
	return otherTx.Iterate(nil, func(k, v []byte) bool {
		tx.inMemBatch.Store(string(k), v)
		tx.batch.Put(k, v)
		return true
	})
}

// Commit writes the buffered batch to the database.
func (tx *WriteTx) Commit() error {
	return tx.db.Write(tx.batch, nil)
}

// Discard clears the buffered writes.
func (tx *WriteTx) Discard() {
	tx.batch.Reset()
	tx.inMemBatch = sync.Map{}
}

package arbo

import (
	"testing"

	"github.com/arnaucube/arbo/db"
	"github.com/arnaucube/arbo/db/goleveldb"
	"github.com/arnaucube/arbo/db/pebbledb"
	qt "github.com/frankban/quicktest"
)

// TestLocalDBReaders checks pending and committed tree reads through the ported API.
func TestLocalDBReaders(t *testing.T) {
	backends := map[string]func(db.Options) (db.Database, error){
		"pebble":  func(opts db.Options) (db.Database, error) { return pebbledb.New(opts) },
		"leveldb": func(opts db.Options) (db.Database, error) { return goleveldb.New(opts) },
	}
	for name, open := range backends {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			opts := db.Options{Path: t.TempDir()}
			database, err := open(opts)
			c.Assert(err, qt.IsNil)
			t.Cleanup(func() { c.Check(database.Close(), qt.IsNil) })
			cfg := Config{Database: database, MaxLevels: 32, HashFunction: HashFunctionBlake2b}
			tree, err := NewTree(cfg)
			c.Assert(err, qt.IsNil)

			tx := database.WriteTx()
			defer tx.Discard()
			key, value := []byte{1}, []byte("value")
			c.Assert(tree.AddWithTx(tx, key, value), qt.IsNil)
			_, got, err := tree.GetWithTx(tx, key)
			c.Assert(err, qt.IsNil)
			c.Check(got, qt.DeepEquals, value)
			_, _, err = tree.GetWithTx(database, key)
			c.Check(err, qt.Equals, ErrKeyNotFound)
			pendingRoot, err := tree.RootWithTx(tx)
			c.Assert(err, qt.IsNil)
			c.Assert(tx.Commit(), qt.IsNil)
			tx.Discard()

			root, err := tree.RootWithTx(database)
			c.Assert(err, qt.IsNil)
			c.Check(root, qt.DeepEquals, pendingRoot)
			_, got, err = tree.Get(key)
			c.Assert(err, qt.IsNil)
			c.Check(got, qt.DeepEquals, value)
		})
	}
}

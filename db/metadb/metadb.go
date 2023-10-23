// Package metadb opens supported database implementations by type.
package metadb

import (
	"fmt"
	"os"
	"testing"

	"github.com/arnaucube/arbo/db"
	"github.com/arnaucube/arbo/db/goleveldb"
	"github.com/arnaucube/arbo/db/pebbledb"
)

// New opens a database of the requested type in dir.
func New(typ, dir string) (db.Database, error) {
	var database db.Database
	var err error
	opts := db.Options{Path: dir}
	switch typ {
	case db.TypePebble:
		database, err = pebbledb.New(opts)
		if err != nil {
			return nil, err
		}
	case db.TypeLevelDB:
		database, err = goleveldb.New(opts)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid dbType: %q. Available types: %q %q",
			typ, db.TypePebble, db.TypeLevelDB)
	}
	return database, nil
}

// ForTest selects the test database type from DVOTE_DB_TYPE, defaulting to Pebble.
func ForTest() (typ string) {
	if typ := os.Getenv("DVOTE_DB_TYPE"); typ != "" {
		return typ
	}
	return db.TypePebble // default to Pebble
}

// NewTest opens a temporary database and registers its cleanup.
func NewTest(tb testing.TB) db.Database {
	database, err := New(ForTest(), tb.TempDir())
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() {
		if err := database.Close(); err != nil {
			tb.Error(err)
		}
	})
	return database
}

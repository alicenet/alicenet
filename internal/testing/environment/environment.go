// Package environment contains testing functions for setting up a testing environment.
package environment

import (
	"testing"

	"github.com/dgraph-io/badger/v2"
)

// SetupBadgerDatabase for a test and clean up afterwards.
func SetupBadgerDatabase(t *testing.T) *badger.DB {
	t.Helper()
	dir := t.TempDir()
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	return db
}

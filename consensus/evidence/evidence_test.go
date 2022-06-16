package evidence

import (
	"testing"

	"github.com/MadBase/MadNet/consensus/db"
)

func TestNewPool(t *testing.T) {
	t.Parallel()
	db := &db.Database{}
	pool := NewPool(db)
	if pool.database != db {
		t.Errorf("database not set correct. want: %v, got: %v", pool.database, db)
	}
	if pool.store == nil {
		t.Error("store not set")
	}
}

package contractevent

import (
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	// "github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestCreateTable(t *testing.T) {
	dbName := "gorm_test.db"
	os.Remove(dbName)
	defer os.Remove(dbName)
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	ldb, _ := db.DB()
	defer ldb.Close()
	alias := "alias_name"
	err = CreateTable(db, alias)
	if err != nil {
		t.Fatal(err)
	}

}

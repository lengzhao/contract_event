package contractevent

import (
	"os"
	"testing"

	"gorm.io/driver/sqlite"
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

func TestSetNotifyRecord(t *testing.T) {
	dbName := "gorm_test1.db"
	os.Remove(dbName)
	defer os.Remove(dbName)
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	CreateNotifyRecord(db)
	ldb, _ := db.DB()
	defer ldb.Close()
	alias := "alias001"
	err = SetNotifyRecord(db, alias, 100)
	if err != nil {
		t.Fatal(err)
	}

	id, err := GetNotifyRecord(db, alias)
	if err != nil {
		t.Fatal(err)
	}
	if id != 100 {
		t.Fatal("error id,hope:100,get:", id)
	}

	err = SetNotifyRecord(db, alias, 100)
	if err != nil {
		t.Fatal(err)
	}
	err = SetNotifyRecord(db, alias, 10)
	if err != nil {
		t.Fatal(err)
	}
	id, err = GetNotifyRecord(db, alias)
	if err != nil {
		t.Fatal(err)
	}
	if id != 100 {
		t.Fatal("error id,hope:100,get:", id)
	}
}

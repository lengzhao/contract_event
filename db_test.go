package contractevent

import (
	"fmt"
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
	err = CreateEventTable(db, alias)
	if err != nil {
		t.Fatal(err)
	}

	item := DBItem{TX: "0x1234", LogIndex: 1, Others: []byte{11, 22}}
	n, err := InsertItem(db, alias, item)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("insert item:", n, item)
	n2, err := InsertItem(db, alias, item)
	if err != nil {
		t.Fatal(err)
	}
	if n2 != n {
		t.Fatal("different id:", n, n2)
	}

	item2 := DBItem{TX: "0x1234", LogIndex: 2, Others: []byte{11, 22}}
	_, err = InsertItem(db, alias, item2)
	if err != nil {
		t.Fatal(err)
	}

	items, err := ListItems(db, alias, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatal(items)
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

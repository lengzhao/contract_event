package contractevent

import (
	"fmt"

	"gorm.io/gorm"
)

type DBItem struct {
	gorm.Model
	TX       string `gorm:"uniqueIndex:idx_log;column:tx"`
	LogIndex uint   `gorm:"uniqueIndex:idx_log;column:log_index"`
	Others   []byte
}

func dyncTable(db *gorm.DB, alias string) *gorm.DB {
	return db.Table("event_" + alias)
}

func CreateTable(db *gorm.DB, alias string) error {
	err := dyncTable(db, alias).AutoMigrate(&DBItem{})
	if err != nil {
		return nil
	}
	name := dyncTable(db, alias).Statement.Table
	rst := db.Exec(fmt.Sprintf("CREATE UNIQUE INDEX idx_%s ON %s(tx)", name, name))
	return rst.Error
}

func InsertItem(db *gorm.DB, alias string, item DBItem) (uint, error) {
	rst := dyncTable(db, alias).Create(&item)
	if rst.Error != nil {
		var it DBItem
		it.TX = item.TX
		it.LogIndex = item.LogIndex
		dyncTable(db, alias).Where(&it).First(&it)
		if it.ID > 0 {
			return it.ID, nil
		}
		return 0, rst.Error
	}
	return item.ID, nil
}

func ListItems(db *gorm.DB, alias string, offest, limit int) ([]DBItem, error) {
	var out []DBItem
	rst := dyncTable(db, alias).Offset(offest).Limit(limit).Find(out)
	return out, rst.Error
}

func DeleteItem(db *gorm.DB, alias string, id uint) error {
	rst := dyncTable(db, alias).Delete(&DBItem{}, id)
	return rst.Error
}

func ItemsTotal(db *gorm.DB, alias string) (uint, error) {
	var it DBItem
	rst := dyncTable(db, alias).Last(&it)
	return it.ID, rst.Error
}

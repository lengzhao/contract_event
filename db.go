package contractevent

import (
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBConf struct {
	Type     string `json:"type,omitempty" yaml:"type"`
	DSN      string `json:"dsn,omitempty" yaml:"dsn"`
	LogLevel int    `json:"log_level,omitempty" yaml:"log_level"`
}

func NewDB(conf DBConf) (*gorm.DB, error) {
	gConf := gorm.Config{}
	gConf.Logger = logger.New(log.StandardLogger(), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.LogLevel(conf.LogLevel),
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
	})

	switch conf.Type {
	case "mysql":
		return gorm.Open(mysql.Open(conf.DSN), &gConf)
	case "postgres":
		return gorm.Open(postgres.Open(conf.DSN), &gConf)
	case "sqlite":
		return gorm.Open(sqlite.Open(conf.DSN), &gConf)
	case "sqlserver":
		return gorm.Open(sqlserver.Open(conf.DSN), &gConf)
	default:
		return nil, fmt.Errorf("not support")
	}
}

type DBItem struct {
	gorm.Model
	TX       string `gorm:"column:tx"`
	LogIndex uint   `gorm:"column:log_index"`
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
	rst := db.Exec(fmt.Sprintf("CREATE UNIQUE INDEX idx_%s ON %s(tx,log_index)", name, name))
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
		log.Warn("fail to insert item:", alias, it.TX, it.LogIndex, rst.Error)
		return 0, rst.Error
	}
	return item.ID, nil
}

func ListItems(db *gorm.DB, alias string, offest, limit int) ([]DBItem, error) {
	var out []DBItem
	rst := dyncTable(db, alias).Offset(offest).Limit(limit).Find(&out)
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

type NotifyRecord struct {
	gorm.Model
	Alias      string `gorm:"uniqueIndex;column:alias"`
	NotifiedID uint   `gorm:"column:nid"`
}

func CreateNotifyRecord(db *gorm.DB) error {
	return db.AutoMigrate(&NotifyRecord{})
}

func GetNotifyRecord(db *gorm.DB, alias string) (uint, error) {
	var out NotifyRecord
	rst := db.Model(&NotifyRecord{}).Where("alias = ?", alias).First(&out)
	if errors.Is(rst.Error, gorm.ErrRecordNotFound) {
		return 0, nil
	}

	return out.NotifiedID, rst.Error
}

func SetNotifyRecord(db *gorm.DB, alias string, nid uint) error {
	var record NotifyRecord
	db.Model(&NotifyRecord{}).Where("alias = ?", alias).First(&record)
	if nid <= record.NotifiedID {
		return nil
	}
	record.Alias = alias
	record.NotifiedID = nid
	if record.ID == 0 {
		rst := db.Create(&record)
		return rst.Error
	}
	rst := db.Model(&NotifyRecord{}).Where("alias = ?", alias).Where("nid < ?", nid).Update("nid", nid)
	return rst.Error
}

type BlockRecord struct {
	gorm.Model
	Alias   string `gorm:"uniqueIndex;column:alias"`
	BlockID uint64 `gorm:"column:block_id"`
}

func CreateBlockRecord(db *gorm.DB) error {
	return db.AutoMigrate(&BlockRecord{})
}

func GetBlockRecord(db *gorm.DB, alias string) (uint64, error) {
	var out BlockRecord
	rst := db.Model(&BlockRecord{}).Where("alias = ?", alias).First(&out)
	if errors.Is(rst.Error, gorm.ErrRecordNotFound) {
		return 0, nil
	}

	return out.BlockID, rst.Error
}

func SetBlockRecord(db *gorm.DB, alias string, blockID uint64) error {
	var record BlockRecord
	db.Model(&BlockRecord{}).Where("alias = ?", alias).First(&record)
	if blockID <= record.BlockID {
		return nil
	}
	record.Alias = alias
	record.BlockID = blockID
	if record.ID == 0 {
		rst := db.Create(&record)
		return rst.Error
	}
	rst := db.Model(&BlockRecord{}).Where("alias = ?", alias).Where("block_id < ?", blockID).Update("block_id", blockID)
	return rst.Error
}

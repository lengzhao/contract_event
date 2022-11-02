package contractevent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type NotifyTask struct {
	alias   string
	db      *gorm.DB
	webHook string
}

func NewNotifyTask(db *gorm.DB, alias, webHook string) *NotifyTask {
	SetNotifyRecord(db, alias, 0)
	return &NotifyTask{alias: alias, db: db, webHook: webHook}
}

func (t *NotifyTask) Run(limit uint) error {
	id, err := GetNotifyRecord(t.db, t.alias)
	if err != nil {
		log.Errorln("fail to get record id:", err)
		return err
	}
	items, err := ListItems(t.db, t.alias, int(id), int(limit))
	if err != nil {
		return err
	}
	last := id
	for _, it := range items {
		info := make(map[string]interface{})
		info["local_id"] = it.ID
		json.Unmarshal(it.Others, &info)
		data, _ := json.Marshal(info)
		resp, err := http.DefaultClient.Post(t.webHook, "application/json", bytes.NewReader(data))
		if err != nil {
			log.Errorln("fail to Post:", err)
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Errorln("notify, get wrong code(hope 200 OK):", resp.Status)
			return fmt.Errorf("error response code:%s", resp.Status)
		}
		if it.ID > last {
			last = it.ID
		}
	}
	return SetNotifyRecord(t.db, t.alias, last)
}

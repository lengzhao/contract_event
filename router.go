package contractevent

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ginRouter struct {
	db *gorm.DB
}

func HttpRouter(router *gin.Engine, db *gorm.DB) {
	lr := ginRouter{db}
	router.GET("/logs", lr.getEvent)
	router.GET("/unnotified_logs", lr.requestUnnotifiedEvent)
}

type reqLogParam struct {
	Alias  string
	Offset int
	Limit  int
}

type RespItems struct {
	Alias  string                   `json:"alias,omitempty"`
	Offset int                      `json:"offset,omitempty"`
	Total  uint                     `json:"total,omitempty"`
	Items  []map[string]interface{} `json:"items,omitempty"`
}

func (lr *ginRouter) getEvent(c *gin.Context) {
	var param reqLogParam
	err := c.BindQuery(&param)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if param.Alias == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request alias"})
		return
	}
	if param.Limit < 1 {
		param.Limit = 20
	}
	var out RespItems
	out.Alias = param.Alias
	out.Offset = param.Offset
	out.Total, _ = ItemsTotal(lr.db, param.Alias)
	items, _ := ListItems(lr.db, param.Alias, param.Offset, param.Limit)
	for _, it := range items {
		info := make(map[string]interface{})
		info["local_id"] = it.ID
		json.Unmarshal(it.Others, &info)
		out.Items = append(out.Items, info)
	}
	c.JSON(http.StatusOK, out)
}

type reqUnnotifiedParam struct {
	Alias string
	Limit int
}

func (lr *ginRouter) requestUnnotifiedEvent(c *gin.Context) {
	var param reqUnnotifiedParam
	err := c.BindQuery(&param)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if param.Alias == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request alias"})
		return
	}
	if param.Limit < 1 {
		param.Limit = 20
	}
	var out RespItems
	out.Alias = param.Alias
	out.Total, _ = ItemsTotal(lr.db, param.Alias)
	offset, err := GetNotifyRecord(lr.db, param.Alias)
	if err != nil {
		return
	}
	items, _ := ListItems(lr.db, param.Alias, int(offset), param.Limit)
	last := offset
	for _, it := range items {
		info := make(map[string]interface{})
		info["local_id"] = it.ID
		json.Unmarshal(it.Others, &info)
		out.Items = append(out.Items, info)
		if it.ID > last {
			last = it.ID
		}
	}
	SetNotifyRecord(lr.db, param.Alias, last)
	c.JSON(http.StatusOK, out)
}

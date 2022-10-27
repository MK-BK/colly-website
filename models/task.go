package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

const (
	TaskURL     = 0
	TaskContent = 1
)

const (
	StatsuCreate   = "create"
	StatsuRunning  = "running"
	StatsuFailed   = "Failed"
	StatsuComplete = "complete"
)

type Task struct {
	gorm.Model
	URL  string `json:"url"`
	Type int    `json:"type"`
}

type TaskResult struct {
	gorm.Model
	TaskID   uint        `json:"task_id"`
	TaskType int         `json:"task_type"`
	Data     ResultDatas `gorm:"TYPE:json"`
	Status   string      `json:"status"`
}

type ResultData struct {
	URL     string `json:"url"`
	Content string `json:"content"` //type为0时此字段为空，1时此字段保存页面文本内容
}

type ResultDatas []ResultData

func (r *ResultDatas) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	return json.Unmarshal(bytes, r)
}

func (r ResultDatas) Value() (driver.Value, error) {
	return json.Marshal(r)
}

package models

const (
	TaskURL     = 0
	TaskContent = 1
)

const (
	StatsuCreate   = "Create"
	StatsuRunning  = "Running"
	StatsuFailed   = "Failed"
	StatsuComplete = "Complete"
)

type Task struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	Type int    `json:"type"`
}

type TaskResult struct {
	TaskID   string       `json:"task_id"`
	TaskType int          `json:"task_type"`
	Data     []ResultData `gorm:"TYPE:json"`
	Status   string       `json:"status"`
}

type ResultData struct {
	URL     string `json:"url"`
	Content string `json:"content"` //type为0时此字段为空，1时此字段保存页面文本内容
}

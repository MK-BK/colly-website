package task

import (
	"bytes"
	"context"
	"strings"

	"colly-website/db"
	"colly-website/models"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	log             = logrus.New()
	defaultTaskSize = 20
)

type TaskManager struct {
	chs chan *models.Task
}

func (t *TaskManager) Create(ctx context.Context, task *models.Task) error {
	if err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(task).Error; err != nil {
			return err
		}

		result := &models.TaskResult{
			TaskID:   task.ID,
			TaskType: task.Type,
			Data:     make([]models.ResultData, 0),
			Status:   models.StatsuCreate,
		}

		if err := tx.Save(result).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	t.chs <- task
	return nil
}

func (t *TaskManager) Get(ctx context.Context, taskID uint) (*models.TaskResult, error) {
	result := models.TaskResult{}
	if err := db.DB.First(&result, taskID).Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (t *TaskManager) Init() {
	for i := 0; i < defaultTaskSize; i++ {
		go func() {
			for {
				task, ok := <-t.chs
				if !ok {
					return
				}

				result, err := t.Get(context.Background(), task.ID)
				if err != nil {
					continue
				}

				// update result.Status -> ‘running’
				log.Info("update task status -> running!")
				if err := db.DB.Model(result).Update("status", models.StatsuRunning).Error; err != nil {
					log.Error(err)
					continue
				}

				c := colly.NewCollector(
					colly.AllowedDomains(parseDomain(task.URL)),
					colly.MaxDepth(1),
				)

				c.OnResponse(func(resp *colly.Response) {
					url := resp.Request.URL.String()

					if task.Type == models.TaskURL {
						result.Data = append(result.Data, models.ResultData{
							URL: url,
						})
					} else {
						result.Data = append(result.Data, models.ResultData{
							URL:     url,
							Content: parseContent(resp.Body),
						})
					}
				})

				c.OnHTML("a", func(e *colly.HTMLElement) {
					c.Visit(e.Attr("href"))
				})

				c.Visit(task.URL)

				log.Info("update task status -> conplete && data!")
				if err := db.DB.Debug().Model(&result).Updates(models.TaskResult{
					Status: models.StatsuComplete,
					Data:   result.Data,
				}).Error; err != nil {
					log.Error(err)
					continue
				}

			}
		}()
	}
}

func NewTaskManager() *TaskManager {
	m := &TaskManager{
		chs: make(chan *models.Task, defaultTaskSize),
	}

	m.Init()
	return m
}

func parseContent(body []byte) string {
	document, err := goquery.NewDocumentFromReader(bytes.NewBuffer(body))
	if err != nil {
		return ""
	}

	return strings.ReplaceAll(strings.ReplaceAll(document.Not("script").Text(), "\n", ""), "\t", "")
}

func parseDomain(url string) string {
	arrs := strings.Split(url, "://")
	if len(arrs) > 0 {
		return arrs[1]
	}
	return ""
}

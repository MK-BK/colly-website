package task

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"colly-website/models"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/patrickmn/go-cache"
)

var log = logrus.New()
var concurrency = 100

type TaskManager struct {
	chs   chan *models.Task
	Cache *cache.Cache
}

func NewTaskManager() *TaskManager {
	m := &TaskManager{
		chs:   make(chan *models.Task, concurrency),
		Cache: cache.New(30*time.Minute, 30*time.Minute),
	}

	m.Init()
	return m
}

func (t *TaskManager) Init() {
	for i := 0; i < concurrency; i++ {
		go func() {
			for {
				task, ok := <-t.chs
				if !ok {
					return
				}

				result := &models.TaskResult{
					TaskID:   task.ID,
					TaskType: task.Type,
					Status:   models.StatsuRunning,
					Data:     make([]models.ResultData, 0),
				}

				t.Cache.SetDefault(task.ID, result)

				URL, err := url.Parse(task.URL)
				if err != nil {
					log.Error(err)
					continue
				}

				log.Warn(URL.Host)

				c := colly.NewCollector(
					colly.AllowedDomains(URL.Host),
					colly.Async(true),
					colly.ParseHTTPErrorResponse(),
				)

				rule := &colly.LimitRule{
					RandomDelay: time.Second,
					Parallelism: 10,
				}

				c.Limit(rule)

				c.SetClient(&http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: true,
						},
					},
				})

				c.OnResponse(func(resp *colly.Response) {
					if resp.StatusCode != http.StatusOK {
						return
					}

					if strings.HasPrefix(resp.Request.URL.String(), task.URL) && strings.Index(resp.Headers.Get("Content-Type"), "text/html") > -1 {
						if task.Type == models.TaskURL {
							result.Data = append(result.Data, models.ResultData{
								URL: resp.Request.URL.String(),
							})
						} else {
							result.Data = append(result.Data, models.ResultData{
								URL:     resp.Request.URL.String(),
								Content: parseContent(resp.Body),
							})
						}
					}
				})

				c.OnHTML("a", func(e *colly.HTMLElement) {
					link := e.Attr("href")
					c.Visit(e.Request.AbsoluteURL(link))
					// if strings.HasPrefix(href, task.URL) {
					// 	c.Visit(href)
					// } else if strings.HasPrefix(href, "/") && !strings.HasPrefix(href, "//") {
					// 	c.Visit(task.URL + href)
					// }
				})

				c.OnError(func(resp *colly.Response, err error) {
					log.Error("OnError:", err)
				})
				start := time.Now()

				log.Infof("%+v: start to visit URL: %s", start, task.URL)
				c.Visit(task.URL)

				c.Wait()
				log.Infof("Colly Visit complete, %+v, spend: %+v", task.ID, time.Now().Sub(start))

				result.Status = models.StatsuComplete
				t.Cache.SetDefault(task.ID, result)
			}
		}()
	}
}

func parseContent(body []byte) string {
	document, err := goquery.NewDocumentFromReader(bytes.NewBuffer(body))
	if err != nil {
		return ""
	}

	text := make([]string, 0)
	document.Find("div").Each(func(i int, s *goquery.Selection) {
		s.ReplaceWithSelection(s.Find("script")).ReplaceWithSelection(s.Find("style")).ReplaceWithSelection(s.Find("textarea")).ReplaceWithSelection(s.Find("noscript")).Each(func(i int, s *goquery.Selection) {
			txt := s.Text()
			text = append(text, strings.Join(strings.Fields(strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(txt, "\n", ""), "\t", ""))), ""))
		})
	})

	return strings.Join(text, "")
}

func (t *TaskManager) Create(ctx context.Context, task *models.Task) error {
	task.ID = uuid.New().String()

	result := &models.TaskResult{
		TaskID:   task.ID,
		TaskType: task.Type,
		Data:     make([]models.ResultData, 0),
		Status:   models.StatsuCreate,
	}

	t.Cache.SetDefault(result.TaskID, result)
	t.chs <- task

	return nil
}

func (t *TaskManager) Get(ctx context.Context, taskID string) (*models.TaskResult, error) {
	v, ok := t.Cache.Get(taskID)
	if !ok {
		return nil, errors.New("result not found")
	}

	if result, ok := v.(*models.TaskResult); ok {
		if result.Status == models.StatsuComplete {
			return result, nil
		}

		return &models.TaskResult{
			TaskID:   result.TaskID,
			TaskType: result.TaskType,
			Status:   result.Status,
		}, nil
	}

	return nil, errors.New("result not found")
}

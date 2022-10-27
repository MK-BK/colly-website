package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"colly-website/db"
	"colly-website/models"
	"colly-website/task"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	conf        = &models.Config{}
	log         = logrus.New()
	taskManager *task.TaskManager
)

func main() {
	if err := loadConfig("config.json"); err != nil {
		panic(err)
	}

	if err := db.InitDB(conf); err != nil {
		panic(err)
	}

	taskManager = task.NewTaskManager()
	startServer()
}

func loadConfig(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, conf)
}

func startServer() {
	e := gin.Default()

	e.GET("/:id", getTask)
	e.POST("/", createTask)

	if err := http.ListenAndServe(":"+conf.Listen, e); err != nil {
		log.Fatal(err)
	}
}

func createTask(ctx *gin.Context) {
	task := models.Task{}

	if err := ctx.ShouldBind(&task); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)

	}
	if err := taskManager.Create(ctx, &task); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"task_id": task.ID,
	})
}

func getTask(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("param id must empty"))
	}

	taskID, err := strconv.Atoi(id)
	result, err := taskManager.Get(ctx.Request.Context(), uint(taskID))
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}

	ctx.JSON(http.StatusOK, result)
}

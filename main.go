package main

import (
	"errors"
	"net/http"

	"colly-website/models"
	"colly-website/task"

	"github.com/gin-gonic/gin"
)

var taskManager *task.TaskManager

func main() {
	taskManager = task.NewTaskManager()
	startServer()
}

func startServer() {
	e := gin.Default()

	e.GET("/:id", getTask)
	e.POST("/", createTask)

	if err := http.ListenAndServe(":10086", e); err != nil {
		panic(err)
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

	result, err := taskManager.Get(ctx.Request.Context(), id)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}

	ctx.JSON(http.StatusOK, result)
}

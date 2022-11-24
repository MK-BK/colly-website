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

	result, err := taskManager.Get(ctx, id)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
	}

	// f, err := os.OpenFile("/Users/len/go/src/colly-website/log.txt", os.O_APPEND|os.O_RDWR, 0664)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// b, _ := json.Marshal(result)
	// _, err = f.Write(b)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// ctx.Writer.WriteHeader(http.StatusOK)
	// ctx.Header("Content-Disposition", "attachment; filename=a.tar")
	// ctx.Header("Content-Type", "application/octet-stream")
	// ctx.Header("Content-Length", fmt.Sprintf("%d", len(b)))
	// ctx.Writer.Write(b) //the memory take up 1.2~1.7G

	ctx.JSON(http.StatusOK, result)
}

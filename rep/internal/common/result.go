package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetResult(context *gin.Context, tasks map[string]Task) {
	taskId := context.Param("id")

	task, exists := tasks[taskId]

	if !exists {
		context.AbortWithStatusJSON(http.StatusNotFound, "Task not found")
		return
	}

	if task.Status != DONE {
		context.AbortWithStatusJSON(http.StatusNotFound, "Task not done")
		return
	}

	context.Writer.Header().Set("Content-Type", "application/zip")
	context.File(taskId + "/output" + "/result.zip")
}

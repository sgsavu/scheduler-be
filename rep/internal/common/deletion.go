package common

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func purgeTask(taskId string, tasks map[string]Task) {
	delete(tasks, taskId)

	err := os.RemoveAll(TASKS_DIR + "/" + taskId)
	if err != nil {
		fmt.Println(err)
	}
}

func PeriodicPurge(tasks map[string]Task) {
	for {
		for taskId, task := range tasks {
			if task.Status == DONE {
				purgeTask(taskId, tasks)
			}
		}
		time.Sleep(getPeriodicPurgeInterval())
	}
}

func DeleteTask(context *gin.Context, tasks map[string]Task) {
	taskId := context.Param("id")

	task, exists := tasks[taskId]

	if !exists {
		context.AbortWithStatusJSON(http.StatusNotFound, "Task not found")
		return
	}

	if task.Status == WORKING {
		err := task.Process.Kill()
		if err != nil {
			fmt.Println("failed to kill process: ", err)
		}
		task.Status = CANCELLED
		tasks[taskId] = task
		context.JSON(http.StatusOK, "Success")
		return
	}

	purgeTask(taskId, tasks)
	context.JSON(http.StatusOK, "Success")
}

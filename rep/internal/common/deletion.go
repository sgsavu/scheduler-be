package common

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func getPeriodicPurge() time.Duration {
	periodicPurge := os.Getenv("PERIODIC_PURGE_INTERVAL")
	Atoi, err := strconv.Atoi(periodicPurge)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return time.Duration(Atoi) * time.Millisecond
}

func purgeTask(taskId string, tasks map[string]Task) {
	delete(tasks, taskId)

	err := os.RemoveAll(taskId)
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
		time.Sleep(getPeriodicPurge())
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

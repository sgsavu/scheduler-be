package common

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupTerminationRoutine(cmd *exec.Cmd, taskId string, tasks map[string]Task) {
	err := cmd.Wait()

	timeNow := time.Now()
	expiryTime := timeNow.Add(getPeriodicPurge())
	task := tasks[taskId]
	task.TerminationTime = &timeNow
	task.ExpiryTime = &expiryTime

	if task.Status == CANCELLED {
		purgeTask(taskId, tasks)
		return
	}

	if err != nil {
		task.Status = FAILED
	} else {
		task.Status = DONE
		ZipAndCleanDirectory(taskId + "/output")
	}

	tasks[taskId] = task

	for _, context := range task.Listeners {
		data, _ := json.Marshal(gin.H{task.ID: task})
		sendEvent(context, "onChange", data)
	}

	if task.Status == FAILED {
		purgeTask(taskId, tasks)
		return
	}

	removalErr := os.RemoveAll(taskId + "/input")
	if removalErr != nil {
		fmt.Println(removalErr)
	}
}

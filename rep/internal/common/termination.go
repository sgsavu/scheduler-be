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
	expiryTime := timeNow.Add(getPeriodicPurgeInterval())
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
		ZipAndCleanDirectory(TASKS_DIR + "/" + taskId + "/" + TASK_OUTPUT)
	}

	tasks[taskId] = task

	data, _ := json.Marshal(gin.H{task.ID: task})

	sse := TaskSSE{
		ID:    taskId,
		Event: "onChange",
		Data:  data,
	}

	task.Channel <- sse

	if task.Status == FAILED {
		purgeTask(taskId, tasks)
		return
	}

	removalErr := os.RemoveAll(TASKS_DIR + "/" + taskId + "/" + TASK_INPUT)
	if removalErr != nil {
		fmt.Println(removalErr)
	}

	if task.Type == INFERRING {
		removalErr = os.RemoveAll(TASKS_DIR + "/" + taskId + "/" + TASK_MODEL)
		if removalErr != nil {
			fmt.Println(removalErr)
		}
	}
}

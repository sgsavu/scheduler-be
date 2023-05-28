package common

import (
	"bufio"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
)

func HandleCmdErrors(pipe io.ReadCloser, id string, tasks map[string]Task) {
	readBytes, _ := io.ReadAll(pipe)
	task := tasks[id]
	task.FailureReason = string(readBytes)
	tasks[id] = task
}

func HandleCmdOutput(pipe io.ReadCloser, taskId string, channel chan TaskSSE) {
	reader := bufio.NewReader(pipe)

	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			return
		}

		data, _ := json.Marshal(gin.H{taskId: string(line)})

		taskSSE := TaskSSE{
			ID:    taskId,
			Event: "onStatus",
			Data:  data,
		}

		channel <- taskSSE
	}
}

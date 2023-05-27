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

func HandleCmdOutput(pipe io.ReadCloser, taskId string, listeners map[string]*gin.Context) {
	reader := bufio.NewReader(pipe)

	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			return
		}

		for _, context := range listeners {
			data, _ := json.Marshal(gin.H{taskId: string(line)})
			sendEvent(context, "onStatus", data)
		}
	}
}

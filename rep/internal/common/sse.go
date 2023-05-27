package common

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var eventId = 0

func establishSSE(context *gin.Context) {
	context.Writer.Header().Set("Content-Type", "text/event-stream")
	context.Writer.Header().Set("Cache-Control", "no-cache")
	context.Writer.Flush()
}

func sendEvent(context *gin.Context, eventName string, data []byte) {
	context.Writer.Write([]byte(fmt.Sprintf("id: %d\n", eventId)))
	context.Writer.Write([]byte("event: " + eventName + "\n"))
	context.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", data)))
	context.Writer.Flush()
	eventId++
}

func SubToAll(context *gin.Context, tasks map[string]Task) {
	establishSSE(context)

	connId := uuid.New().String()

	for _, task := range tasks {
		task.Listeners[connId] = context
	}

	data, _ := json.Marshal(tasks)
	sendEvent(context, "onChange", data)

	for {
		select {
		case <-context.Request.Context().Done():
			for _, task := range tasks {
				delete(task.Listeners, connId)
			}

			return
		}
	}
}

func SubToTask(context *gin.Context, tasks map[string]Task) {
	taskId := context.Param("id")

	task, exists := tasks[taskId]

	if !exists {
		context.AbortWithStatusJSON(http.StatusNotFound, "Task not found")
		return
	}

	establishSSE(context)

	connId := uuid.New().String()
	subscribersMap := task.Listeners
	subscribersMap[connId] = context

	for {
		select {
		case <-context.Request.Context().Done():
			delete(subscribersMap, connId)
			return
		}
	}
}

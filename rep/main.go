package main

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TaskType string
type TaskStatus string

const (
	INFERRING TaskType = "INFERRING"
	TRAINING  TaskType = "TRAINING"
)

const (
	WORKING TaskStatus = "WORKING"
	DONE    TaskStatus = "DONE"
	FAILED  TaskStatus = "FAILED"
)

type InferringTaskPayload struct {
	Model      []*multipart.FileHeader `form:"model" binding:"required"`
	Input      []*multipart.FileHeader `form:"input" binding:"required"`
	IndexRatio int                     `form:"indexRatio" binding:"required"`
	Pitch      int                     `form:"pitch" binding:"required"`
}

type TrainingTaskCommon struct {
	BatchSize  int    `form:"batchSize" binding:"required"`
	Epochs     int    `form:"epochs" binding:"required"`
	ModelName  string `form:"modelName" binding:"required"`
	SampleRate int    `form:"sampleRate" binding:"required"`
}

type TrainingTaskPayload struct {
	TrainingTaskCommon
	Dataset []*multipart.FileHeader `form:"dataset" binding:"required"`
}

type TrainingTask struct {
	TaskListenersExt
	Type            TaskType
	CreationTime    time.Time
	TerminationTime *time.Time `json:",omitempty"`
	Status          TaskStatus
	TrainingTaskCommon
	Process *os.Process `json:"-"`
}

type TaskListenersExt struct {
	ID        string
	Listeners map[string]*gin.Context `json:"-"`
}

const PERIODIC_PURGE_INTERVAL = 24 * time.Hour

var tasks = make(map[string]TrainingTask)

func (c TaskListenersExt) Write(p []byte) (int, error) {
	// need to find a way to remove this shit from python directly
	noN := strings.ReplaceAll(string(p), "\n", "")
	noR := strings.ReplaceAll(noN, "\r", "")

	for _, context := range c.Listeners {
		context.Writer.Write([]byte(fmt.Sprintf("id: %s\n", c.ID)))
		context.Writer.Write([]byte("event: onStatus\n"))
		context.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", noR)))
		context.Writer.Flush()
	}

	return len(p), nil
}

func trainPipe(c *gin.Context) {
	var trainingTaskPayload TrainingTaskPayload

	if parseRequestError := c.ShouldBind(&trainingTaskPayload); parseRequestError != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, parseRequestError.Error())
		return
	}

	taskId := uuid.New().String()
	dataset := trainingTaskPayload.Dataset

	for index, file := range dataset {
		fileExtension := filepath.Ext(file.Filename)

		if fileExtension != ".wav" {
			continue
		}

		savePath := fmt.Sprintf("%s/%d%s", taskId, index, fileExtension)

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
			return
		}
	}

	tasks[taskId] = TrainingTask{
		CreationTime: time.Now(),
		Status:       WORKING,
		TaskListenersExt: TaskListenersExt{
			ID:        taskId,
			Listeners: make(map[string]*gin.Context),
		},
		TrainingTaskCommon: trainingTaskPayload.TrainingTaskCommon,
		Type:               TRAINING,
	}

	cmd := exec.Command("python", "../scripts/test.py")
	cmd.Stdout = TaskListenersExt{
		ID:        tasks[taskId].ID,
		Listeners: tasks[taskId].Listeners,
	}

	if startCommandError := cmd.Start(); startCommandError != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, startCommandError.Error())
		return
	}

	task := tasks[taskId]
	task.Process = cmd.Process
	tasks[taskId] = task

	c.JSON(http.StatusOK, taskId)

	time.AfterFunc(1*time.Second, func() {
		err := cmd.Wait()

		timeNow := time.Now()
		task := tasks[taskId]
		task.TerminationTime = &timeNow

		if err != nil {
			fmt.Println("Error", err)
			task.Status = FAILED
		} else {
			task.Status = DONE
		}

		tasks[taskId] = task

		for _, context := range task.Listeners {

			context.Writer.Write([]byte(fmt.Sprintf("id: %s\n", task.ID)))
			context.Writer.Write([]byte("event: onChange\n"))
			data, _ := json.Marshal(gin.H{task.ID: task})
			context.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", data)))
			context.Writer.Flush()
		}
	})
}

func inferPipe(c *gin.Context) {
	c.JSON(http.StatusOK, "Success")
}

func subToAll(context *gin.Context) {
	context.Writer.Header().Set("Content-Type", "text/event-stream")
	context.Writer.Header().Set("Cache-Control", "no-cache")
	context.Writer.Flush()

	connId := uuid.New().String()

	for _, task := range tasks {
		subscribersMap := task.Listeners
		subscribersMap[connId] = context
	}

	context.Writer.Write([]byte(fmt.Sprintf("id: %s\n", "initial")))
	context.Writer.Write([]byte("event: onChange\n"))
	data, _ := json.Marshal(tasks)
	context.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", data)))
	context.Writer.Flush()

	for {
		select {
		case <-context.Request.Context().Done():
			for _, task := range tasks {
				subscribersMap := task.Listeners
				delete(subscribersMap, connId)

			}
			return
		}
	}
}

func subToTask(context *gin.Context) {
	taskId := context.Param("id")

	task, exists := tasks[taskId]

	if !exists {
		context.AbortWithStatusJSON(http.StatusNotFound, "Task not found")
		return
	}

	context.Writer.Header().Set("Content-Type", "text/event-stream")
	context.Writer.Header().Set("Cache-Control", "no-cache")
	context.Writer.Flush()

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

func getResult(context *gin.Context) {
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

	context.File(taskId + "/result.zip")
}

func purgeTask(taskId string) {
	delete(tasks, taskId)

	err := os.RemoveAll(taskId)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Directory", taskId, "removed successfully")
	}

	fmt.Println("purgeTask", tasks)
}

func periodicPurge() {
	for {
		for taskId, task := range tasks {
			if task.Status == DONE || task.Status == FAILED {
				purgeTask(taskId)
			}
		}
		time.Sleep(PERIODIC_PURGE_INTERVAL)
	}
}

func deleteTask(context *gin.Context) {
	taskId := context.Param("id")

	task, exists := tasks[taskId]

	if !exists {
		context.AbortWithStatusJSON(http.StatusNotFound, "Task not found")
		return
	}

	if task.Status == WORKING {
		if err := task.Process.Kill(); err != nil {
			fmt.Println("failed to kill process: ", err)
		}
	}

	purgeTask(taskId)
}

func main() {
	router := gin.Default()

	v1 := router.Group("/v1")
	{
		tasks := v1.Group("/tasks")
		{
			tasks.GET("", subToAll)
			tasks.GET("/:id")
			tasks.DELETE("/:id", deleteTask)
			tasks.GET("/:id/status", subToTask)
			tasks.GET("/:id/result", getResult)
		}

		infer := v1.Group("/infer")
		{
			infer.POST("/", inferPipe)
		}

		train := v1.Group("/train")
		{
			train.POST("/", trainPipe)
		}
	}

	go periodicPurge()

	router.Use(static.Serve("/", static.LocalFile("../ui/dist", false)))
	router.Run()
}

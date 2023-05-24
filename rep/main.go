package main

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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
	CANCELLED TaskStatus = "CANCELLED"
	DONE      TaskStatus = "DONE"
	FAILED    TaskStatus = "FAILED"
	WORKING   TaskStatus = "WORKING"
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

type Task struct {
	ID              string
	Listeners       map[string]*gin.Context `json:"-"`
	Type            TaskType
	CreationTime    time.Time
	TerminationTime *time.Time `json:",omitempty"`
	Status          TaskStatus
	TrainingTaskCommon
	Process       *os.Process `json:"-"`
	FailureReason string      `json:",omitempty"`
}

const PERIODIC_PURGE_INTERVAL = 24 * time.Hour

var tasks = make(map[string]Task)

func handleCmdErrors(s io.ReadCloser, id string) {
	readBytes, _ := io.ReadAll(s)
	task := tasks[id]
	task.FailureReason = string(readBytes)
	tasks[id] = task
}

func handleCmdOutput(pipe io.ReadCloser, taskId string, listeners map[string]*gin.Context) {
	reader := bufio.NewReader(pipe)

	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			return
		}

		for _, context := range listeners {
			context.Writer.Write([]byte(fmt.Sprintf("id: %s\n", taskId)))
			context.Writer.Write([]byte("event: onStatus\n"))
			context.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", string(line))))
			context.Writer.Flush()
		}
	}
}

func zipResult(taskId string) {
	outputPath := taskId + "/output"

	files, err := ioutil.ReadDir(outputPath)
	if err != nil {
		panic(err)
	}

	archive, err := os.Create(outputPath + "/result.zip")
	if err != nil {
		panic(err)
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)

	for _, file := range files {
		filePath := outputPath + "/" + file.Name()

		f1, err := os.Open(filePath)
		if err != nil {
			panic(err)
		}

		w1, err := zipWriter.Create(file.Name())
		if err != nil {
			panic(err)
		}
		if _, err := io.Copy(w1, f1); err != nil {
			panic(err)
		}

		f1.Close()
		os.Remove(filePath)
	}

	zipWriter.Close()
}

func trainPipe(c *gin.Context) {
	var trainingTaskPayload TrainingTaskPayload

	parseRequestError := c.ShouldBind(&trainingTaskPayload)
	if parseRequestError != nil {
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

		savePath := fmt.Sprintf("%s/input/%d%s", taskId, index, fileExtension)

		err := c.SaveUploadedFile(file, savePath)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
			return
		}
	}

	cmd := exec.Command("python", "../scripts/test.py", taskId)
	stdout, stdOutErr := cmd.StdoutPipe()
	if stdOutErr != nil {
		fmt.Println(stdOutErr)
	}
	stderr, stdErrErr := cmd.StderrPipe()
	if stdErrErr != nil {
		fmt.Println(stdErrErr)
	}

	tasks[taskId] = Task{
		CreationTime:       time.Now(),
		Status:             WORKING,
		ID:                 taskId,
		Listeners:          make(map[string]*gin.Context),
		TrainingTaskCommon: trainingTaskPayload.TrainingTaskCommon,
		Type:               TRAINING,
	}

	go handleCmdOutput(stdout, taskId, tasks[taskId].Listeners)
	go handleCmdErrors(stderr, taskId)

	startCommandError := cmd.Start()
	if startCommandError != nil {
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

		if task.Status == CANCELLED {
			purgeTask(taskId)
			return
		}

		if err != nil {
			task.Status = FAILED
		} else {
			task.Status = DONE
			zipResult(taskId)
		}

		tasks[taskId] = task

		for _, context := range task.Listeners {
			context.Writer.Write([]byte(fmt.Sprintf("id: %s\n", task.ID)))
			context.Writer.Write([]byte("event: onChange\n"))
			data, _ := json.Marshal(gin.H{task.ID: task})
			context.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", data)))
			context.Writer.Flush()
		}

		if task.Status == FAILED {
			purgeTask(taskId)
			return
		}

		removalErr := os.RemoveAll(taskId + "/input")
		if removalErr != nil {
			fmt.Println(err)
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

	context.Writer.Header().Set("Content-Type", "application/zip")
	context.File(taskId + "/output" + "/result.zip")
}

func purgeTask(taskId string) {
	delete(tasks, taskId)

	err := os.RemoveAll(taskId)
	if err != nil {
		fmt.Println(err)
	}
}

func periodicPurge() {
	for {
		for taskId, task := range tasks {
			if task.Status == DONE {
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
		err := task.Process.Kill()
		if err != nil {
			fmt.Println("failed to kill process: ", err)
		}
		task.Status = CANCELLED
		tasks[taskId] = task
		context.JSON(http.StatusOK, "Success")
		return
	}

	purgeTask(taskId)
	context.JSON(http.StatusOK, "Success")
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

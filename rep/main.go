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

type CommonTaskPayload struct {
	Name string `json:"name,omitempty" form:"name" binding:"required"`
}

type InferringTaskCommon struct {
	IndexRatio float32 `json:",omitempty" form:"indexRatio" binding:"required"`
	Pitch      int     `json:",omitempty" form:"pitch" binding:"required"`
}

type InferringTaskPayload struct {
	InferringTaskCommon
	Model []*multipart.FileHeader `form:"model" binding:"required"`
	Input []*multipart.FileHeader `form:"input" binding:"required"`
	CommonTaskPayload
}

type TrainingTaskCommon struct {
	BatchSize  int `json:",omitempty" form:"batchSize" binding:"required"`
	Epochs     int `json:",omitempty" form:"epochs" binding:"required"`
	SampleRate int `json:",omitempty" form:"sampleRate" binding:"required"`
}

type TrainingTaskPayload struct {
	TrainingTaskCommon
	Dataset []*multipart.FileHeader `form:"dataset" binding:"required"`
	CommonTaskPayload
}

type Task struct {
	ID              string
	Listeners       map[string]*gin.Context `json:"-"`
	Type            TaskType
	CreationTime    time.Time
	TerminationTime *time.Time `json:",omitempty"`
	ExpiryTime      *time.Time `json:",omitempty"`
	Status          TaskStatus
	Process         *os.Process `json:"-"`
	FailureReason   string      `json:",omitempty"`
	Name            string      `json:",omitempty" form:"name" binding:"required"`
	TrainingTaskCommon
	InferringTaskCommon
}

const PERIODIC_PURGE_INTERVAL = 1 * time.Minute

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

func handleCmdErrors(pipe io.ReadCloser, id string, tasks map[string]Task) {
	readBytes, _ := io.ReadAll(pipe)
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
			sendEvent(context, "onStatus", line)
		}
	}
}

func zipAndCleanDirectory(path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}

	archive, err := os.Create(path + "/result.zip")
	if err != nil {
		panic(err)
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)

	for _, file := range files {
		filePath := path + "/" + file.Name()

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

func setupTerminationRoutine(cmd *exec.Cmd, taskId string, tasks map[string]Task) {
	err := cmd.Wait()

	timeNow := time.Now()
	expiryTime := timeNow.Add(PERIODIC_PURGE_INTERVAL)
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
		zipAndCleanDirectory(taskId + "/output")
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

func saveFormFiles(c *gin.Context, dirPath string, files []*multipart.FileHeader) error {
	for index, file := range files {
		fileExtension := filepath.Ext(file.Filename)
		if fileExtension != ".wav" {
			continue
		}

		savePath := fmt.Sprintf("%s/%d%s", dirPath, index, fileExtension)

		err := c.SaveUploadedFile(file, savePath)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
			return err
		}
	}

	return nil
}

func trainPipe(context *gin.Context, tasks map[string]Task) {
	var trainTaskPayload TrainingTaskPayload

	parseRequestError := context.ShouldBind(&trainTaskPayload)
	if parseRequestError != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, parseRequestError.Error())
		return
	}

	taskId := uuid.New().String()
	dataset := trainTaskPayload.Dataset

	saveFormFiles(context, taskId+"/input", dataset)

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
		ID:                 taskId,
		Listeners:          make(map[string]*gin.Context),
		CreationTime:       time.Now(),
		Status:             WORKING,
		Type:               TRAINING,
		Name:               trainTaskPayload.Name,
		TrainingTaskCommon: trainTaskPayload.TrainingTaskCommon,
	}

	go handleCmdOutput(stdout, taskId, tasks[taskId].Listeners)
	go handleCmdErrors(stderr, taskId, tasks)

	startCommandError := cmd.Start()
	if startCommandError != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, startCommandError.Error())
		return
	}

	task := tasks[taskId]
	task.Process = cmd.Process
	tasks[taskId] = task

	go setupTerminationRoutine(cmd, taskId, tasks)

	context.JSON(http.StatusOK, taskId)
}

func inferPipe(context *gin.Context, tasks map[string]Task) {
	var inferTaskPayload InferringTaskPayload

	parseRequestError := context.ShouldBind(&inferTaskPayload)
	if parseRequestError != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, parseRequestError.Error())
		return
	}

	taskId := uuid.New().String()
	dataset := inferTaskPayload.Input

	saveFormFiles(context, taskId+"/input", dataset)

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
		ID:                  taskId,
		Listeners:           make(map[string]*gin.Context),
		CreationTime:        time.Now(),
		Status:              WORKING,
		Type:                INFERRING,
		Name:                inferTaskPayload.Name,
		InferringTaskCommon: inferTaskPayload.InferringTaskCommon,
	}

	go handleCmdOutput(stdout, taskId, tasks[taskId].Listeners)
	go handleCmdErrors(stderr, taskId, tasks)

	startCommandError := cmd.Start()
	if startCommandError != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, startCommandError.Error())
		return
	}

	task := tasks[taskId]
	task.Process = cmd.Process
	tasks[taskId] = task

	go setupTerminationRoutine(cmd, taskId, tasks)

	context.JSON(http.StatusOK, taskId)
}

func subToAll(context *gin.Context, tasks map[string]Task) {
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

func subToTask(context *gin.Context, tasks map[string]Task) {
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

func getResult(context *gin.Context, tasks map[string]Task) {
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

func purgeTask(taskId string, tasks map[string]Task) {
	delete(tasks, taskId)

	err := os.RemoveAll(taskId)
	if err != nil {
		fmt.Println(err)
	}
}

func periodicPurge(tasks map[string]Task) {
	for {
		for taskId, task := range tasks {
			if task.Status == DONE {
				purgeTask(taskId, tasks)
			}
		}
		time.Sleep(PERIODIC_PURGE_INTERVAL)
	}
}

func deleteTask(context *gin.Context, tasks map[string]Task) {
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

func main() {
	var tasks = make(map[string]Task)

	router := gin.Default()

	v1 := router.Group("/v1")
	{
		tasksGroup := v1.Group("/tasks")
		{
			tasksGroup.GET("", func(ctx *gin.Context) {
				subToAll(ctx, tasks)
			})
			tasksGroup.GET("/:id")
			tasksGroup.DELETE("/:id", func(ctx *gin.Context) {
				deleteTask(ctx, tasks)
			})
			tasksGroup.GET("/:id/status", func(ctx *gin.Context) {
				subToTask(ctx, tasks)
			})
			tasksGroup.GET("/:id/result", func(ctx *gin.Context) {
				getResult(ctx, tasks)
			})
		}

		infer := v1.Group("/infer")
		{
			infer.POST("", func(ctx *gin.Context) {
				inferPipe(ctx, tasks)
			})
		}

		train := v1.Group("/train")
		{
			train.POST("", func(ctx *gin.Context) {
				trainPipe(ctx, tasks)
			})
		}
	}

	go periodicPurge(tasks)

	router.Use(static.Serve("/", static.LocalFile("../ui/dist", false)))
	router.Run()
}

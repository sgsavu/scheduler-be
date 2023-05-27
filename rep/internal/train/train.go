package train

import (
	"fmt"
	"net/http"
	"os/exec"
	"rep/internal/common"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TrainPipe(context *gin.Context, tasks map[string]common.Task) {
	var trainTaskPayload common.TrainingTaskPayload

	parseRequestError := context.ShouldBind(&trainTaskPayload)
	if parseRequestError != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, parseRequestError.Error())
		return
	}

	taskId := uuid.New().String()
	dataset := trainTaskPayload.Dataset

	common.SaveFormFiles(context, taskId+"/input", dataset)

	cmd := exec.Command("python", "../scripts/test.py", taskId)
	stdout, stdOutErr := cmd.StdoutPipe()
	if stdOutErr != nil {
		fmt.Println(stdOutErr)
	}
	stderr, stdErrErr := cmd.StderrPipe()
	if stdErrErr != nil {
		fmt.Println(stdErrErr)
	}

	tasks[taskId] = common.Task{
		ID:                 taskId,
		Listeners:          make(map[string]*gin.Context),
		CreationTime:       time.Now(),
		Status:             common.WORKING,
		Type:               common.TRAINING,
		Name:               trainTaskPayload.Name,
		TrainingTaskCommon: trainTaskPayload.TrainingTaskCommon,
	}

	go common.HandleCmdOutput(stdout, taskId, tasks[taskId].Listeners)
	go common.HandleCmdErrors(stderr, taskId, tasks)

	startCommandError := cmd.Start()
	if startCommandError != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, startCommandError.Error())
		return
	}

	task := tasks[taskId]
	task.Process = cmd.Process
	tasks[taskId] = task

	go common.SetupTerminationRoutine(cmd, taskId, tasks)

	context.JSON(http.StatusOK, taskId)
}

package infer

import (
	"fmt"
	"net/http"
	"os/exec"
	"rep/internal/common"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func InferPipe(context *gin.Context, tasks map[string]common.Task) {
	var inferTaskPayload common.InferringTaskPayload

	parseRequestError := context.ShouldBind(&inferTaskPayload)
	if parseRequestError != nil {
		context.AbortWithStatusJSON(http.StatusInternalServerError, parseRequestError.Error())
		return
	}

	taskId := uuid.New().String()
	input := inferTaskPayload.Input
	model := inferTaskPayload.Model

	errSavingInput := common.SaveFormFiles(context, common.TASKS_DIR+"/"+taskId+"/"+common.TASK_INPUT, input, common.ALLOWED_AUDIO_FORMATS)
	if errSavingInput != nil {
		return
	}
	errSavingModel := common.SaveFormFiles(context, common.TASKS_DIR+"/"+taskId+"/"+common.TASK_MODEL, model, common.ALLOWED_MODEL_FORMATS)
	if errSavingModel != nil {
		return
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

	tasks[taskId] = common.Task{
		ID:                  taskId,
		Listeners:           make(map[string]*gin.Context),
		CreationTime:        time.Now(),
		Status:              common.WORKING,
		Type:                common.INFERRING,
		Name:                inferTaskPayload.Name,
		InferringTaskCommon: inferTaskPayload.InferringTaskCommon,
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

package infer

import (
	"fmt"
	"os/exec"
	"rep/internal/common"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func InferPipe(context *fiber.Ctx, tasks map[string]common.Task) {
	var inferTaskPayload common.InferringTaskPayload

	parseRequestError := context.BodyParser(&inferTaskPayload)
	if parseRequestError != nil {
		context.SendStatus(500)
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
		Channel:             make(chan common.TaskSSE),
		CreationTime:        time.Now(),
		Status:              common.WORKING,
		Type:                common.INFERRING,
		Name:                inferTaskPayload.Name,
		InferringTaskCommon: inferTaskPayload.InferringTaskCommon,
	}

	go common.HandleCmdOutput(stdout, taskId, tasks[taskId].Channel)
	go common.HandleCmdErrors(stderr, taskId, tasks)

	startCommandError := cmd.Start()
	if startCommandError != nil {
		context.SendStatus(500)
		return
	}

	task := tasks[taskId]
	task.Process = cmd.Process
	tasks[taskId] = task

	go common.SetupTerminationRoutine(cmd, taskId, tasks)

	context.Status(200)
	context.SendString(taskId)
}

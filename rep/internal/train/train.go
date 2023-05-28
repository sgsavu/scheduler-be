package train

import (
	"fmt"
	"os/exec"
	"rep/internal/common"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func TrainPipe(context *fiber.Ctx, tasks map[string]common.Task) {

	var trainTaskPayload common.TrainingTaskPayload

	parseRequestError := context.BodyParser(&trainTaskPayload)
	if parseRequestError != nil {
		context.SendStatus(500)
		return
	}

	form, err := context.MultipartForm()
	if err != nil {
		context.SendStatus(500)
		return
	}

	dataset := form.File["dataset"]

	taskId := uuid.New().String()

	if len(dataset) == 0 {
		context.SendStatus(400)
		return
	}

	err2 := common.SaveFormFiles(context, common.TASKS_DIR+"/"+taskId+"/"+common.TASK_INPUT, dataset, common.ALLOWED_AUDIO_FORMATS)
	if err2 != nil {
		context.SendStatus(500)
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
		ID:                 taskId,
		Channel:            make(chan common.TaskSSE),
		CreationTime:       time.Now(),
		Status:             common.WORKING,
		Type:               common.TRAINING,
		Name:               trainTaskPayload.Name,
		TrainingTaskCommon: trainTaskPayload.TrainingTaskCommon,
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

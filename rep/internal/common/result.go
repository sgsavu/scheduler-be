package common

import (
	"github.com/gofiber/fiber/v2"
)

func GetResult(context *fiber.Ctx, tasks map[string]Task) {
	taskId := context.Params("id")

	task, exists := tasks[taskId]

	if !exists || task.Status != DONE {
		context.SendStatus(404)
		return
	}

	context.Set("Content-Type", "application/zip")
	context.Download(TASKS_DIR + "/" + taskId + "/" + TASK_OUTPUT + "/result.zip")
}

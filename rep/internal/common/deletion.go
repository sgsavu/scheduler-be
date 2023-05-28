package common

import (
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

func purgeTask(taskId string, tasks map[string]Task) {
	delete(tasks, taskId)

	err := os.RemoveAll(TASKS_DIR + "/" + taskId)
	if err != nil {
		fmt.Println(err)
	}
}

func PeriodicPurge(tasks map[string]Task) {
	for {
		for taskId, task := range tasks {
			if task.Status == DONE {
				purgeTask(taskId, tasks)
			}
		}
		time.Sleep(getPeriodicPurgeInterval())
	}
}

func DeleteTask(context *fiber.Ctx, tasks map[string]Task) {
	taskId := context.Params("id")

	task, exists := tasks[taskId]

	if !exists {
		context.SendStatus(404)
		return
	}

	if task.Status == WORKING {
		err := task.Process.Kill()
		if err != nil {
			fmt.Println("failed to kill process: ", err)
		}
		task.Status = CANCELLED
		tasks[taskId] = task
		context.SendStatus(200)
		return
	}

	purgeTask(taskId, tasks)
	context.SendStatus(200)
}

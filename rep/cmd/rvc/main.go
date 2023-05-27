package main

import (
	common "rep/internal/common"
	infer "rep/internal/infer"
	train "rep/internal/train"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func main() {
	var tasks = make(map[string]common.Task)

	router := gin.Default()

	v1 := router.Group("/v1")
	{
		tasksGroup := v1.Group("/tasks")
		{
			tasksGroup.GET("", func(ctx *gin.Context) {
				common.SubToAll(ctx, tasks)
			})
			tasksGroup.GET("/:id")
			tasksGroup.DELETE("/:id", func(ctx *gin.Context) {
				common.DeleteTask(ctx, tasks)
			})
			tasksGroup.GET("/:id/status", func(ctx *gin.Context) {
				common.SubToTask(ctx, tasks)
			})
			tasksGroup.GET("/:id/result", func(ctx *gin.Context) {
				common.GetResult(ctx, tasks)
			})
		}

		inferGroup := v1.Group("/infer")
		{
			inferGroup.POST("", func(ctx *gin.Context) {
				infer.InferPipe(ctx, tasks)
			})
		}

		trainGroup := v1.Group("/train")
		{
			trainGroup.POST("", func(ctx *gin.Context) {
				train.TrainPipe(ctx, tasks)
			})
		}
	}

	common.LoadEnvVars()

	go common.PeriodicPurge(tasks)

	router.Use(static.Serve("/", static.LocalFile("../ui/dist", false)))
	router.Run()
}

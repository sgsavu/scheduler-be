package main

import (
	common "rep/internal/common"
	"rep/internal/infer"
	"rep/internal/train"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	var tasks = make(map[string]common.Task)

	common.LoadEnvVars()
	go common.PeriodicPurge(tasks)

	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
		BodyLimit:   100 * 1024 * 1024,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Cache-Control",
		AllowCredentials: true,
	}))

	app.Static("/", "../ui/dist")

	app.Route("/v1", func(app fiber.Router) {
		app.Post("/infer", func(ctx *fiber.Ctx) error {
			infer.InferPipe(ctx, tasks)
			return nil
		})

		app.Post("/train", func(ctx *fiber.Ctx) error {
			train.TrainPipe(ctx, tasks)
			return nil
		})

		app.Route("/tasks", func(app fiber.Router) {
			app.Get("", func(ctx *fiber.Ctx) error {
				common.SubToAll(ctx, tasks)
				return nil
			})

			// app.Get("/:id", func(ctx *fiber.Ctx) error {
			// 	common.GetTask(ctx, tasks)
			// 	return nil
			// })

			app.Delete("/:id", func(ctx *fiber.Ctx) error {
				common.DeleteTask(ctx, tasks)
				return nil
			})

			// app.Get("/:id/status", func(ctx *fiber.Ctx) error {
			// 	common.SubToTask(ctx, tasks)
			// 	return nil
			// })

			app.Get("/:id/result", func(ctx *fiber.Ctx) error {
				common.GetResult(ctx, tasks)
				return nil
			})
		})
	})

	app.Listen(":8080")
}

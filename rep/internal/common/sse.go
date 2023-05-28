package common

import (
	"bufio"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

func establishSSE(context *fiber.Ctx) {
	context.Set("Content-Type", "text/event-stream")
	context.Set("Cache-Control", "no-cache")
	context.Set("Connection", "keep-alive")
	context.Set("Transfer-Encoding", "chunked")
}

func forward(master chan TaskSSE, child chan TaskSSE) {
	for {
		taskSSE := <-child
		master <- taskSSE
	}
}

func SubToAll(context *fiber.Ctx, tasks map[string]Task) {
	establishSSE(context)

	masterChan := make(chan TaskSSE)

	for _, task := range tasks {
		go forward(masterChan, task.Channel)
	}

	context.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		var init = true
		for {

			if init {

				data, _ := json.Marshal(tasks)

				fmt.Fprintf(w, "id: %d\n", 1)
				fmt.Fprintf(w, "event: %s\n", "onChange")
				fmt.Fprintf(w, "data: %s\n\n", data)
				w.Flush()
				init = false
				continue
			}

			taskSSE := <-masterChan
			fmt.Fprintf(w, "id: %s\n", taskSSE.ID)
			fmt.Fprintf(w, "event: %s\n", taskSSE.Event)
			fmt.Fprintf(w, "data: %s\n\n", taskSSE.Data)
			err := w.Flush()
			if err != nil {
				fmt.Printf("Error while flushing: %v. Closing http connection.\n", err)
				break
			}
		}
	}))
}

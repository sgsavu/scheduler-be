package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func handleCmdErrors(pipe io.ReadCloser) {
	readBytes, _ := io.ReadAll(pipe)
	fmt.Println(readBytes)
}

func handleCmdOutput(pipe io.ReadCloser) {
	reader := bufio.NewReader(pipe)

	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			return
		}

		fmt.Println(line)
		// for _, context := range listeners {
		// 	// data, _ := json.Marshal(gin.H{taskId: string(line)})
		// }
	}
}

func startTask(command string) (*os.Process, error) {
	commandSplit := strings.Split(command, " ")
	cmd := exec.Command(commandSplit[0], commandSplit[1:]...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to grab stdout pipe - %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to grab stderr pipe - %v", err)
	}

	go handleCmdOutput(stdout)
	go handleCmdErrors(stderr)

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start task - %v", err)
	}

	return cmd.Process, nil
}

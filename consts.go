package main

import (
	"time"
)

type TaskStatus string

const (
	ON  TaskStatus = "ON"
	OFF TaskStatus = "OFF"
)

type Task struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Status    TaskStatus `json:"status"`
	Command   string     `json:"command"`
	CreatedAt time.Time  `json:"creationTime"`
}

type CreateTaskRequestBody struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

type TaskUpdate struct {
	Name    string     `json:"name,omitempty"`
	Status  TaskStatus `json:"status,omitempty"`
	Command TaskStatus `json:"command,omitempty"`
}

const TASKS_DIR = "tasks"

const MAX_UPLOAD_SIZE = 2147483648 // 2GB
const PURGE_INTERVAL = 24 * time.Hour

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"scheduler/utils"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool
var runningTasks = map[string]*os.Process{}

const getTasksQuery = "SELECT * FROM tasks"
const getTaskQuery = "SELECT * FROM tasks WHERE id = $1"
const createTaskQuery = "INSERT INTO tasks (name, status, command, created_at) VALUES ($1, $2, $3, $4)"
const deleteTaskQuery = "DELETE FROM tasks WHERE id = $1"

func getTasks() ([]Task, error) {
	rows, err := db.Query(context.Background(), getTasksQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query db - %v", err)
	}
	defer rows.Close()

	var tasks = []Task{}
	for rows.Next() {
		var task Task

		err := rows.Scan(
			&task.ID,
			&task.Name,
			&task.Status,
			&task.Command,
			&task.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to extract row - %v", err)
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func handleGetTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := getTasks()
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get tasks - %v", err))
		return
	}

	utils.JSON(w, http.StatusOK, tasks)
}

func createTask(name string, command string) error {
	_, err := db.Exec(context.Background(), createTaskQuery,
		name,
		OFF,
		command,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("cannot update/insert record - %v", err)
	}

	return nil
}

func handleCreateTask(w http.ResponseWriter, r *http.Request) {
	body := new(CreateTaskRequestBody)
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to deserialize json - %v", err))
		return
	}

	err = createTask(body.Name, body.Command)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create task - %v", err))
		return
	}

	utils.JSON(w, http.StatusOK, nil)
}

func getTask(taskID string) (*Task, error) {
	rows, err := db.Query(context.Background(), getTaskQuery, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query db - %v", err)
	}
	defer rows.Close()

	var task = new(Task)
	for rows.Next() {
		err := rows.Scan(
			&task.ID,
			&task.Name,
			&task.Status,
			&task.Command,
			&task.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan record - %v", err)
		}
	}

	return task, nil
}

func handleGetTask(w http.ResponseWriter, r *http.Request) {
	taskId := r.PathValue("id")

	task, err := getTask(taskId)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get task - %v", err))
		return
	}

	utils.JSON(w, http.StatusOK, task)
}

func updateTask(taskID string, update *TaskUpdate) error {
	task, err := getTask(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task - %v", err)
	}

	var params []interface{}
	var updates []string
	counter := 1

	if update.Name != "" {
		updates = append(updates, fmt.Sprintf("name = $%d", counter))
		params = append(params, update.Name)
		counter++
	}
	if update.Status != "" {
		var skip = false

		switch update.Status {
		case OFF:
			if task.Status == OFF {
				skip = true
			}

			process, exists := runningTasks[taskID]
			if !exists {
				return fmt.Errorf("task is not running")
			}

			err := process.Kill()
			if err != nil {
				return fmt.Errorf("failed to kill process - %v", err)
			}

			delete(runningTasks, taskID)
			task.Status = OFF
		case ON:
			if task.Status == ON {
				skip = true
			}

			process, err := startTask(task.Command)
			if err != nil {
				return fmt.Errorf("failed to start process - %v", err)
			}

			runningTasks[taskID] = process
			task.Status = ON
		default:
			return fmt.Errorf("invalid status value value")
		}

		if !skip {
			updates = append(updates, fmt.Sprintf("status = $%d", counter))
			params = append(params, update.Status)
			counter++
		}
	}
	if update.Command != "" {
		if task.Status == ON {
			return fmt.Errorf("cannot change command of running task")
		}
		updates = append(updates, fmt.Sprintf("command = $%d", counter))
		params = append(params, update.Command)
		counter++
	}

	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = $%d", strings.Join(updates, ", "), counter)
	params = append(params, taskID)

	_, err = db.Exec(context.Background(), query, params)
	if err != nil {
		return fmt.Errorf("failed db query - %v", err)
	}

	return nil
}

func handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	taskId := r.PathValue("id")

	body := new(TaskUpdate)
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to deserialize json - %v", err))
		return
	}

	err = updateTask(taskId, body)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to modify task - %v", err))
		return
	}

	utils.JSON(w, http.StatusOK, nil)
}

func deleteTask(taskId string) error {
	_, err := db.Exec(context.Background(), deleteTaskQuery, taskId)
	if err != nil {
		return fmt.Errorf("failed db query - %v", err)
	}

	// if stored on disk, clear all log files
	// err := os.RemoveAll(TASKS_DIR + "/" + taskId)
	// if err != nil {
	// 	fmt.Printf("failed to delete task: %v", err)
	// }

	return nil
}

func handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	taskId := r.PathValue("id")

	task, err := getTask(taskId)
	if err != nil {
		utils.JSONError(w, http.StatusNotFound, "task not found")
		return
	}

	if task.Status == ON {
		utils.JSONError(w, http.StatusNotFound, "cannot delete running task")
		return
	}

	err = deleteTask(taskId)
	if err != nil {
		utils.JSONError(w, http.StatusNotFound, fmt.Sprintf("failed to delete task - %v", err))
		return
	}

	utils.JSON(w, http.StatusOK, nil)
}

func registerEndpoints() {
	router := http.NewServeMux()

	router.HandleFunc("GET /api/v1/tasks", handleGetTasks)
	router.HandleFunc("POST /api/v1/tasks", handleCreateTask)

	router.HandleFunc("PATCH /api/v1/tasks", handleUpdateTask)
	router.HandleFunc("GET /api/v1/tasks/{id}", handleGetTask)
	router.HandleFunc("DELETE /api/v1/tasks/{id}", handleDeleteTask)

	port := 8080
	address := fmt.Sprintf("localhost:%d", port)

	fmt.Printf("server starting on - http://%s\n", address)
	err := http.ListenAndServe(address, router)
	if err != nil {
		fmt.Printf("error starting server - %v\n", err)
		os.Exit(1)
	}
}

func connectToDB() {
	connStr := "postgres://jibreel@localhost:5432/scheduler"

	var err error
	db, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		fmt.Printf("unable to connect to database - %v\n", err)
		os.Exit(1)
	}
}

func main() {
	connectToDB()
	registerEndpoints()
}

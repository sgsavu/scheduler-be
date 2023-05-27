package common

import (
	"mime/multipart"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type TaskType string
type TaskStatus string

const (
	INFERRING TaskType = "INFERRING"
	TRAINING  TaskType = "TRAINING"
)

const (
	CANCELLED TaskStatus = "CANCELLED"
	DONE      TaskStatus = "DONE"
	FAILED    TaskStatus = "FAILED"
	WORKING   TaskStatus = "WORKING"
)

type CommonTaskPayload struct {
	Name string `json:"name,omitempty" form:"name" binding:"required"`
}

type Task struct {
	ID              string
	Listeners       map[string]*gin.Context `json:"-"`
	Type            TaskType
	CreationTime    time.Time
	TerminationTime *time.Time `json:",omitempty"`
	ExpiryTime      *time.Time `json:",omitempty"`
	Status          TaskStatus
	Process         *os.Process `json:"-"`
	FailureReason   string      `json:",omitempty"`
	Name            string      `json:",omitempty" form:"name" binding:"required"`
	TrainingTaskCommon
	InferringTaskCommon
}

type InferringTaskCommon struct {
	IndexRatio float32 `json:",omitempty" form:"indexRatio" binding:"required"`
	Pitch      int     `json:",omitempty" form:"pitch" binding:"required"`
}

type InferringTaskPayload struct {
	InferringTaskCommon
	Model []*multipart.FileHeader `form:"model" binding:"required"`
	Input []*multipart.FileHeader `form:"input" binding:"required"`
	CommonTaskPayload
}

type TrainingTaskCommon struct {
	BatchSize  int `json:",omitempty" form:"batchSize" binding:"required"`
	Epochs     int `json:",omitempty" form:"epochs" binding:"required"`
	SampleRate int `json:",omitempty" form:"sampleRate" binding:"required"`
}

type TrainingTaskPayload struct {
	TrainingTaskCommon
	Dataset []*multipart.FileHeader `form:"dataset" binding:"required"`
	CommonTaskPayload
}

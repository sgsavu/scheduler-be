package common

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func SaveFormFiles(c *gin.Context, dirPath string, files []*multipart.FileHeader) error {
	for index, file := range files {
		fileExtension := filepath.Ext(file.Filename)
		if fileExtension != ".wav" {
			continue
		}

		savePath := fmt.Sprintf("%s/%d%s", dirPath, index, fileExtension)

		err := c.SaveUploadedFile(file, savePath)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
			return err
		}
	}

	return nil
}

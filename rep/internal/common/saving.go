package common

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func contains(allowedExtensions []string, extension string) bool {
	for _, allowedExtension := range allowedExtensions {
		if extension == allowedExtension {
			return true
		}
	}

	return false
}

func SaveFormFiles(c *gin.Context, dirPath string, files []*multipart.FileHeader, allowedExtensions []string) error {
	totalSize := int64(0)

	for index, file := range files {
		fileExtension := filepath.Ext(file.Filename)

		fmt.Println(file.Filename, file.Size)

		if !contains(allowedExtensions, fileExtension) {
			continue
		}

		totalSize += file.Size

		if totalSize > getMaxUploadSize() {
			c.AbortWithStatusJSON(http.StatusBadRequest, "Files too large")
			return gin.Error{}
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

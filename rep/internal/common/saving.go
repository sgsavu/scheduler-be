package common

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

func contains(allowedExtensions []string, extension string) bool {
	for _, allowedExtension := range allowedExtensions {
		if extension == allowedExtension {
			return true
		}
	}

	return false
}

func SaveFormFiles(c *fiber.Ctx, dirPath string, files []*multipart.FileHeader, allowedExtensions []string) error {
	totalSize := int64(0)

	for index, file := range files {
		fileExtension := filepath.Ext(file.Filename)

		if !contains(allowedExtensions, fileExtension) {
			continue
		}

		totalSize += file.Size

		if totalSize > getMaxUploadSize() {
			c.SendStatus(513)
			return &fiber.Error{}
		}

		savePath := fmt.Sprintf("%s/%d%s", dirPath, index, fileExtension)
		os.MkdirAll(dirPath, os.ModePerm)

		err := c.SaveFile(file, savePath)
		if err != nil {
			fmt.Println(err)
			c.SendStatus(500)
			return err
		}
	}

	return nil
}

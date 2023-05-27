package common

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
)

func ZipAndCleanDirectory(path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}

	archive, err := os.Create(path + "/result.zip")
	if err != nil {
		panic(err)
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)

	for _, file := range files {
		filePath := path + "/" + file.Name()

		f1, err := os.Open(filePath)
		if err != nil {
			panic(err)
		}

		w1, err := zipWriter.Create(file.Name())
		if err != nil {
			panic(err)
		}
		if _, err := io.Copy(w1, f1); err != nil {
			panic(err)
		}

		f1.Close()
		os.Remove(filePath)
	}

	zipWriter.Close()
}

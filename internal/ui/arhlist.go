package ui

import (
	"log"
	"os"
	"strings"
)

func SearchDataFiles() ([]string, error) {
	fileList := make([]string, 0)
	arhFiles, err := os.ReadDir("./arh")
	if err == nil {
		for _, f := range arhFiles {
			if strings.HasPrefix(f.Name(), "stop_") {
				fileList = append(fileList, "arh/"+f.Name())
			}
		}
	} else {
		log.Println(err)
	}

	homeFiles, err := os.ReadDir(".")
	if err != nil {
		return fileList, nil
	}

	for _, f := range homeFiles {
		if strings.HasPrefix(f.Name(), "stop_") {
			fileList = append(fileList, f.Name())
		}
	}

	return fileList, nil
}

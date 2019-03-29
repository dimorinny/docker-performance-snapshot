package main

import (
	"os"
	"path/filepath"
)

func createFile(directoryPath string, fileName string) (*os.File, error) {
	if !directoryExists(directoryPath) {
		err := os.MkdirAll(directoryPath, os.ModePerm)
		check(err)
	}

	return os.Create(
		filepath.Join(directoryPath, fileName),
	)
}

func directoryExists(directoryPath string) bool {
	_, err := os.Stat(directoryPath)
	if err != nil {
		return !os.IsNotExist(err)
	}
	return true
}

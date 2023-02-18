package utils

import (
	"bytes"
	"os"
	"strings"
)

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func WriteFile(filePath string, body string) error {
	bbody := []byte(body)
	return os.WriteFile(filePath, bbody, 0644)
}

func FileAppend(filePath string, appendString string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	if !strings.HasSuffix(appendString, "\n") {
		appendString = appendString + "\n"
	}
	_, err = f.WriteString(appendString)
	return err
}

func FileRemoveLine(filePath string, line string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	cleanContent := bytes.Replace(content, []byte(line+"\n"), []byte(""), -1)
	return os.WriteFile(filePath, cleanContent, 0644)
}

func MakeDir(path string) error {
	return os.MkdirAll(path, 0750)
}

func RmDir(path string) error {
	return os.RemoveAll(path)
}

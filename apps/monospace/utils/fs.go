package utils

import (
	"bytes"
	"os"
	"strings"
)

func IsDir(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		return stat.IsDir(), nil
	}
	// not exists is not an error in this context
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func FileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, err
	} else if os.IsNotExist(err) { // not exists is not an error in this context
		return false, nil
	}
	return false, err
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

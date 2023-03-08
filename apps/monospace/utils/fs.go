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

func FileExistsNoErr(filePath string) bool {
	res, err := FileExists(filePath)
	if err != nil {
		return false
	}
	return res
}

func WriteFile(filePath string, body string) error {
	bbody := []byte(body)
	// #nosec G306 - this is the purpose of this function to write a file
	return os.WriteFile(filePath, bbody, 0640)
}

func FileAppend(filePath string, appendString string) error {
	// #nosec G304 - we need to pass a variable of fileName to append to
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(appendString, "\n") {
		appendString = appendString + "\n"
	}
	_, err = f.WriteString(appendString)
	if err != nil {
		return err
	}
	return f.Close()

}

func FileRemoveLine(filePath string, line string) error {
	// #nosec G304 - this is the purpose of this function to read a file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	cleanContent := bytes.Replace(content, []byte(line+"\n"), []byte(""), -1)
	// #nosec G306 - we want to allow group access
	return os.WriteFile(filePath, cleanContent, 0640)
}

func MakeDir(path string) error {
	return os.MkdirAll(path, 0750)
}

func RmDir(path string) error {
	return os.RemoveAll(path)
}

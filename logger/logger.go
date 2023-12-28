package logger

import (
	"io"
	"log"
	"os"
)

func SetLoggerFile(fileName string) (*os.File, error) {
	f, err := os.OpenFile("logger/test.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	log.SetOutput(io.MultiWriter(f, os.Stdout))
	return f, nil
}

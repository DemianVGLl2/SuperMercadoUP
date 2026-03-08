package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

var fileLogger *log.Logger

func Init(filename string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	fileLogger = log.New(file, "", 0)
	return nil
}

func Log(role, action, detail string) {
	entry := fmt.Sprintf("[%s] [%s] %s: %s",
		time.Now().Format("2006-01-02 15:04:05"), role, action, detail)
	fmt.Println(entry)
	if fileLogger != nil {
		fileLogger.Println(entry)
	}
}

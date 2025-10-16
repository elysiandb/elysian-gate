package logger

import (
	"fmt"
	"os"
	"time"
)

const fileName = "elysianGate.log"

func Info(msg string) {
	go log("INFO", msg)
}

func Error(msg string) {
	go log("ERROR", msg)
}

func log(level, msg string) {
	entry := fmt.Sprintf("[%s] %s %s\n", level, time.Now().Format("2006-01-02 15:04:05"), msg)
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(entry)
}

// server/util/logger.go
package utils

import (
	"log"
	"os"
	"time"
)

type Logger struct {
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
}

var defaultLogger *Logger

func InitLogger() {
	defaultLogger = &Logger{
		info:  log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile),
		warn:  log.New(os.Stdout, "[WARN] ", log.Ldate|log.Ltime|log.Lshortfile),
		error: log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func Info(format string, v ...interface{}) {
	if defaultLogger == nil {
		InitLogger()
	}
	defaultLogger.info.Printf(format, v...)
}

func Warn(format string, v ...interface{}) {
	if defaultLogger == nil {
		InitLogger()
	}
	defaultLogger.warn.Printf(format, v...)
}

func Error(format string, v ...interface{}) {
	if defaultLogger == nil {
		InitLogger()
	}
	defaultLogger.error.Printf(format, v...)
}

func LogRequest(method, path string, duration time.Duration) {
	Info("Request: %s %s - %v", method, path, duration)
}

func LogError(err error, context string) {
	Error("Error in %s: %v", context, err)
}

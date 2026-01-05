package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

type Logger struct {
	file      *os.File
	debugMode bool
}

var defaultLogger *Logger

func Init(logPath string, debugMode bool) error {
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("impossibile aprire file di log: %w", err)
	}

	defaultLogger = &Logger{
		file:      f,
		debugMode: debugMode,
	}

	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	return nil
}

func Close() error {
	if defaultLogger != nil && defaultLogger.file != nil {
		return defaultLogger.file.Close()
	}
	return nil
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	if !l.debugMode && level == DEBUG {
		return
	}

	prefix := map[Level]string{
		DEBUG: "[DEBUG]",
		INFO:  "[INFO] ",
		WARN:  "[WARN] ",
		ERROR: "[ERROR]",
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	log.Printf("%s %s %s", timestamp, prefix[level], message)
}

func Debug(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(DEBUG, format, args...)
	}
}

func Info(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(INFO, format, args...)
	}
}

func Warn(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(WARN, format, args...)
	}
}

func Error(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(ERROR, format, args...)
	}
}

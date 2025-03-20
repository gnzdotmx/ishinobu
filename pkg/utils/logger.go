package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Logger struct {
	file      *os.File
	log       *log.Logger
	verbosity int
}

func NewLogger() *Logger {
	timestamp := time.Now().Format("20060102T150405")
	filename := fmt.Sprintf("ishinobu_%s.log", timestamp)
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Failed to create log file: %v\n", err)
		os.Exit(1)
	}

	return NewLoggerWithFile(file)
}

func NewLoggerWithFile(file *os.File) *Logger {
	return &Logger{
		file:      file,
		log:       log.New(file, "", log.LstdFlags),
		verbosity: 1,
	}
}

func (l *Logger) SetVerbosity(level int) {
	l.verbosity = level
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.verbosity >= 1 {
		l.log.Printf("INFO: "+format, v...)
	}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.verbosity >= 2 {
		l.log.Printf("DEBUG: "+format, v...)
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.log.Printf("ERROR: "+format, v...)
}

func (l *Logger) Close() error {
	return l.file.Close()
}

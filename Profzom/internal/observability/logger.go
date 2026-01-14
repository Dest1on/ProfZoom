package observability

import (
	"log"
	"os"
)

type Logger struct {
	info  *log.Logger
	error *log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		info:  log.New(os.Stdout, "INFO: ", log.LstdFlags|log.LUTC),
		error: log.New(os.Stderr, "ERROR: ", log.LstdFlags|log.LUTC),
	}
}

func (l *Logger) Info(msg string) {
	l.info.Println(msg)
}

func (l *Logger) Error(msg string) {
	l.error.Println(msg)
}

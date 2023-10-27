package log

import (
	"log"
	"os"
)

type Logger struct {
	logger *log.Logger
}

func NewLogger(output *os.File) *Logger {
	return &Logger{
		logger: log.New(output, "", log.Ldate|log.Ltime),
	}
}

func (l *Logger) Debug(v ...interface{}) {
	l.logger.Println(append([]interface{}{"[DEBUG]"}, v...)...)
}

func (l *Logger) Info(v ...interface{}) {
	l.logger.Println(append([]interface{}{"[INFO]"}, v...)...)
}

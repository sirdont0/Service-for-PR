package infra

import "log"

type Logger interface {
    Infof(format string, v ...interface{})
    Errorf(format string, v ...interface{})
}

type stdLogger struct{}

func NewStdLogger() Logger { return &stdLogger{} }
func (l *stdLogger) Infof(format string, v ...interface{})  { log.Printf("[INFO] "+format, v...) }
func (l *stdLogger) Errorf(format string, v ...interface{}) { log.Printf("[ERROR] "+format, v...) }

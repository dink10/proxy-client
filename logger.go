package proxy_client

import "fmt"

type DefaultLogger struct {
}

func NewLogger() *DefaultLogger {
	return &DefaultLogger{}
}

func (l *DefaultLogger) Error(args ...interface{}) {
	fmt.Println(args...)
}

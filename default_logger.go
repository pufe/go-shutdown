package shutdown

import "log"

type Logger interface {
	Info(msg string)
	Error(err error, msg string)
}

type defaultLogger struct{}

func (defaultLogger) Info(msg string) {
	log.Println(msg)
}
func (defaultLogger) Error(err error, msg string) {
	log.Printf("%s (%s)\n", msg, err)
}


var logger Logger = defaultLogger{}

func SetLogger(log Logger) {
	logger = log
}

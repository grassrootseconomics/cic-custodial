package server

import "github.com/zerodha/logf"

type asynqLogger struct {
	lo *logf.Logger
}

func asynqCompatibleLogger(lo logf.Logger) asynqLogger {
	return asynqLogger{
		lo: &lo,
	}
}

func (l asynqLogger) Debug(args ...interface{}) {
	l.lo.Debug("asynq", "debug", args[0])
}

func (l asynqLogger) Info(args ...interface{}) {
	l.lo.Info("asynq", "info", args[0])
}

func (l asynqLogger) Warn(args ...interface{}) {
	l.lo.Warn("asynq", "warn", args[0])
}

func (l asynqLogger) Error(args ...interface{}) {
	l.lo.Error("asynq", "error", args[0])
}

func (l asynqLogger) Fatal(args ...interface{}) {
	l.lo.Fatal("asynq", "fatal", args[0])
}

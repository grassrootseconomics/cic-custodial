package logg

import "github.com/zerodha/logf"

type AsynqLogger struct {
	Lo *logf.Logger
}

func AsynqCompatibleLogger(lo logf.Logger) AsynqLogger {
	return AsynqLogger{
		Lo: &lo,
	}
}

func (l AsynqLogger) Debug(args ...interface{}) {
	l.Lo.Debug("asynq", "debug", args[0])
}

func (l AsynqLogger) Info(args ...interface{}) {
	l.Lo.Info("asynq", "info", args[0])
}

func (l AsynqLogger) Warn(args ...interface{}) {
	l.Lo.Warn("asynq", "warn", args[0])
}

func (l AsynqLogger) Error(args ...interface{}) {
	l.Lo.Error("asynq", "error", args[0])
}

func (l AsynqLogger) Fatal(args ...interface{}) {
	l.Lo.Fatal("asynq", "fatal", args[0])
}

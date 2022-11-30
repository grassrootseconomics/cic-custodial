package logg

import "github.com/zerodha/logf"

type AsynqLogg struct {
	logg *logf.Logger
}

func NewAsynqLogg(lo logf.Logger) AsynqLogg {
	return AsynqLogg{
		logg: &lo,
	}
}

func (l AsynqLogg) Debug(args ...interface{}) {
	l.logg.Debug("asynq", "debug", args[0])
}

func (l AsynqLogg) Info(args ...interface{}) {
	l.logg.Info("asynq", "info", args[0])
}

func (l AsynqLogg) Warn(args ...interface{}) {
	l.logg.Warn("asynq", "warn", args[0])
}

func (l AsynqLogg) Error(args ...interface{}) {
	l.logg.Error("asynq", "error", args[0])
}

func (l AsynqLogg) Fatal(args ...interface{}) {
	l.logg.Fatal("asynq", "fatal", args[0])
}

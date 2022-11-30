package logg

import "github.com/zerodha/logf"

type LoggOpts struct {
	Debug  bool
	Caller bool
	Color  bool
}

func NewLogg(o LoggOpts) logf.Logger {
	loggConfig := logf.Opts{
		EnableColor:  o.Color,
		EnableCaller: o.Caller,
	}

	if o.Debug {
		loggConfig.Level = logf.DebugLevel
	} else {
		loggConfig.Level = logf.InfoLevel
	}

	return logf.New(loggConfig)
}

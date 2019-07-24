package logger

import (
	"github.com/borderstech/logmatic"
)

var log *logmatic.Logger

// Get initialises the Log object and return it
func Get() *logmatic.Logger {
	if log == nil {
		log = logmatic.NewLogger()
		log.SetLevel(logmatic.TRACE)
	}

	return log
}

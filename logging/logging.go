package logging

import (
	"github.com/op/go-logging"
)

const (
	MODULE_NAME = "adb-client"
)

func init() {
}

func GetLogger(name string) *logging.Logger {
	var log, _ = logging.GetLogger(MODULE_NAME)
	var format = logging.MustStringFormatter(`%{color}%{shortfunc} ▶ %{level:.5s} %{message}%{color:reset}`)
	logging.SetFormatter(format)
	logging.SetLevel(logging.DEBUG, MODULE_NAME)
	return log
}

func SetLevel(level logging.Level) {
	logging.SetLevel(level, MODULE_NAME)
}

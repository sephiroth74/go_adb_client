package logging

import (
	"github.com/op/go-logging"
)

func init() {
}

func GetLogger(name string) *logging.Logger {
	var log, _ = logging.GetLogger("adb")
	var format = logging.MustStringFormatter(`%{color}%{shortfunc} ▶ %{level:.5s} %{message}%{color:reset}`)
	logging.SetFormatter(format)
	return log
}

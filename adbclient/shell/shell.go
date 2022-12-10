package shell

import (
	"it.sephiroth/adbclient/transport"
	"it.sephiroth/adbclient/types"
)

type Shell[T types.Serial] struct {
	Serial T
	Adb    *string
}

func NewShell[T types.Serial](adb *string, serial T) *Shell[T] {
	var s = new(Shell[T])
	s.Serial = serial
	s.Adb = adb
	return s
}

func (s Shell[T]) Execute(command string, args ...string) (transport.Result, error) {
	// var serial = strings.Clone(s.serial.Serial())
	return transport.NewProcessBuilder(s.Serial).Path(s.Adb).Command("shell").Args(command).Args(args...).Invoke()
}

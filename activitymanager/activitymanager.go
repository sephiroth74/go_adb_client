package activitymanager

import (
	"it.sephiroth/adbclient/shell"
	"it.sephiroth/adbclient/transport"
	"it.sephiroth/adbclient/types"
)

type ActivityManager[T types.Serial] struct {
	Shell *shell.Shell[T]
}

func (a ActivityManager[T]) Broadcast(intent *types.Intent) {
	a.Shell.Execute("am", 0, "broadcast", intent.String())
}

func (a ActivityManager[T]) ForceStop(packageName string) (transport.Result, error) {
	return a.Shell.Executef("am force-stop %s", 0, packageName)
}

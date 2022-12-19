package activitymanager

import (
	"github.com/sephiroth74/go_adb_client/shell"
	"github.com/sephiroth74/go_adb_client/transport"
	"github.com/sephiroth74/go_adb_client/types"
)

type ActivityManager struct {
	Shell *shell.Shell
}

func (a ActivityManager) Broadcast(intent *types.Intent) {
	a.Shell.Execute("am", 0, "broadcast", intent.String())
}

func (a ActivityManager) ForceStop(packageName string) (transport.Result, error) {
	return a.Shell.Executef("am force-stop %s", 0, packageName)
}

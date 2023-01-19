package activitymanager

import (
	"github.com/sephiroth74/go_adb_client/process"
	"github.com/sephiroth74/go_adb_client/shell"
	"github.com/sephiroth74/go_adb_client/transport"
	"github.com/sephiroth74/go_adb_client/types"
)

type ActivityManager struct {
	Shell *shell.Shell
}

func (a ActivityManager) Broadcast(intent *types.Intent) (process.OutputResult, error) {
	cmd := a.Shell.NewCommand().Withargs("am", "broadcast", intent.String())
	return process.SimpleOutput(cmd, a.Shell.Conn.Verbose)
	// return a.Shell.ExecuteWithTimeout("am", 0, "broadcast", intent.String())
}

func (a ActivityManager) Start(intent *types.Intent) (transport.Result, error) {
	return a.Shell.ExecuteWithTimeout("am", 0, "start", intent.String())
}

func (a ActivityManager) ForceStop(packageName string) (transport.Result, error) {
	return a.Shell.Executef("am force-stop %s", packageName)
}

package activitymanager

import (
	"fmt"

	"github.com/sephiroth74/go_adb_client/process"
	"github.com/sephiroth74/go_adb_client/shell"
	"github.com/sephiroth74/go_adb_client/types"
)

type ActivityManager struct {
	Shell *shell.Shell
}

func (a ActivityManager) Broadcast(intent *types.Intent) (process.OutputResult, error) {
	cmd := a.Shell.NewCommand().WithArgs("am", "broadcast", intent.String())
	return process.SimpleOutput(cmd, a.Shell.Conn.Verbose)
}

func (a ActivityManager) Start(intent *types.Intent) (process.OutputResult, error) {
	return process.SimpleOutput(a.Shell.NewCommand().WithArgs("am", "start", intent.String()), a.Shell.Conn.Verbose)
}

func (a ActivityManager) StartService(intent *types.Intent) (process.OutputResult, error) {
	return process.SimpleOutput(a.Shell.NewCommand().WithArgs("am", "startservice", intent.String()), a.Shell.Conn.Verbose)
}

func (a ActivityManager) ForceStop(packageName string) error {
	result, err := process.SimpleOutput(a.Shell.NewCommand().WithArgs(fmt.Sprintf("am force-stop %s", packageName)), a.Shell.Conn.Verbose)
	if err != nil {
		return err
	}

	if !result.IsOk() {
		return result.NewError()
	}

	return nil
}

package adbclient

import (
	"os"

	"it.sephiroth/adbclient/activitymanager"
	"it.sephiroth/adbclient/packagemanager"
	"it.sephiroth/adbclient/transport"
	"it.sephiroth/adbclient/types"
)

type Device[T types.Serial] struct {
	Client *Client[T]
}

func NewDevice[T types.Serial](client *Client[T]) *Device[T] {
	device := new(Device[T])
	device.Client = client
	return device
}

func (d Device[T]) ActivityManager() *activitymanager.ActivityManager[T] {
	return &activitymanager.ActivityManager[T]{
		Shell: d.Client.Shell,
	}
}

func (d Device[T]) PackageManager() *packagemanager.PackageManager[T] {
	return &packagemanager.PackageManager[T]{
		Shell: d.Client.Shell,
	}
}

func (d Device[T]) Name() *string {
	return d.Client.Shell.GetProp("ro.build.product")
}

func (d Device[T]) ApiLevel() *string {
	return d.Client.Shell.GetProp("ro.build.version.sdk")
}

func (d Device[T]) Version() *string {
	return d.Client.Shell.GetProp("ro.build.version.release")
}

func (d Device[T]) SaveScreenCap(output string) (transport.Result, error) {
	return d.Client.Shell.Execute("screencap", 0, "-p", output)
}

func (d Device[T]) WriteScreenCap(output *os.File) (transport.Result, error) {
	var pb = d.Client.NewProcess()
	pb.Command("exec-out")
	pb.Args("screencap", "-p")
	pb.Verbose(true)
	pb.Stdout(output)
	return pb.Invoke()
}

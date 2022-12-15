package adbclient

import (
	"os"

	"github.com/sephiroth74/go_adb_client/activitymanager"
	"github.com/sephiroth74/go_adb_client/input"
	"github.com/sephiroth74/go_adb_client/packagemanager"
	"github.com/sephiroth74/go_adb_client/transport"
	"github.com/sephiroth74/go_adb_client/types"
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

// PowerOff Power off the device (turn the screen off).
// If the screen is already off it returns false, true otherwise
func (d Device[T]) PowerOff() (bool, error) {
	screenon, err := d.IsScreenOn()
	if err != nil {
		return false, err
	}

	if screenon {
		return d.Power()
	} else {
		return true, nil
	}
}

// PowerOn Power on the device (turn the screen off).
// If the screen is already on it returns false, true otherwise
func (d Device[T]) PowerOn() (bool, error) {
	screenon, err := d.IsScreenOn()
	if err != nil {
		return false, err
	}

	if !screenon {
		return d.Power()
	} else {
		return false, nil
	}
}

// Power Send a KEYCODE_POWER input event to the device
func (d Device[T]) Power() (bool, error) {
	result, err := d.Client.Shell.SendKeyEvent(input.KEYCODE_POWER)
	if err != nil {
		return false, err
	}
	return result.IsOk(), nil
}

// IsScreenOn Return true if the device screen is on
func (d Device[T]) IsScreenOn() (bool, error) {
	result, err := d.Client.Shell.Execute("dumpsys input_method | egrep 'screenOn *=' | sed 's/ *screenOn = \\(.*\\)/\\1/g'", 0)
	if err != nil {
		return false, err
	}

	if result.IsOk() {
		return result.Output() == "true", nil
	} else {
		return false, result.NewError()
	}
}

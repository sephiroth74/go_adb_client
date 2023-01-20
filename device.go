package adbclient

import (
	"bufio"
	"io"
	"os"

	"github.com/sephiroth74/go_adb_client/activitymanager"
	"github.com/sephiroth74/go_adb_client/input"
	"github.com/sephiroth74/go_adb_client/packagemanager"
	"github.com/sephiroth74/go_adb_client/process"
)

type Device struct {
	Client *Client
}

func NewDevice(client *Client) *Device {
	device := new(Device)
	device.Client = client
	return device
}

func (d Device) ActivityManager() *activitymanager.ActivityManager {
	return &activitymanager.ActivityManager{
		Shell: d.Client.Shell,
	}
}

func (d Device) PackageManager() *packagemanager.PackageManager {
	return &packagemanager.PackageManager{
		Shell: d.Client.Shell,
	}
}

func (d Device) Name() *string {
	return d.Client.Shell.GetProp("ro.build.product")
}

func (d Device) ApiLevel() *string {
	return d.Client.Shell.GetProp("ro.build.version.sdk")
}

func (d Device) Version() *string {
	return d.Client.Shell.GetProp("ro.build.version.release")
}

func (d Device) SaveScreenCap(output string) (process.OutputResult, error) {
	cmd := d.Client.Shell.NewCommand().WithArgs("screencap", "-p", output)
	return process.SimpleOutput(cmd, d.Client.Conn.Verbose)
	// return d.Client.Shell.ExecuteWithTimeout("screencap", 0, "-p", output)
}

func (d Device) WriteScreenCap(output *os.File) (process.OutputResult, error) {
	var writer io.Writer = bufio.NewWriter(output)
	cmd := d.Client.NewAdbCommand().WithCommand("exec-out").WithArgs("screencap", "-p").WithStdOut(writer)
	return process.SimpleOutput(cmd, d.Client.Conn.Verbose)
	// var pb = d.Client.NewProcess()
	// pb.WithCommand("exec-out")
	// pb.WithArgs("screencap", "-p")
	// pb.WithStdout(&writer)
	// return pb.Invoke()
}

// PowerOffOn send the power button input key
func (d Device) PowerOffOn() (bool, error) {
	return d.Power()
}

// PowerOff Power off the device (turn the screen off).
// If the screen is already off it returns false, true otherwise
func (d Device) PowerOff() (bool, error) {
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
func (d Device) PowerOn() (bool, error) {
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
func (d Device) Power() (bool, error) {
	result, err := d.Client.Shell.SendKeyEvent(input.KEYCODE_POWER)
	if err != nil {
		return false, err
	}
	return result.IsOk(), nil
}

// IsScreenOn Return true if the device screen is on
func (d Device) IsScreenOn() (bool, error) {
	cmd := d.Client.Shell.NewCommand().WithArgs("dumpsys input_method | egrep 'screenOn *=' | sed 's/ *screenOn = \\(.*\\)/\\1/g'")
	result, err := process.SimpleOutput(cmd, d.Client.Shell.Conn.Verbose)
	// result, err := d.Client.Shell.ExecuteWithTimeout("dumpsys input_method | egrep 'screenOn *=' | sed 's/ *screenOn = \\(.*\\)/\\1/g'", 0)
	if err != nil {
		return false, err
	}

	if result.IsOk() {
		return result.Output() == "true", nil
	} else {
		return false, result.NewError()
	}
}

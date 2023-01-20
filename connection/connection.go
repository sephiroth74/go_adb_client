package connection

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sephiroth74/go_adb_client/process"
	"github.com/sephiroth74/go_adb_client/types"
	"github.com/sephiroth74/go_adb_client/util/constants"

	"pkg.re/essentialkaos/ek.v12/env"
)

var RebootTimeout = time.Duration(30) * time.Second

type Connection struct {
	ADBPath string
	Verbose bool
}

func NewConnection(verbose bool) *Connection {
	path := env.Which("adb")
	conn := new(Connection)
	conn.Verbose = verbose
	conn.ADBPath = path
	return conn
}

func (c Connection) NewAdbCommand() *process.ADBCommand {
	return process.NewADBCommand(c.ADBPath)
}

// Version returns the adb version
func (c Connection) Version() (string, error) {
	cmd := c.NewAdbCommand().WithArgs("--version")
	result, err := process.SimpleOutput(cmd, c.Verbose)
	if err != nil {
		return "", err
	}

	lines := strings.Split(result.Output(), "\n")

	if len(lines) > 0 {
		r := regexp.MustCompile(`.*\s([\w]+\.[\w]+\.[\w]+)`)
		m := r.FindStringSubmatch(lines[0])
		if len(m) == 2 {
			return m[1], nil
		}
	}
	return "", nil
}

// Connect try to connect to the device with the given address
func (c Connection) Connect(addr string, timeout time.Duration) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithCommand("connect").WithArgs(addr).WithTimeout(timeout)
	return process.SimpleOutput(cmd, c.Verbose)
}

// Reconnect try to reconnect to the given device address
func (c Connection) Reconnect(t types.ReconnectType, timeout time.Duration) (process.OutputResult, error) {
	return process.SimpleOutput(c.NewAdbCommand().
		WithCommand("reconnect").
		WithArgs(string(t)).
		WithTimeout(timeout), c.Verbose)
}

// Disconnect disconnect the given device
func (c Connection) Disconnect(addr string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithCommand("disconnect").WithArgs(addr)
	return process.SimpleOutput(cmd, c.Verbose)
}

// DisconnectAll disconnect from any connected devices
func (c Connection) DisconnectAll() (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithCommand("disconnect")
	return process.SimpleOutput(cmd, c.Verbose)
}

// GetState gets the connection state with the given device address
func (c Connection) GetState(addr string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithCommand("get-state").WithSerial(addr)
	return process.SimpleOutput(cmd, c.Verbose)
}

// WaitForDevice returns when the device is ready
func (c Connection) WaitForDevice(addr string, timeout time.Duration) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithSerial(addr).
		WithTimeout(timeout).
		WithCommand("wait-for-device").
		WithArgs("shell", "while [[ -z $(getprop sys.boot_completed) ]]; do sleep 1; done; input keyevent 143")
	return process.SimpleOutput(cmd, c.Verbose)
}

func (c Connection) Root(addr string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithSerial(addr).WithCommand("root")
	return process.SimpleOutput(cmd, c.Verbose)
}

func (c Connection) UnRoot(addr string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithSerial(addr).WithCommand("unroot")
	return process.SimpleOutput(cmd, c.Verbose)
}

func (c Connection) IsRoot(addr string) (bool, error) {
	result, err := process.SimpleOutput(
		c.NewAdbCommand().WithCommand("shell").WithArgs("whoami").WithSerial(addr),
		c.Verbose,
	)

	if err != nil {
		return false, err
	}

	return strings.Contains(result.Output(), "root"), nil
}

func (c Connection) Reboot(addr string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithSerial(addr).WithCommand("reboot")
	return process.SimpleOutput(cmd, c.Verbose)
}

func (c Connection) Remount(addr string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithSerial(addr).WithCommand("remount")
	return process.SimpleOutput(cmd, c.Verbose)
}

func (c Connection) Mount(addr string, dir string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithSerial(addr).WithCommand("shell").WithArgs(fmt.Sprintf("mount -o rw,remount %s", dir))
	return process.SimpleOutput(cmd, c.Verbose)
}

func (c Connection) Unmount(addr string, dir string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithSerial(addr).WithCommand("shell").WithArgs(fmt.Sprintf("mount -o ro,remount %s", dir))
	return process.SimpleOutput(cmd, c.Verbose)
}

func (c Connection) ListDevices() ([]*types.Device, error) {
	cmd := c.NewAdbCommand().
		WithCommand("devices").
		WithArgs("-l")

	result, err := process.SimpleOutput(cmd, c.Verbose)

	if err != nil {
		return nil, err
	}

	var devices []*types.Device

	r := regexp.MustCompile(`(?P<ip>[^\s]+)[\s]+device product:(?P<device_product>[^\s]+)\smodel:(?P<model>[^\s]+)\sdevice:(?P<device>[^\s]+)\stransport_id:(?P<transport_id>[^\s]+)`)
	lines := strings.Split(result.Output(), "\n")
	if len(lines) > 1 {
		for x := 1; x < len(lines); x++ {
			m := r.FindStringSubmatch(lines[x])

			if len(m) == 6 {
				device, err := types.NewDevice(&m[1])

				if err != nil {
					continue
				}

				device.Product = []byte(m[2])
				device.Model = []byte(m[3])
				device.Device = []byte(m[4])
				device.Transport = []byte(m[5])

				devices = append(devices, device)
			}
		}
	}
	return devices, nil
}

func (c Connection) BugReport(addr string, dst string) (process.OutputResult, error) {
	return process.SimpleOutput(c.NewAdbCommand().WithSerial(addr).WithCommand("bugreport").WithArgs(dst), c.Verbose)
}

func (c Connection) Pull(addr string, src string, dst string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithSerial(addr).WithCommand("pull").WithArgs(src, dst)
	return process.SimpleOutput(cmd, c.Verbose)
}

func (c Connection) Push(addr string, src string, dst string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithSerial(addr).WithCommand("push").WithArgs(src, dst)
	return process.SimpleOutput(cmd, c.Verbose)
}

func (c Connection) Install(addr string, src string, args ...string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithCommand("install").WithSerial(addr).WithArgs(args...).AddArgs(src)
	return process.SimpleOutput(cmd, c.Verbose)
}

func (c Connection) Uninstall(addr string, packageName string, args ...string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithCommand("uninstall").WithSerial(addr).WithArgs(args...).AddArgs(packageName)
	return process.SimpleOutput(cmd, c.Verbose)
}

func (c Connection) Which(addr string, command string) (string, error) {
	result, err := process.SimpleOutput(
		c.NewAdbCommand().WithSerial(addr).WithArgs("shell", constants.WHICH, command),
		c.Verbose,
	)
	if err != nil {
		return "", err
	}

	return result.Output(), nil
}

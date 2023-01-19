package connection

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sephiroth74/go_adb_client/process"
	"github.com/sephiroth74/go_adb_client/transport"
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

func (c Connection) NewProcessBuilder(verbose ...bool) *transport.ProcessBuilder {
	v := c.Verbose
	if len(verbose) > 0 {
		v = verbose[0]
	}
	return transport.NewProcessBuilder().
		Verbose(v).
		WithPath(&c.ADBPath)
}

func (c Connection) Version() (string, error) {
	cmd := c.NewAdbCommand().Withargs("--version")
	result, err := process.SimpleOutput(cmd, c.Verbose)
	// result, err := c.NewProcessBuilder().WithCommand("--version").Invoke()
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

func (c Connection) Connect(addr string, timeout time.Duration) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithCommand("connect").Withargs(addr).WithTimeout(timeout)
	return process.SimpleOutput(cmd, c.Verbose)
	// p := c.NewProcessBuilder().
	// WithCommand("connect").
	// WithArgs(addr).
	// WithTimeout(timeout)
	// return p.Invoke()
}

func (c Connection) Reconnect(addr string, timeout time.Duration) (transport.Result, error) {
	return c.NewProcessBuilder().
		WithCommand("reconnect").
		WithArgs(addr).
		WithTimeout(timeout).
		Invoke()
}

func (c Connection) Disconnect(addr string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithCommand("disconnect").Withargs(addr)
	return process.SimpleOutput(cmd, c.Verbose)
	// return c.NewProcessBuilder().
	// WithCommand("disconnect").
	// WithArgs(addr).
	// Invoke()
}

func (c Connection) DisconnectAll() (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithCommand("disconnect")
	return process.SimpleOutput(cmd, c.Verbose)
	// return c.NewProcessBuilder().
	// WithCommand("disconnect").
	// Invoke()
}

func (c Connection) GetState(addr string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithCommand("get-state").WithSerial(addr)
	return process.SimpleOutput(cmd, c.Verbose)
}

func (c Connection) WaitForDevice(addr string, timeout time.Duration) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithSerial(addr).
		WithTimeout(timeout).
		WithCommand("wait-for-device").
		Withargs("shell", "while [[ -z $(getprop sys.boot_completed) ]]; do sleep 1; done; input keyevent 143")
	return process.SimpleOutput(cmd, c.Verbose)
	// return c.NewProcessBuilder().
	// 	WithSerialAddr(addr).
	// 	WithTimeout(timeout).
	// 	WithCommand("wait-for-device").
	// 	WithArgs("shell", "while [[ -z $(getprop sys.boot_completed) ]]; do sleep 1; done; input keyevent 143").
	// 	Invoke()
}

func (c Connection) Root(addr string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithSerial(addr).WithCommand("root")
	return process.SimpleOutput(cmd, c.Verbose)
	// return c.NewProcessBuilder().
	// WithSerialAddr(addr).
	// WithCommand("root").
	// Invoke()
}

func (c Connection) UnRoot(addr string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithSerial(addr).WithCommand("unroot")
	return process.SimpleOutput(cmd, c.Verbose)
	// return c.NewProcessBuilder().
	// WithSerialAddr(addr).
	// WithCommand("unroot").
	// Invoke()
}

func (c Connection) IsRoot(addr string) (bool, error) {
	result, err := c.NewProcessBuilder().
		WithSerialAddr(addr).
		WithCommand("shell").
		WithArgs("whoami").Invoke()
	if err != nil {
		return false, err
	}
	return result.Output() == "root", nil
}

func (c Connection) Reboot(addr string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithSerial(addr).WithCommand("reboot")
	return process.SimpleOutput(cmd, c.Verbose)
	// return c.NewProcessBuilder().
		// WithSerialAddr(addr).
		// WithCommand("reboot").
		// Invoke()
}

func (c Connection) Remount(addr string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithSerial(addr).WithCommand("remount")
	return process.SimpleOutput(cmd, c.Verbose)
	// return c.NewProcessBuilder().
		// WithSerialAddr(addr).
		// WithCommand("remount").
		// Invoke()
}

func (c Connection) Mount(addr string, dir string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithSerial(addr).WithCommand("shell").Withargs(fmt.Sprintf("mount -o rw,remount %s", dir))
	return process.SimpleOutput(cmd, c.Verbose)
	// return c.NewProcessBuilder().
		// WithSerialAddr(addr).
		// WithCommand("shell").
		// WithArgs(fmt.Sprintf("mount -o rw,remount %s", dir)).Invoke()
}

func (c Connection) Unmount(addr string, dir string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithSerial(addr).WithCommand("shell").Withargs(fmt.Sprintf("mount -o ro,remount %s", dir))
	return process.SimpleOutput(cmd, c.Verbose)
	// return c.NewProcessBuilder().
		// WithSerialAddr(addr).
		// WithCommand("shell").
		// WithArgs(fmt.Sprintf("mount -o ro,remount %s", dir)).
		// Invoke()
}

func (c Connection) ListDevices() ([]*types.Device, error) {
	cmd := c.NewAdbCommand().
		WithCommand("devices").
		Withargs("-l")

	result, err := process.SimpleOutput(cmd, c.Verbose)

	// result, err := c.NewProcessBuilder().
		// WithCommand("devices").
		// WithArgs("-l").
		// Invoke()

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

func (c Connection) BugReport(addr string, dst string) *transport.ProcessBuilder {
	return c.NewProcessBuilder().
		WithSerialAddr(addr).
		WithCommand("bugreport").
		WithArgs(dst)
}

func (c Connection) Pull(addr string, src string, dst string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithSerial(addr).WithCommand("pull").Withargs(src, dst)
	return process.SimpleOutput(cmd, c.Verbose)
	// return c.NewProcessBuilder().
		// WithSerialAddr(addr).
		// WithCommand("pull").
		// WithArgs(src, dst).
		// Invoke()
}

func (c Connection) Push(addr string, src string, dst string) (process.OutputResult, error) {
	cmd := c.NewAdbCommand().WithSerial(addr).WithCommand("push").Withargs(src, dst)
	return process.SimpleOutput(cmd, c.Verbose)
	// return c.NewProcessBuilder().
		// WithSerialAddr(addr).
		// WithCommand("push").
		// WithArgs(src, dst).
		// Invoke()
}

func (c Connection) Install(addr string, src string, args ...string) (transport.Result, error) {
	cmd := []string{"install"}
	cmd = append(cmd, args...)
	cmd = append(cmd, src)
	return c.NewProcessBuilder().
		WithSerialAddr(addr).
		WithArgs(args...).
		WithArgs(src).
		WithCommand("install").
		Invoke()
}

func (c Connection) Uninstall(addr string, packageName string, args ...string) (transport.Result, error) {
	cmd := []string{"uninstall"}
	cmd = append(cmd, args...)
	cmd = append(cmd, packageName)

	return c.NewProcessBuilder().
		WithSerialAddr(addr).
		WithCommand("uninstall").
		WithArgs(args...).
		WithArgs(packageName).
		Invoke()
}

func (c Connection) Which(addr string, command string) (string, error) {
	result, err := c.NewProcessBuilder().
		WithSerialAddr(addr).
		WithCommand(constants.WHICH).
		WithArgs(command).
		Invoke()

	if err != nil {
		return "", err
	}

	return result.Output(), nil
}

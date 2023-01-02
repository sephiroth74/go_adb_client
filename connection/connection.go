package connection

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sephiroth74/go_adb_client/transport"
	"github.com/sephiroth74/go_adb_client/types"
	"github.com/sephiroth74/go_adb_client/util/constants"

	"pkg.re/essentialkaos/ek.v12/env"
)

var timeout = time.Duration(5) * time.Second
var RebootTimeout = time.Duration(30) * time.Second
var WaitForDeviceTimeout = time.Duration(1) * time.Minute

type Connection struct {
	ADBPath string
	Verbose bool
}

func NewConnection(verbose bool) *Connection {
	path := env.Which("adb")
	conn := new(Connection)
	conn.ADBPath = path
	conn.Verbose = verbose
	return conn
}

func (c Connection) NewProcessBuilder() *transport.ProcessBuilder {
	return transport.NewProcessBuilder().Verbose(c.Verbose).WithPath(&c.ADBPath)
}

func (c Connection) Version() (string, error) {
	result, err := c.NewProcessBuilder().WithPath(&c.ADBPath).WithCommand("--version").Invoke()
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

func (c Connection) Connect(addr string, timeout time.Duration) (transport.Result, error) {
	p := c.NewProcessBuilder().
		WithPath(&c.ADBPath).
		WithCommand("connect").
		WithArgs(addr).
		WithTimeout(timeout)
	return p.Invoke()
}

func (c Connection) Reconnect(addr string, timeout time.Duration) (transport.Result, error) {
	return c.NewProcessBuilder().
		WithPath(&c.ADBPath).
		WithCommand("reconnect").
		WithArgs(addr).
		WithTimeout(timeout).
		Invoke()
}

func (c Connection) Disconnect(addr string) (transport.Result, error) {
	return c.NewProcessBuilder().WithPath(&c.ADBPath).WithCommand("disconnect").WithArgs(addr).Invoke()
}

func (c Connection) DisconnectAll() (transport.Result, error) {
	return c.NewProcessBuilder().WithPath(&c.ADBPath).WithCommand("disconnect").Invoke()
}

func (c Connection) GetState(addr string) (transport.Result, error) {
	return c.NewProcessBuilder().
		WithPath(&c.ADBPath).
		WithSerialAddr(addr).
		WithCommand("get-state").
		Invoke()
}

func (c Connection) WaitForDevice(addr string, timeout time.Duration) (transport.Result, error) {
	return c.WaitForDeviceWithTimeout(addr, timeout)
}

func (c Connection) WaitForDeviceWithTimeout(addr string, timeout time.Duration) (transport.Result, error) {
	return c.NewProcessBuilder().
		WithPath(&c.ADBPath).
		WithSerialAddr(addr).
		WithTimeout(timeout).
		WithCommand("wait-for-device").
		WithArgs("shell", "while [[ -z $(getprop sys.boot_completed) ]]; do sleep 1; done; input keyevent 143").
		Invoke()
}

func (c Connection) Root(addr string) (transport.Result, error) {
	return c.NewProcessBuilder().WithPath(&c.ADBPath).WithSerialAddr(addr).WithCommand("root").Invoke()
}

func (c Connection) UnRoot(addr string) (transport.Result, error) {
	return c.NewProcessBuilder().WithPath(&c.ADBPath).WithSerialAddr(addr).WithCommand("unroot").Invoke()
}

func (c Connection) IsRoot(addr string) (bool, error) {
	result, err := c.NewProcessBuilder().
		WithPath(&c.ADBPath).
		WithSerialAddr(addr).
		WithCommand("shell").
		WithArgs("whoami").Invoke()
	if err != nil {
		return false, err
	}
	return result.Output() == "root", nil
}

func (c Connection) Reboot(addr string) (transport.Result, error) {
	return c.NewProcessBuilder().WithPath(&c.ADBPath).WithSerialAddr(addr).WithCommand("reboot").Invoke()
}

func (c Connection) Remount(addr string) (transport.Result, error) {
	return c.NewProcessBuilder().WithPath(&c.ADBPath).WithSerialAddr(addr).WithCommand("remount").Invoke()
}

func (c Connection) Mount(addr string, dir string) (transport.Result, error) {
	return c.NewProcessBuilder().
		WithPath(&c.ADBPath).
		WithSerialAddr(addr).
		WithCommand("shell").
		WithArgs(fmt.Sprintf("mount -o rw,remount %s", dir)).Invoke()
}

func (c Connection) Unmount(addr string, dir string) (transport.Result, error) {
	return c.NewProcessBuilder().
		WithPath(&c.ADBPath).
		WithSerialAddr(addr).
		WithCommand("shell").
		WithArgs(fmt.Sprintf("mount -o ro,remount %s", dir)).
		Invoke()
}

func (c Connection) ListDevices() ([]*types.Device, error) {
	result, err := c.NewProcessBuilder().WithPath(&c.ADBPath).WithCommand("devices").WithArgs("-l").Invoke()

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

func (c Connection) BugReport(addr string, dst string) (transport.Result, error) {
	return c.NewProcessBuilder().WithPath(&c.ADBPath).WithSerialAddr(addr).WithCommand("bugreport").WithArgs(dst).Invoke()
}

func (c Connection) Pull(addr string, src string, dst string) (transport.Result, error) {
	return c.NewProcessBuilder().WithPath(&c.ADBPath).WithSerialAddr(addr).WithCommand("pull").WithArgs(src, dst).Invoke()
}

func (c Connection) Push(addr string, src string, dst string) (transport.Result, error) {
	return c.NewProcessBuilder().WithPath(&c.ADBPath).WithSerialAddr(addr).WithCommand("push").WithArgs(src, dst).Invoke()
}

func (c Connection) Install(addr string, src string, args ...string) (transport.Result, error) {
	cmd := []string{"install"}
	cmd = append(cmd, args...)
	cmd = append(cmd, src)
	return c.NewProcessBuilder().
		WithPath(&c.ADBPath).
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
		WithPath(&c.ADBPath).
		WithSerialAddr(addr).
		WithCommand("uninstall").
		WithArgs(args...).
		WithArgs(packageName).
		Invoke()
}

func (c Connection) Which(addr string, command string) (string, error) {
	result, err := c.NewProcessBuilder().
		WithPath(&c.ADBPath).
		WithSerialAddr(addr).
		WithCommand(constants.WHICH).
		WithArgs(command).
		Invoke()

	if err != nil {
		return "", err
	}

	return result.Output(), nil
}

package connection

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"it.sephiroth/adbclient/transport"
	"it.sephiroth/adbclient/types"
	"it.sephiroth/adbclient/util/constants"

	"pkg.re/essentialkaos/ek.v12/env"
)

var timeout = time.Duration(5) * time.Second
var reboot_timeout = time.Duration(30) * time.Second
var wait_for_device_timeout = time.Duration(1) * time.Minute

type Connection struct {
	ADBPath string
}

func NewConnection() *Connection {
	path := env.Which("adb")
	conn := new(Connection)
	conn.ADBPath = path
	return conn
}

func (c Connection) Version() (string, error) {
	result, err := transport.Invoke(&c.ADBPath, timeout, "--version")
	if err != nil {
		return "", err
	}

	lines := strings.Split(result.GetOutput(), "\n")

	if len(lines) > 0 {
		r := regexp.MustCompile(`.*\s([\w]+\.[\w]+\.[\w]+)`)
		m := r.FindStringSubmatch(lines[0])
		if len(m) == 2 {
			return m[1], nil
		}
	}
	return "", nil
}

func (c Connection) Connect(addr string) (transport.Result, error) {
	return transport.Invoke(&c.ADBPath, timeout, "connect", addr)
}

func (c Connection) Reconnect(addr string) (transport.Result, error) {
	return transport.Invoke(&c.ADBPath, timeout, "reconnect", addr)
}

func (c Connection) Disconnect(addr string) (transport.Result, error) {
	return transport.Invoke(&c.ADBPath, timeout, "disconnect", addr)
}

func (c Connection) DisconnectAll() (transport.Result, error) {
	return transport.Invoke(&c.ADBPath, timeout, "disconnect")
}

func (c Connection) GetState(addr string) (transport.Result, error) {
	return transport.Invoke(&c.ADBPath, timeout, "-s", addr, "get-state")
}

func (c Connection) WaitForDevice(addr string) (transport.Result, error) {
	return c.WaitForDeviceWithTimeout(addr, wait_for_device_timeout)
}

func (c Connection) WaitForDeviceWithTimeout(addr string, timeout time.Duration) (transport.Result, error) {
	return transport.Invoke(&c.ADBPath, timeout, "-s", addr, "wait-for-device", "shell", "while [[ -z $(getprop sys.boot_completed) ]]; do sleep 1; done; input keyevent 143")
}

func (c Connection) Root(addr string) (transport.Result, error) {
	return transport.Invoke(&c.ADBPath, timeout, "-s", addr, "root")
}

func (c Connection) UnRoot(addr string) (transport.Result, error) {
	return transport.Invoke(&c.ADBPath, timeout, "-s", addr, "unroot")
}

func (c Connection) IsRoot(addr string) (bool, error) {
	result, err := transport.Invoke(&c.ADBPath, timeout, "-s", addr, "shell", "whoami")
	if err != nil {
		return false, err
	}
	return result.GetOutput() == "root", nil
}

func (c Connection) Reboot(addr string) (transport.Result, error) {
	return transport.Invoke(&c.ADBPath, reboot_timeout, "-s", addr, "reboot")
}

func (c Connection) Remount(addr string) (transport.Result, error) {
	return transport.Invoke(&c.ADBPath, reboot_timeout, "-s", addr, "remount")
}

func (c Connection) Mount(addr string, dir string) (transport.Result, error) {
	return transport.Invoke(&c.ADBPath, reboot_timeout, "-s", addr, "shell", fmt.Sprintf("mount -o rw,remount %s", dir))
}

func (c Connection) Unmount(addr string, dir string) (transport.Result, error) {
	return transport.Invoke(&c.ADBPath, reboot_timeout, "-s", addr, "shell", fmt.Sprintf("mount -o ro,remount %s", dir))
}

func (c Connection) ListDevices() ([]*types.Device, error) {
	result, err := transport.Invoke(&c.ADBPath, timeout, "devices", "-l")

	if err != nil {
		return nil, err
	}

	var devices = []*types.Device{}

	r := regexp.MustCompile(`(?P<ip>[^\s]+)[\s]+device product:(?P<device_product>[^\s]+)\smodel:(?P<model>[^\s]+)\sdevice:(?P<device>[^\s]+)\stransport_id:(?P<transport_id>[^\s]+)`)
	lines := strings.Split(result.GetOutput(), "\n")
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
	return transport.Invoke(&c.ADBPath, 0, "-s", addr, "bugreport", dst)
}

func (c Connection) Pull(addr string, src string, dst string) (transport.Result, error) {
	return transport.Invoke(&c.ADBPath, 0, "-s", addr, "pull", src, dst)
}

func (c Connection) Push(addr string, src string, dst string) (transport.Result, error) {
	return transport.Invoke(&c.ADBPath, 0, "-s", addr, "push", src, dst)
}

func (c Connection) Which(addr string, command string) {
	result, err := transport.Invoke(&c.ADBPath, 0, "-s", addr, constants.WHICH, command)

	if err != nil {
		println(err.Error())
		return
	}

	println("result: %s", result.Repr())
}
package adbclient

import (
	"fmt"
	streams "github.com/sephiroth74/go_streams"
	"net"
	"time"

	"github.com/reactivex/rxgo/v2"
	"github.com/sephiroth74/go_adb_client/connection"
	"github.com/sephiroth74/go_adb_client/events"
	"github.com/sephiroth74/go_adb_client/mdns"
	"github.com/sephiroth74/go_adb_client/shell"
	"github.com/sephiroth74/go_adb_client/transport"
	"github.com/sephiroth74/go_adb_client/types"
)

type Client struct {
	Conn    *connection.Connection
	Mdns    *mdns.Mdns
	Channel chan rxgo.Item
	Address types.Serial
	Shell   *shell.Shell
}

func NewClient(device types.Serial, verbose bool) *Client {
	var conn = connection.NewConnection(verbose)
	client := new(Client)
	client.Conn = conn
	client.Mdns = mdns.NewMdns(client.Conn)
	client.Address = device
	client.Channel = make(chan rxgo.Item)
	client.Shell = shell.NewShell(client.Conn, device)
	return client
}

func NullClient(verbose bool) *Client {
	return NewClient(types.ClientAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5555}, verbose)
}

func (c Client) NewProcess() *transport.ProcessBuilder {
	return c.Conn.NewProcessBuilder().WithSerial(&c.Address)
}

func (c Client) DeferredDispatch(eventType events.EventType) {
	defer func() { go func() { c.Channel <- rxgo.Of(events.AdbEvent{Event: eventType}) }() }()
}

func (c Client) Dispatch(eventType events.EventType, data interface{}) {
	go func() { c.Channel <- rxgo.Of(events.AdbEvent{Event: eventType, Item: data}) }()
}

func WaitAndReturn(result *transport.Result, err error, timeout time.Duration) (transport.Result, error) {
	if err != nil {
		return *result, err
	}
	time.Sleep(timeout)
	return *result, err
}

func (c Client) Connect(timeout time.Duration) (transport.Result, error) {
	conn, err := c.IsConnected()
	if err == nil && conn {
		return transport.OkResult("Already Connected"), nil
	}
	result, err := c.Conn.Connect(c.Address.GetSerialAddress(), timeout)

	if err != nil {
		return transport.ErrorResult(result.Output()), err
	}

	conn, err = c.IsConnected()
	if err != nil {
		return transport.ErrorResult("Unable to connect"), err
	}

	if conn {
		defer c.Dispatch(events.Connected, c.Address)
		return transport.OkResult(fmt.Sprintf("connected to %s", c.Address.String())), nil
	} else {
		return transport.ErrorResult(fmt.Sprintf("Unable to connect to %s", c.Address.String())), nil
	}
}

func (c Client) Reconnect(timeout time.Duration) (transport.Result, error) {
	return c.Conn.Reconnect(c.Address.GetSerialAddress(), timeout)
}

func (c Client) IsConnected() (bool, error) {
	result, err := c.Conn.GetState(c.Address.GetSerialAddress())
	if err != nil {
		return false, err
	}
	return result.IsOk(), nil
}

func (c Client) Disconnect() (transport.Result, error) {
	connected, err := c.IsConnected()
	if err == nil && !connected {
		return transport.OkResult(""), nil
	}

	result, err := c.Conn.Disconnect(c.Address.GetSerialAddress())

	if err == nil && result.IsOk() {
		defer c.Dispatch(events.Disconnect, c.Address)
	}

	return result, err
}

func (c Client) DisconnectAll() (transport.Result, error) {
	return c.Conn.DisconnectAll()
}

func (c Client) WaitForDevice() (transport.Result, error) {
	return c.Conn.WaitForDevice(c.Address.GetSerialAddress(), 0)
}

func (c Client) WaitForDeviceWithTimeout(timeout time.Duration) (transport.Result, error) {
	return c.Conn.WaitForDeviceWithTimeout(c.Address.GetSerialAddress(), timeout)
}

func (c Client) Root() (transport.Result, error) {
	result, err := c.Conn.Root(c.Address.GetSerialAddress())
	return WaitAndReturn(&result, err, time.Duration(1)*time.Second)
}

func (c Client) IsRoot() (bool, error) {
	return c.Conn.IsRoot(c.Address.GetSerialAddress())
}

func (c Client) UnRoot() (transport.Result, error) {
	result, err := c.Conn.UnRoot(c.Address.GetSerialAddress())
	return WaitAndReturn(&result, err, time.Duration(1)*time.Second)
}

func (c Client) ListDevices() ([]*types.Device, error) {
	return c.Conn.ListDevices()
}

func (c Client) Reboot() (transport.Result, error) {
	return c.Conn.Reboot(c.Address.GetSerialAddress())
}

func (c Client) Remount() (transport.Result, error) {
	result, err := c.Conn.Remount(c.Address.GetSerialAddress())
	return WaitAndReturn(&result, err, time.Duration(1)*time.Second)
}

func (c Client) Mount(dir string) (transport.Result, error) {
	result, err := c.Conn.Unmount(c.Address.GetSerialAddress(), dir)
	return WaitAndReturn(&result, err, time.Duration(1)*time.Second)
}

func (c Client) Unmount(dir string) (transport.Result, error) {
	result, err := c.Conn.Unmount(c.Address.GetSerialAddress(), dir)
	return WaitAndReturn(&result, err, time.Duration(1)*time.Second)
}

// BugReport Execute and return the result of the command 'adb bugreport'
// dst: optional target local folder/filename for the bugreport
func (c Client) BugReport(dst string) (transport.Result, error) {
	result, err := c.Conn.BugReport(c.Address.GetSerialAddress(), dst)
	return WaitAndReturn(&result, err, 0)
}

// Pull a file from the device.
// src is the file to be pulled from the device.
// dst is the destination filepath on the host.
func (c Client) Pull(src string, dst string) (transport.Result, error) {
	return c.Conn.Pull(c.Address.GetSerialAddress(), src, dst)
}

// Push a file to the connected device.
// src is the host file to be pushed.
// dst is the target device where the file should be pushed to.
func (c Client) Push(src string, dst string) (transport.Result, error) {
	return c.Conn.Push(c.Address.GetSerialAddress(), src, dst)
}

func (c Client) Install(src string, options *InstallOptions) (transport.Result, error) {
	var args []string
	if options != nil {
		if options.KeepData {
			args = append(args, "-r")
		}
		if options.AllowTestPackages {
			args = append(args, "-t")
		}
		if options.AllowDowngrade {
			args = append(args, "-d")
		}
		if options.GrantPermissions {
			args = append(args, "-g")
		}
	}
	return c.Conn.Install(c.Address.GetSerialAddress(), src, args...)
}

func (c Client) Uninstall(packageName string) (transport.Result, error) {
	return c.Conn.Uninstall(c.Address.GetSerialAddress(), packageName)
}

func (c Client) Logcat(options types.LogcatOptions) (transport.Result, error) {
	var args []string

	if options.Expr != "" {
		args = append(args, "-e", options.Expr)
	}

	if options.Dump {
		args = append(args, "-d")
	}

	if options.Filename != "" {
		args = append(args, options.Filename)
	}

	if options.Format != "" {
		args = append(args, "-v", options.Format)
	}

	if len(options.Pids) > 0 {
		args = append(args, "--pid")
		args = append(args, options.Pids...)
	}

	if len(options.Tags) > 0 {
		tags := streams.Map(options.Tags, func(tag types.LogcatTag) string {
			return tag.String()
		})
		args = append(args, tags...)
		args = append(args, "*:S")
	}

	if options.Since != nil {
		args = append(args, "-t")
		args = append(args, options.Since.Format("01-02 15:04:05.000"))
	}

	pb := c.NewProcess().WithArgs(args...).WithCommand("logcat")

	if options.Timeout > 0 {
		pb.WithTimeout(options.Timeout)
	}

	return pb.Invoke()
}

//
//
//

func (c Client) GetIsConnected() bool {
	result, err := c.IsConnected()
	if err != nil {
		return false
	}
	return result
}

func (c Client) GetIsRoot() bool {
	if c.GetIsConnected() {
		result, err := c.IsRoot()
		if err != nil {
			return false
		}
		return result
	}
	return false
}

func (c Client) MustRoot() bool {
	if c.GetIsConnected() {
		if c.GetIsRoot() {
			return true
		} else {
			_, err := c.Root()
			if err != nil {
				return false
			}
			return c.GetIsRoot()
		}
	}
	return false
}

type InstallOptions struct {
	// -r reinstall an existing app, keeping its data
	KeepData bool
	// -t allow test packages
	AllowTestPackages bool
	// -d allow version code downgrade
	AllowDowngrade bool
	// -g grant all runtime permissions
	GrantPermissions bool
}

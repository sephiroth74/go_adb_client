package adbclient

import (
	"fmt"
	"github.com/sephiroth74/go_adb_client/util"
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

func NewClient(device types.Serial) *Client {
	var conn = connection.NewConnection()
	client := new(Client)
	client.Conn = conn
	client.Mdns = mdns.NewMdns(client.Conn)
	client.Address = device
	client.Channel = make(chan rxgo.Item)
	client.Shell = shell.NewShell(&conn.ADBPath, device)
	return client
}

func NullClient() *Client {
	return NewClient(types.ClientAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5555})
}

func (c Client) NewProcess() *transport.ProcessBuilder {
	pb := transport.NewProcessBuilder(c.Address)
	pb.Path(&c.Conn.ADBPath)
	return pb
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

func (c Client) Connect() (transport.Result, error) {
	conn, err := c.IsConnected()
	if err == nil && conn {
		return transport.OkResult("Already Connected"), nil
	}
	result, err := c.Conn.Connect(c.Address.GetSerialAddress())

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

func (c Client) Reconnect() (transport.Result, error) {
	return c.Conn.Reconnect(c.Address.GetSerialAddress())
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
	return c.Conn.WaitForDevice(c.Address.GetSerialAddress())
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
	return c.Conn.Install(src, args...)
}

func (c Client) Uninstall(packageName string) (transport.Result, error) {
	return c.Conn.Uninstall(packageName)
}

func (c Client) Logcat(options LogcatOptions) (transport.Result, error) {
	args := []string{"logcat"}

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
		tags, _ := util.Map(options.Tags, func(tag LogcatTag) (string, error) {
			return tag.String(), nil
		})
		args = append(args, tags...)
		args = append(args, "*:S")
	}

	if options.Since != "" {
		args = append(args, options.Since)
	}

	return transport.Invoke(&c.Conn.ADBPath, 0, args...)
}

//
//
//

func (c Client) TryIsConnected() bool {
	result, err := c.IsConnected()
	if err != nil {
		return false
	}
	return result
}

func (c Client) TryIsRoot() bool {
	if c.TryIsConnected() {
		result, err := c.IsRoot()
		if err != nil {
			return false
		}
		return result
	}
	return false
}

func (c Client) TryRoot() bool {
	if c.TryIsConnected() {
		if c.TryIsRoot() {
			return true
		} else {
			_, err := c.Root()
			if err != nil {
				return false
			}
			return c.TryIsRoot()
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

type LogcatOptions struct {
	// -e Only prints lines where the log message matches <expr>, where <expr> is a regular expression.
	Expr string
	// -d	Dumps the log to the screen and exits.
	Dump bool
	// -f <filename>	Writes log message output to <filename>. The default is stdout.
	Filename string
	// -s	Equivalent to the filter expression '*:S', which sets priority for all tags to silent and is used to precede a list of filter expressions that add content.
	Tags []LogcatTag
	// -v <format>	Sets the output format for log messages. The default is the threadtime format
	Format string
	// -t '<time>'	Prints the most recent lines since the specified time. This option includes -d functionality. See the -P option for information about quoting parameters with embedded spaces.
	Since string
	// --pid=<pid> ...
	Pids []string
}

type LogcatLevel string

const (
	LogcatVerbose LogcatLevel = "V"
	LogcatDebug   LogcatLevel = "D"
	LogcatInfo    LogcatLevel = "I"
	LogcatWarn    LogcatLevel = "W"
	LogcatError   LogcatLevel = "E"
)

type LogcatTag struct {
	Name  string
	Level LogcatLevel
}

func (l LogcatTag) String() string {
	return fmt.Sprintf("%s:%s", l.Name, l.Level)
}

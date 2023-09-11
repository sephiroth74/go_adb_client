package adbclient

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sephiroth74/go-processbuilder"
	streams "github.com/sephiroth74/go_streams"

	"github.com/reactivex/rxgo/v2"
	"github.com/sephiroth74/go_adb_client/connection"
	"github.com/sephiroth74/go_adb_client/events"
	"github.com/sephiroth74/go_adb_client/logging"
	"github.com/sephiroth74/go_adb_client/mdns"
	"github.com/sephiroth74/go_adb_client/process"
	"github.com/sephiroth74/go_adb_client/shell"
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
	processbuilder.SetLogger(&logging.Log)
	return client
}

func NullClient(verbose bool) *Client {
	return NewClient(types.ClientAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5555}, verbose)
}

func (c Client) NewAdbCommand() *process.ADBCommand {
	return c.Conn.NewAdbCommand().WithSerialAddr(&c.Address)
}

func (c Client) DeferredDispatch(eventType events.EventType) {
	defer func() { go func() { c.Channel <- rxgo.Of(events.AdbEvent{Event: eventType}) }() }()
}

func (c Client) Dispatch(eventType events.EventType, data interface{}) {
	go func() { c.Channel <- rxgo.Of(events.AdbEvent{Event: eventType, Item: data}) }()
}

func WaitAndReturnOutput(result *process.OutputResult, err error, timeout time.Duration) (process.OutputResult, error) {
	if err != nil {
		return *result, err
	}
	time.Sleep(timeout)
	return *result, err
}

func (c Client) Connect(timeout time.Duration) (process.OutputResult, error) {
	if conn := c.GetIsConnected(); conn {
		return process.NewSuccessOutputResult("Already Connected"), nil
	}
	result, err := c.Conn.Connect(c.Address.GetSerialAddress(), timeout)

	if err != nil {
		return result, err
	}

	if conn := c.GetIsConnected(); !conn {
		return process.NewErrorOutputResult(fmt.Sprintf("Unable to connect to %s", c.Address.String())), nil
	} else {
		defer c.Dispatch(events.Connected, c.Address)
		return process.NewSuccessOutputResult(fmt.Sprintf("connected to %s", c.Address.String())), nil
	}
}

func (c Client) Reconnect(t types.ReconnectType, timeout time.Duration) (process.OutputResult, error) {
	return c.Conn.Reconnect(t, timeout)
}

func (c Client) IsConnected() (bool, error) {
	result, err := c.Conn.GetState(c.Address.GetSerialAddress())
	if err != nil {
		if result.HasError() {
			return false, nil
		}
		return false, err
	}
	return result.IsOk(), nil
}

func (c Client) Disconnect() (process.OutputResult, error) {
	if conn := c.GetIsConnected(); !conn {
		return process.NewSuccessOutputResult("already disconnected"), nil
	}

	result, err := c.Conn.Disconnect(c.Address.GetSerialAddress())

	if err == nil && result.IsOk() {
		defer c.Dispatch(events.Disconnect, c.Address)
	}

	return result, err
}

func (c Client) DisconnectAll() (process.OutputResult, error) {
	return c.Conn.DisconnectAll()
}

func (c Client) WaitForDevice(timeout time.Duration) (process.OutputResult, error) {
	return c.Conn.WaitForDevice(c.Address.GetSerialAddress(), timeout)
}

func (c Client) Root() error {
	result, err := c.Conn.Root(c.Address.GetSerialAddress())
	result, err = WaitAndReturnOutput(&result, err, time.Duration(1)*time.Second)
	if err != nil {
		return err
	}
	if !result.IsOk() {
		return result.NewError()
	}
	return nil
}

func (c Client) IsRoot() (bool, error) {
	return c.Conn.IsRoot(c.Address.GetSerialAddress())
}

func (c Client) UnRoot() error {
	result, err := c.Conn.UnRoot(c.Address.GetSerialAddress())
	result, err = WaitAndReturnOutput(&result, err, time.Duration(1)*time.Second)

	if err != nil {
		return err
	}

	if !result.IsOk() {
		return result.NewError()
	}

	return nil
}

func (c Client) ListDevices() ([]*types.Device, error) {
	return c.Conn.ListDevices()
}

func (c Client) Reboot() (process.OutputResult, error) {
	return c.Conn.Reboot(c.Address.GetSerialAddress())
}

func (c Client) Remount() (process.OutputResult, error) {
	result, err := c.Conn.Remount(c.Address.GetSerialAddress())
	return WaitAndReturnOutput(&result, err, time.Duration(1)*time.Second)
}

func (c Client) Mount(dir string) (process.OutputResult, error) {
	result, err := c.Conn.Mount(c.Address.GetSerialAddress(), dir)
	return WaitAndReturnOutput(&result, err, time.Duration(1)*time.Second)
}

func (c Client) Unmount(dir string) (process.OutputResult, error) {
	result, err := c.Conn.Unmount(c.Address.GetSerialAddress(), dir)
	return WaitAndReturnOutput(&result, err, time.Duration(1)*time.Second)
}

// BugReport ExecuteWithTimeout and return the result of the command 'adb bugreport'
// dst: optional target local folder/filename for the bugreport
func (c Client) BugReport(dst string) (process.OutputResult, error) {
	return c.Conn.BugReport(c.Address.GetSerialAddress(), dst)
}

// Pull a file from the device.
// src is the file to be pulled from the device.
// dst is the destination filepath on the host.
func (c Client) Pull(src string, dst string) (process.OutputResult, error) {
	return c.Conn.Pull(c.Address.GetSerialAddress(), src, dst)
}

// Push a file to the connected device.
// src is the host file to be pushed.
// dst is the target device where the file should be pushed to.
func (c Client) Push(src string, dst string) (process.OutputResult, error) {
	return c.Conn.Push(c.Address.GetSerialAddress(), src, dst)
}

func (c Client) Install(src string, options *InstallOptions) (process.OutputResult, error) {
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

func (c Client) Uninstall(packageName string) (process.OutputResult, error) {
	return c.Conn.Uninstall(c.Address.GetSerialAddress(), packageName)
}

func (c Client) ClearLogcat() error {
	_, err := process.SimpleOutput(c.NewAdbCommand().WithCommand("logcat").WithArgs("-b", "all", "-c"), c.Conn.Verbose)
	// _, err := c.NewProcess().WithCommand("logcat").WithArgs("-b", "all", "-c").Invoke()
	return err
}

func (c Client) Logcat(options types.LogcatOptions) (process.OutputResult, error) {
	var args []string

	if options.Filename != "" && options.File != nil {
		return process.OutputResult{}, errors.New("filename and file cannot be used togethere")
	}

	if options.Expr != "" {
		args = append(args, "-e", options.Expr)
	}

	if options.Dump {
		args = append(args, "-d")
	}

	if options.Filename != "" {
		args = append(args, "-f", options.Filename)
	}

	if options.Format != "" {
		args = append(args, "-v", options.Format)
	}

	if len(options.Pids) > 0 {
		args = append(args, "--pid")
		args = append(args, options.Pids...)
	}

	if options.Since != nil {
		args = append(args, "-T")
		args = append(args, options.Since.Format("01-02 15:04:05.000"))
	}

	if len(options.Tags) > 0 {
		tags := streams.Map(options.Tags, func(tag types.LogcatTag) string {
			return tag.String()
		})
		args = append(args, tags...)
		args = append(args, "*:S")
	}

	// pb := c.NewProcess().WithArgs(args...).WithCommand("logcat")
	cmd := c.NewAdbCommand().WithArgs(args...).WithCommand("logcat")

	if options.Timeout > 0 {
		cmd.WithTimeout(options.Timeout)
		// pb.WithTimeout(options.Timeout)
	}

	if options.File != nil {
		var writer io.Writer = bufio.NewWriter(options.File)
		// pb.WithStdout(&writer)
		cmd.WithStdOut(writer)
	}

	return process.SimpleOutput(cmd, c.Conn.Verbose)
	// return pb.Invoke()
}

func (c Client) LogcatPipe(options types.LogcatOptions) (*processbuilder.Processbuilder, error) {
	var args []string

	if options.Expr != "" {
		args = append(args, "-e", options.Expr)
	}

	if options.Format != "" {
		args = append(args, "-v", options.Format)
	}

	if len(options.Pids) > 0 {
		args = append(args, "--pid")
		args = append(args, options.Pids...)
	}

	if options.Since != nil {
		args = append(args, "-T")
		args = append(args, options.Since.Format("01-02 15:04:05.000"))
	}

	if len(options.Tags) > 0 {
		tags := streams.Map(options.Tags, func(tag types.LogcatTag) string {
			return tag.String()
		})
		args = append(args, fmt.Sprintf("%s *:S", strings.Join(tags, " ")))
	}

	pb := c.NewAdbCommand().WithArgs(args...).WithCommand("logcat")

	if options.Timeout > 0 {
		pb.WithTimeout(options.Timeout)
	}

	cmd := pb.ToCommand()

	p, err := processbuilder.PipeOutput(
		processbuilder.Option{Timeout: pb.Timeout},
		cmd,
	)

	if err != nil {
		return nil, err
	}

	return p, nil
}

func (c Client) GetMemInfo() (map[string]int, error) {
	result, err := c.Shell.Cat("/proc/meminfo")
	if err != nil {
		return nil, err
	}

	r := regexp.MustCompile(`^([\w\(\)]+):\s*([\d]+)\s*`)
	hashmap := make(map[string]int)

	lines := result.OutputLines(true)
	for _, line := range lines {
		m := r.FindStringSubmatch(line)
		if len(m) == 3 {
			intVar, err := strconv.Atoi(m[2])
			if err != nil {
				return nil, err
			}
			hashmap[m[1]] = intVar
		}
	}

	return hashmap, nil
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
			err := c.Root()
			if err != nil {
				return false
			}
			return c.GetIsRoot()
		}
	}
	return false
}

func (c Client) DisableVerity() (string, error) {
	return c.toggleVerity(false)
}

func (c Client) EnableVerity() (string, error) {
	return c.toggleVerity(true)
}

func (c Client) toggleVerity(enabled bool) (string, error) {
	if !c.GetIsConnected() {
		return "", errors.New("not connected")
	}

	if !c.GetIsRoot() {
		return "", errors.New("must be root")
	}

	var cmd string

	if enabled {
		cmd = "enable-verity"
	} else {
		cmd = "disable-verity"
	}

	output, err := process.SimpleOutput(c.NewAdbCommand().WithCommand(cmd), c.Conn.Verbose)
	return output.Output(), err
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

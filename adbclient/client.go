package adbclient

import (
	"fmt"
	"time"

	"github.com/reactivex/rxgo/v2"
	"it.sephiroth/adbclient/activitymanager"
	"it.sephiroth/adbclient/connection"
	"it.sephiroth/adbclient/events"
	"it.sephiroth/adbclient/mdns"
	"it.sephiroth/adbclient/shell"
	"it.sephiroth/adbclient/transport"
	"it.sephiroth/adbclient/types"
)

type Client[T types.Serial] struct {
	Conn    *connection.Connection
	Mdns    *mdns.Mdns
	Channel chan rxgo.Item
	Serial  T
}

func NewClient[T types.Serial](device T) *Client[T] {
	client := new(Client[T])
	client.Conn = connection.NewConnection()
	client.Mdns = mdns.NewMdns(client.Conn)
	client.Serial = device
	client.Channel = make(chan rxgo.Item)
	return client
}

func (c Client[T]) ActivityManager() *activitymanager.ActivityManager[T] {
	return &activitymanager.ActivityManager[T]{
		Shell: c.Shell(),
	}
}

func (c Client[T]) Shell() *shell.Shell[T] {
	var s = shell.NewShell(&c.Conn.ADBPath, c.Serial)
	return s
}

func (c Client[T]) DeferredDispatch(eventType events.EventType) {
	defer func() { go func() { c.Channel <- rxgo.Of(events.AdbEvent{Event: eventType}) }() }()
}

func (c Client[T]) Dispatch(eventType events.EventType, data interface{}) {
	go func() { c.Channel <- rxgo.Of(events.AdbEvent{Event: eventType, Item: data}) }()
}

func WaitAndReturn(result *transport.Result, err error, timeout time.Duration) (transport.Result, error) {
	if err != nil {
		return *result, err
	}
	time.Sleep(timeout)
	return *result, err
}

func (c Client[T]) Connect() (transport.Result, error) {
	conn, err := c.IsConnected()
	if err == nil && conn {
		return transport.OkResult("Already Connected"), nil
	}
	result, err := c.Conn.Connect(c.Serial.Serial())

	if err != nil {
		return transport.ErrorResult(result.Output()), err
	}

	conn, err = c.IsConnected()
	if err != nil {
		return transport.ErrorResult("Unable to connect"), err
	}

	if conn {
		defer c.Dispatch(events.Connected, c.Serial)
		return transport.OkResult(fmt.Sprintf("connected to %s", c.Serial.String())), nil
	} else {
		return transport.ErrorResult(fmt.Sprintf("Unable to connect to %s", c.Serial.String())), nil
	}
}

func (c Client[T]) Reconnect() (transport.Result, error) {
	return c.Conn.Reconnect(c.Serial.Serial())
}

func (c Client[T]) IsConnected() (bool, error) {
	result, err := c.Conn.GetState(c.Serial.Serial())
	if err != nil {
		return false, err
	}
	return result.IsOk(), nil
}

func (c Client[T]) Disconnect() (transport.Result, error) {
	connected, err := c.IsConnected()
	if err == nil && !connected {
		return transport.OkResult(""), nil
	}

	result, err := c.Conn.Disconnect(c.Serial.Serial())

	if err == nil && result.IsOk() {
		defer c.Dispatch(events.Disconnect, c.Serial)
	}

	return result, err
}

func (c Client[T]) DisconnectAll() (transport.Result, error) {
	return c.Conn.DisconnectAll()
}

func (c Client[T]) WaitForDevice() (transport.Result, error) {
	return c.Conn.WaitForDevice(c.Serial.Serial())
}

func (c Client[T]) WaitForDeviceWithTimeout(timeout time.Duration) (transport.Result, error) {
	return c.Conn.WaitForDeviceWithTimeout(c.Serial.Serial(), timeout)
}

func (c Client[T]) Root() (transport.Result, error) {
	result, err := c.Conn.Root(c.Serial.Serial())
	return WaitAndReturn(&result, err, time.Duration(1)*time.Second)
}

func (c Client[T]) IsRoot() (bool, error) {
	return c.Conn.IsRoot(c.Serial.Serial())
}

func (c Client[T]) UnRoot() (transport.Result, error) {
	result, err := c.Conn.UnRoot(c.Serial.Serial())
	return WaitAndReturn(&result, err, time.Duration(1)*time.Second)
}

func (c Client[T]) ListDevices() ([]*types.Device, error) {
	return c.Conn.ListDevices()
}

func (c Client[T]) Reboot() (transport.Result, error) {
	return c.Conn.Reboot(c.Serial.Serial())
}

func (c Client[T]) Remount() (transport.Result, error) {
	result, err := c.Conn.Remount(c.Serial.Serial())
	return WaitAndReturn(&result, err, time.Duration(1)*time.Second)
}

func (c Client[T]) Mount(dir string) (transport.Result, error) {
	result, err := c.Conn.Unmount(c.Serial.Serial(), dir)
	return WaitAndReturn(&result, err, time.Duration(1)*time.Second)
}

func (c Client[T]) Unmount(dir string) (transport.Result, error) {
	result, err := c.Conn.Unmount(c.Serial.Serial(), dir)
	return WaitAndReturn(&result, err, time.Duration(1)*time.Second)
}

// Execute and return the result of the command 'adb bugreport'
// dst: optional target local folder/filename for the bugreport
func (c Client[T]) BugReport(dst string) (transport.Result, error) {
	result, err := c.Conn.BugReport(c.Serial.Serial(), dst)
	return WaitAndReturn(&result, err, 0)
}

// Pull a file from the device.
// src is the file to be pulled from the device.
// dst is the destination filepath on the host.
func (c Client[T]) Pull(src string, dst string) (transport.Result, error) {
	return c.Conn.Pull(c.Serial.Serial(), src, dst)
}

// Push a file to the connected device.
// src is the host file to be pushed.
// dst is the target device where the file should be pushed to.
func (c Client[T]) Push(src string, dst string) (transport.Result, error) {
	return c.Conn.Push(c.Serial.Serial(), src, dst)
}

//
//
//

func (c Client[T]) TryIsConnected() bool {
	result, err := c.IsConnected()
	if err != nil {
		return false
	}
	return result
}

func (c Client[T]) TryIsRoot() bool {
	if c.TryIsConnected() {
		result, err := c.IsRoot()
		if err != nil {
			return false
		}
		return result
	}
	return false
}

func (c Client[T]) TryRoot() bool {
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

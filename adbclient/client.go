package adbclient

import (
	"fmt"
	"time"

	"it.sephiroth/adbclient/connection"
	"it.sephiroth/adbclient/mdns"
	"it.sephiroth/adbclient/transport"
	"it.sephiroth/adbclient/types"
)

type Client[T types.Serial] struct {
	Conn   *connection.Connection
	Mdns   *mdns.Mdns
	Serial T
}

func NewClientSerial[T types.Serial](device T) *Client[T] {
	client := new(Client[T])
	client.Conn = connection.NewConnection()
	client.Mdns = mdns.NewMdns(client.Conn)
	client.Serial = device
	return client
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
		return transport.ErrorResult(result.GetOutput()), err
	}

	conn, err = c.IsConnected()

	if err != nil {
		return transport.ErrorResult("Unable to connect"), err
	}

	if conn {
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
	return c.Conn.Disconnect(c.Serial.Serial())
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

// Pull a file(s) from the device
func (c Client[T]) Pull(src string, dst string) (transport.Result, error) {
	return c.Conn.Pull(c.Serial.Serial(), src, dst)
}

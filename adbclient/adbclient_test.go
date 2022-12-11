package adbclient_test

import (
	"net"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/alecthomas/repr"
	"github.com/magiconair/properties"
	"github.com/reactivex/rxgo/v2"
	"github.com/stretchr/testify/assert"

	"it.sephiroth/adbclient"
	"it.sephiroth/adbclient/connection"
	"it.sephiroth/adbclient/logging"
	"it.sephiroth/adbclient/mdns"
	"it.sephiroth/adbclient/types"
)

var device_ip1 = net.IPv4(192, 168, 1, 105)
var device_ip2 = net.IPv4(192, 168, 1, 123)
var device_ip = device_ip2

var log = logging.GetLogger("test")

func NewClient() *adbclient.Client[types.ClientAddr] {
	return adbclient.NewClient(types.ClientAddr{IP: device_ip, Port: 5555})
}

func AssertClientConnected[T types.Serial](t *testing.T, client *adbclient.Client[T]) {
	result, err := client.Connect()
	assert.Nil(t, err)
	assert.True(t, result.IsOk(), result.Output())
}

func TestIsConnected(t *testing.T) {
	var client = adbclient.NewClient(types.ClientAddr{IP: device_ip, Port: 5555})
	result, err := client.Connect()
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	conn, err := client.IsConnected()
	assert.Nil(t, err)
	assert.True(t, conn)
}

func TestConnect(t *testing.T) {
	var client = adbclient.NewClient(types.ClientAddr{IP: device_ip, Port: 5555})

	result, err := client.Connect()
	log.Debugf("result=%s", result.ToString())
	assert.Nil(t, err)
	assert.Equal(t, true, result.IsOk())

	conn, err := client.IsConnected()
	assert.Nil(t, err)
	assert.Equal(t, true, conn)

	result, err = client.Disconnect()
	assert.Nil(t, err)
	assert.Equal(t, true, result.IsOk())

	conn, err = client.IsConnected()
	assert.Nil(t, err)
	assert.False(t, conn)

	result, err = client.DisconnectAll()
	assert.Nil(t, err)
	assert.Equal(t, true, result.IsOk())
}

func TestWaitForDevice(t *testing.T) {
	var client = adbclient.NewClient(types.ClientAddr{IP: device_ip, Port: 5555})

	result, err := client.Connect()
	assert.Nil(t, err)
	assert.Equal(t, true, result.IsOk())

	result, err = client.WaitForDevice()
	assert.Nil(t, err)
	println(result.ToString())
	assert.Equal(t, true, result.IsOk())

	connected, err := client.IsConnected()
	assert.Nil(t, err)
	assert.Equal(t, true, connected)

	is_root, err := client.IsRoot()
	assert.Nil(t, err)
	assert.True(t, is_root)

	result, err = client.Disconnect()
	assert.Nil(t, err)
	assert.Equal(t, true, result.IsOk())
}

func TestRoot(t *testing.T) {
	var client = adbclient.NewClient(types.ClientAddr{IP: device_ip, Port: 5555})
	result, err := client.Connect()
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	result, err = client.Root()
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	value, err := client.IsRoot()
	assert.Nil(t, err)
	assert.True(t, value)

	result, err = client.UnRoot()
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	value, err = client.IsRoot()
	assert.Nil(t, err)
	assert.False(t, value)
}

func TestListDevices(t *testing.T) {
	var client = adbclient.NewClient(types.ClientAddr{IP: device_ip, Port: 5555})
	result, err := client.Connect()
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	list, err := client.ListDevices()

	if err == nil {
		for x := 0; x < len(list); x++ {
			log.Debug("Device: %#v\n", list[x])
		}
	}
}

func TestReboot(t *testing.T) {
	var client = adbclient.NewClient(types.ClientAddr{IP: device_ip, Port: 5555})
	result, err := client.Connect()
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	conn, err := client.IsConnected()
	assert.Nil(t, err)
	assert.True(t, conn)

	result, err = client.Reboot()
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	result, err = client.WaitForDeviceWithTimeout(time.Duration(2) * time.Minute)
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	conn, err = client.IsConnected()
	assert.Nil(t, err)
	assert.True(t, conn)
}

func TestRemount(t *testing.T) {
	var client = adbclient.NewClient(types.ClientAddr{IP: device_ip, Port: 5555})
	result, err := client.Connect()
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	conn, err := client.IsConnected()
	assert.Nil(t, err)
	assert.True(t, conn)

	result, err = client.Root()
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	conn, err = client.IsRoot()
	assert.Nil(t, err)
	assert.True(t, conn)

	result, err = client.Remount()
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	result, err = client.Unmount("/system")
	if err != nil {
		log.Warningf("error=%#v\n", err.Error())
	}
	assert.Nil(t, err)
	assert.True(t, result.IsOk())
}

func TestGetVersion(t *testing.T) {
	conn := connection.NewConnection()
	result, err := conn.Version()
	assert.Nil(t, err)
	assert.NotEmpty(t, result)
	log.Debugf("adb version=%s", result)
}

func TestMdns(t *testing.T) {
	var mdns = mdns.NewMdns(connection.NewConnection())
	result, err := mdns.Check()
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	devices, err := mdns.Services()
	assert.Nil(t, err)

	log.Debugf("Found %d devices", len(devices))

	for i := 0; i < len(devices); i++ {
		log.Debugf("device: %#v", devices[i])
	}

	assert.True(t, len(devices) > 0)

	client2 := adbclient.NewClient(devices[1])
	result, err = client2.Connect()
	assert.Nil(t, err)
	log.Debug(result)

	value, err := client2.IsConnected()
	assert.Nil(t, err)
	assert.True(t, value)
}

func TestBugReport(t *testing.T) {
	var client = NewClient()
	result, err := client.BugReport("")
	assert.Nil(t, err)
	assert.True(t, result.IsOk())
}

func TestPull(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	client.Root()

	path, err := filepath.Abs("./export")
	assert.Nil(t, err)

	result, err := client.Pull("/data/data/com.swisscom.aot.library.sample", path)

	if err != nil {
		log.Error(err.Error())
	}

	log.Debugf("output: %s", result.Output())
	log.Debugf("error: %s", result.Error())

	assert.Nil(t, err)
	assert.True(t, result.IsOk(), result.Output())
	client.UnRoot()
	client.DisconnectAll()
}

func TestPush(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	result, err := client.Push("../README.md", "/sdcard/Download")
	assert.Nil(t, err)
	assert.Truef(t, result.IsOk(), "result: %s", result.Repr())

	client.Disconnect()
}

func TestWhich(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	result, err := client.Shell().Which("which")
	assert.Nil(t, err)

	println(result.Repr())
}

func TestRx(t *testing.T) {
	var client = NewClient()

	observable := rxgo.FromEventSource(client.Channel)

	observable.DoOnNext(func(i interface{}) {
		log.Info("onNext:", repr.String(i))
	})

	observable.DoOnCompleted(func() {
		log.Info("onComplete")
	})

	observable.DoOnError(func(err error) {
		log.Info("onError:", err.Error())
	})

	client.Disconnect()
	client.Connect()
	client.IsConnected()
	// client.Disconnect()

	client = nil
	runtime.GC()

	time.Sleep(time.Duration(2) * time.Second)
}

func TestActivityManager(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	var intent = types.NewIntent()
	intent.Action = "android.action.View"
	intent.Extra.Es["key1"] = "string1"
	intent.Extra.Es["key2"] = "string2"
	intent.Extra.Eia["key_eia1"] = []int{1, 2, 3}

	client.ActivityManager().Broadcast(intent)
}

func TestShellCat(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	assert.True(t, client.TryRoot())

	result, err := client.Shell().Whoami()
	assert.Nil(t, err)
	assert.Equal(t, "root", result.Output())

	result, err = client.Shell().Cat("/system/build.prop")
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	props, err := properties.Load(result.Stdout, properties.UTF8)
	assert.Nil(t, err)

	for _, k := range props.Keys() {
		v, ok := props.Get(k)

		if ok {
			log.Debugf("%s = %s", k, v)
		} else {
			log.Warningf("Error reading key %s", k)
		}
	}

	props.Set("ro.config.system_vol_default", "10")
	assert.Equal(t, 10, props.GetInt("ro.config.system_vol_default", 0))
}

func TestShellGetProp(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	shell := client.Shell()
	prop := shell.GetProp("wlan.driver.status")
	assert.NotNil(t, prop)
	assert.Equal(t, "ok", *prop)

	prop = shell.GetProp("invalid.key.string")
	assert.Nil(t, prop)
}


func TestShellGetProps(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	shell := client.Shell()
	props, err := shell.GetProps()
	assert.Nil(t, err)
	assert.True(t, len(props) > 0)

	for _, v := range props {
		log.Debugf("%s=%s", v.First, v.Second)
	}
}

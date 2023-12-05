package adbclient_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/sephiroth74/go-processbuilder"
	"github.com/sephiroth74/go_adb_client/process"
	"github.com/sephiroth74/go_adb_client/scanner"
	"github.com/sephiroth74/go_adb_client/shell"
	streams "github.com/sephiroth74/go_streams"
	"golang.org/x/sys/unix"
	"pkg.re/essentialkaos/ek.v12/path"

	"github.com/alecthomas/repr"
	"github.com/magiconair/properties"
	"github.com/reactivex/rxgo/v2"
	"github.com/stretchr/testify/assert"

	adbclient "github.com/sephiroth74/go_adb_client"
	"github.com/sephiroth74/go_adb_client/connection"
	"github.com/sephiroth74/go_adb_client/input"
	"github.com/sephiroth74/go_adb_client/logging"
	"github.com/sephiroth74/go_adb_client/mdns"
	"github.com/sephiroth74/go_adb_client/packagemanager"
	"github.com/sephiroth74/go_adb_client/types"
	"gopkg.in/pipe.v2"
)

var device_ip2 = net.IPv4(192, 168, 1, 101)
var device_ip = device_ip2

var local_apk = "~/ArcCustomizeSettings.apk"

func init() {
}

func NewClient() *adbclient.Client {
	return adbclient.NewClient(types.ClientAddr{IP: device_ip, Port: 5555}, logging.Log, true)
}

func AssertClientConnected(t *testing.T, client *adbclient.Client) {
	logger := log.New(os.Stderr)
	logger.Warn("chewy!", "butter", true)

	logging.Log.Info("Hello World!")

	result, err := client.Connect(5 * time.Second)
	assert.Nil(t, err, "Error connecting to %s", client.Address.String())
	assert.True(t, result.IsOk(), "Error: %s", result.Error())
}

func TestReconnect(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	result, err := client.Reconnect(types.ReconnectToDevice, 5*time.Second)
	assert.Nil(t, err)
	assert.True(t, result.IsOk())
	fmt.Println(result.String())
}

func TestDisableVerity(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	err := client.Root()
	assert.Nil(t, err)

	output, err := client.DisableVerity()
	assert.Nil(t, err)

	fmt.Println(output)
}

func TestGetMemInfo(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	result, err := client.GetMemInfo()
	assert.Nil(t, err)

	userData, err := json.Marshal(result)
	assert.Nil(t, err)

	println(string(userData))
}

func TestEnableVerity(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	_, err := client.IsRoot()
	assert.Nil(t, err)

	output, err := client.EnableVerity()
	assert.Nil(t, err)

	fmt.Println(output)
}

func TestIsRoot(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	_, err := client.IsRoot()
	assert.Nil(t, err)

	err = client.UnRoot()
	assert.Nil(t, err)

	value, err := client.IsRoot()
	assert.Nil(t, err)
	assert.False(t, value)

	err = client.Root()
	assert.Nil(t, err)

	value, err = client.IsRoot()
	assert.Nil(t, err)
	assert.True(t, value)
}

func TestGetState(t *testing.T) {
	var client = NewClient()
	res, err := client.IsConnected()
	assert.Nil(t, err)
	assert.True(t, res)
}

func TestWithPipe(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	p1 := pipe.Exec("adb", "logcat")
	p := pipe.Line(
		p1,
		pipe.Write(os.Stdout),
	)

	err := pipe.Run(p)
	assert.Nil(t, err)
}

func TestBugReport(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	dst := "/Users/alessandro/Desktop/AndroidTV"
	report, err := client.BugReport(dst)

	assert.Nil(t, err)
	assert.True(t, report.IsOk())
}

func TestStdout(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	s := bytes.NewBufferString("")
	var w io.Writer = bufio.NewWriter(s)

	result, err := process.SimpleOutput(client.Shell.NewCommand().WithArgs("nproc").WithStdOut(w), true)

	assert.Nil(t, err)
	assert.True(t, result.IsOk())
	assert.True(t, result.Output() == "")
	assert.True(t, result.Error() == "")
	assert.True(t, s.Len() > 0)

	fmt.Println(result.String())
	fmt.Println(s)
}

func TestIsConnected(t *testing.T) {
	var client = NewClient()
	defer client.Disconnect()
	result, err := client.Connect(5 * time.Second)
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	conn, err := client.IsConnected()
	assert.Nil(t, err)
	assert.True(t, conn)
}

func TestActivityManagerForceStop(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	device := adbclient.NewDevice(client)
	client.Root()

	err := device.ActivityManager().ForceStop("com.google.android.tvlauncher")
	assert.Nil(t, err)

	if err != nil {
		fmt.Println(err.Error())
	}
}

func TestDirWalk(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	var device = adbclient.NewDevice(client)
	err := device.Client.Root()
	assert.Nil(t, err)

	dirname := "/sdcard/"
	result, err := device.Client.Shell.ListDir(dirname)
	assert.Nil(t, err)

	logging.Log.Infof("result of listdir: %s", dirname)
	for _, line := range result {
		logging.Log.Info(line.String())
	}

}

func TestRecordScreen(t *testing.T) {
	var client = NewClient()
	defer client.Disconnect()
	AssertClientConnected(t, client)

	var device = adbclient.NewDevice(client)

	var err error

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	defer close(c)

	pb, err := device.Client.Shell.ScreenRecord(shell.ScreenRecordOptions{
		Bitrate:   8000000,
		Timelimit: 5,
		Rotate:    false,
		BugReport: false,
		Size:      &types.Size{Width: 1920, Height: 1080},
		Verbose:   true,
	}, "/sdcard/Download/screenrecord.mp4")

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	fmt.Println("Run...")
	code, _, err := processbuilder.Run(pb)

	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	fmt.Printf("code: %v\n", code)

	if code != int(unix.SIGINT) {
		fmt.Println(err.Error())
	}

	go func() {
		<-c
		fmt.Println("user: cancelled")
		processbuilder.Cancel(pb)
	}()

	// fmt.Println("Wait..")
	// code, _, err = processbuilder.Wait(pb)

	// if err != nil {
	// 	assert.Equal(t, int(unix.SIGINT), code)
	// }

	// fmt.Printf("code: %v\n", code)

	// if code != int(unix.SIGINT) {
	// 	fmt.Println(err.Error())
	// }

	fmt.Println("done.")
}

func TestConnect(t *testing.T) {
	var client = NewClient()

	result, err := client.Connect(5 * time.Second)
	logging.Log.Debugf("result=%s", result.String())
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
	var client = NewClient()
	result, err := client.Connect(5 * time.Second)
	assert.Nil(t, err)
	assert.Equal(t, true, result.IsOk())

	result, err = client.WaitForDevice(1 * time.Minute)
	assert.Nil(t, err)
	println(result.String())
	assert.Equal(t, true, result.IsOk())

	connected, err := client.IsConnected()
	assert.Nil(t, err)
	assert.Equal(t, true, connected)

	result, err = client.Disconnect()
	assert.Nil(t, err)
	assert.Equal(t, true, result.IsOk())
}

func TestRoot(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	err := client.Root()
	assert.Nil(t, err)

	value, err := client.IsRoot()
	assert.Nil(t, err)
	assert.True(t, value)

	err = client.UnRoot()
	assert.Nil(t, err)

	value, err = client.IsRoot()
	assert.Nil(t, err)
	assert.False(t, value)
}

func TestListDevices(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)
	list, err := client.ListDevices()

	if err == nil {
		for x := 0; x < len(list); x++ {
			logging.Log.Debugf("Device: %s\n", list[x].String())
		}
	}
}

func TestReboot(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	conn, err := client.IsConnected()
	assert.Nil(t, err)
	assert.True(t, conn)

	result, err := client.Reboot()
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	result, err = client.WaitForDevice(2 * time.Minute)
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	conn, err = client.IsConnected()
	assert.Nil(t, err)
	assert.True(t, conn)
}

func TestRemount(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	conn, err := client.IsConnected()
	assert.Nil(t, err)
	assert.True(t, conn)

	err = client.Root()
	assert.Nil(t, err)

	conn, err = client.IsRoot()
	assert.Nil(t, err)
	assert.True(t, conn)

	result, err := client.Remount()
	assert.Nil(t, err)
	assert.True(t, result.IsOk())
	logging.Log.Debug(result.String())

	result, err = client.Unmount("/system")
	if err != nil {
		logging.Log.Warnf("error=%#v\n", err.Error())
	}
	assert.Nil(t, err)
	assert.True(t, result.IsOk())
	logging.Log.Debug(result.String())
}

func TestGetVersion(t *testing.T) {
	conn := connection.NewConnection(true)
	result, err := conn.Version()
	assert.Nil(t, err)
	assert.NotEmpty(t, result)
	logging.Log.Debugf("adb version=%s", result)
}

func TestMdnsServices(t *testing.T) {
	var mdns = mdns.NewMdns(connection.NewConnection(true))
	result, err := mdns.Check()
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	devices, err := mdns.Services()
	assert.Nil(t, err)
	logging.Log.Debugf("Found %d devices", len(devices))

	for i := 0; i < len(devices); i++ {
		logging.Log.Debugf("device: %#v", devices[i])
	}

	assert.True(t, len(devices) > 0)

	client2 := adbclient.NewClient(devices[1], logging.Log, true)
	result2, err := client2.Connect(5 * time.Second)
	assert.Nil(t, err)
	logging.Log.Debug(result2.String())

	value, err := client2.IsConnected()
	assert.Nil(t, err)
	assert.True(t, value)
}

func TestPull(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	client.Root()

	path, err := filepath.Abs("./export")
	assert.Nil(t, err)

	result, err := client.Pull("/system/build.prop", path)
	if err != nil {
		logging.Log.Debugf("result: %s", result.Error())
	}

	assert.Nil(t, err)
	assert.True(t, result.IsOk(), result.Output())

	err = client.UnRoot()
	assert.Nil(t, err)

	_, err = client.DisconnectAll()
	assert.Nil(t, err)
}

func TestPush(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	result, err := client.Push("./README.md", "/sdcard/Download")
	assert.Nil(t, err)
	assert.Truef(t, result.IsOk(), "result: %s", result.String())
	client.Disconnect()
}

func TestWhich(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	result, err := client.Shell.Which("which")
	assert.Nil(t, err)
	assert.True(t, len(result) > 0)

	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(result)

	result2, err := client.Shell.Conn.Which(client.Address.GetSerialAddress(), "which")
	assert.Nil(t, err)
	assert.True(t, result2 == result)

	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(result2)

}

func TestRx(t *testing.T) {
	var client = NewClient()

	observable := rxgo.FromEventSource(client.Channel)

	observable.DoOnNext(func(i interface{}) {
		logging.Log.Infof("onNext: %s", repr.String(i))
	})

	observable.DoOnCompleted(func() {
		logging.Log.Info("onComplete")
	})

	observable.DoOnError(func(err error) {
		logging.Log.Infof("onError: %s", err.Error())
	})

	client.Disconnect()
	client.Connect(5 * time.Second)
	client.IsConnected()

	client = nil
	runtime.GC()
	time.Sleep(2 * time.Second)
}

func TestSendBroadcast(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	var device = adbclient.NewDevice(client)

	var intent = types.NewIntent()
	intent.Action = "android.action.View"
	intent.Wait = true
	intent.Extra.Es["key1"] = "string1"
	intent.Extra.Es["key2"] = "string2"
	intent.Extra.Eia["key_eia1"] = []int{1, 2, 3}

	result, err := device.ActivityManager().Broadcast(intent)
	assert.Nil(t, err)
	assert.True(t, result.IsOk(), result.String())

	println(result.String())
}

func TestShellCat(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	assert.True(t, client.MustRoot())

	result, err := client.Shell.Whoami()
	assert.Nil(t, err)
	assert.Equal(t, "root", result.Output())
	println(result.String())

	result, err = client.Shell.Cat("/system/build.prop")
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	props, err := properties.Load(result.StdOut.Bytes(), properties.UTF8)
	assert.Nil(t, err)

	for _, k := range props.Keys() {
		v, ok := props.Get(k)

		if ok {
			logging.Log.Debugf("%s = %s", k, v)
		} else {
			logging.Log.Warnf("Error reading key %s", k)
		}
	}

	props.Set("ro.config.system_vol_default", "10")
	assert.Equal(t, 10, props.GetInt("ro.config.system_vol_default", 0))
}

func TestShellGetProp(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	shell := client.Shell
	prop := shell.GetProp("ro.build.product")
	assert.NotNil(t, prop)
	assert.True(t, len(*prop) > 0)

	prop2, err := shell.GetPropValue("ro.build.product")
	assert.Nil(t, err)
	assert.True(t, prop2 != "")
	assert.True(t, prop2 == *prop)

	logging.Log.Debugf("ro.build.product -> %s\n", *prop)

	prop = shell.GetProp("invalid.key.string")
	assert.Nil(t, prop)
}

func TestListDumpSys(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	result, err := client.Shell.ListDumpSys()
	assert.Nil(t, err)
	assert.True(t, len(result) > 0)

	for _, line := range result {
		fmt.Printf("line -> %s\n", line)
	}
}

func TestShellGetProps(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	shell := client.Shell
	props, err := shell.GetProps()
	println(props)
	assert.Nil(t, err)
	assert.True(t, len(props.Keys()) > 0)

	for _, v := range props.Keys() {
		value := props.GetString(v, "")
		logging.Log.Debugf("%s=%s", v, value)
	}
}

func TestShellGetPropsType(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	shell := client.Shell
	props, err := shell.GetProps()
	assert.Nil(t, err)
	assert.True(t, len(props.Keys()) > 0)

	for _, v := range props.Keys() {
		pt, ok := shell.GetPropType(v)
		assert.Truef(t, ok, "Error getting type of key %s", v)
		if ok {
			logging.Log.Debugf("%s=%s", v, *pt)
		}
	}
}

func TestDevice(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	device := adbclient.Device{Client: client}
	deviceName := device.Name()
	apiLevel := device.ApiLevel()
	version := device.Version()

	assert.NotNil(t, deviceName)
	assert.NotNil(t, apiLevel)
	assert.NotNil(t, version)

	logging.Log.Infof("device name: %s", *deviceName)
	logging.Log.Infof("device api level: %s", *apiLevel)
	logging.Log.Infof("device version release: %s", *version)
}

func TestShellSetProp(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	shell := client.Shell
	prop := shell.GetProp("dalvik.vm.heapsize")
	assert.NotNil(t, prop)

	assert.True(t, *prop != "" && (*prop == "256m" || *prop == "512m"))

	ok := shell.SetProp("dalvik.vm.heapsize", "512m")
	assert.True(t, ok)

	prop = shell.GetProp("dalvik.vm.heapsize")
	assert.NotNil(t, prop)
	assert.Equal(t, "512m", *prop)

	ok = shell.SetProp("dalvik.vm.heapsize", "512m")
	assert.True(t, ok)

	ok = shell.SetProp("debug.hwui.overdraw", "")
	assert.True(t, ok)

	prop = shell.GetProp("debug.hwui.overdraw")
	assert.Equal(t, "", *prop)
}

func TestWriteScreenCap(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	target_file, err := filepath.Abs("./exports/screencap.png")
	assert.Nil(t, err)

	var target_dir = filepath.Dir(target_file)

	logging.Log.Infof("target file: %s", target_file)
	logging.Log.Infof("target dir: %s", target_dir)

	os.RemoveAll(target_dir)
	os.MkdirAll(target_dir, 0755)

	_, err = os.Stat(target_dir)
	assert.Nil(t, err)

	f, err := os.Create(target_file)
	assert.Nil(t, err)
	os.Chmod(target_file, 0755)

	device := adbclient.NewDevice(client)
	result, err := device.WriteScreenCap(f)
	assert.Nil(t, err)

	if err != nil {
		logging.Log.Error(err.Error())
		logging.Log.Error(result.Error())
	}

	assert.True(t, result.IsOk())
	if _, err := os.Stat(target_file); errors.Is(err, os.ErrNotExist) {
		assert.Fail(t, "Screencap not exported")
	}

	os.RemoveAll(target_dir)
}

func TestFileExists(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	if err := client.Root(); err != nil {
		assert.Fail(t, "failed to root")
	}

	exists := client.Shell.Exists("/system/build.prop")
	assert.True(t, exists)
}

func TestSaveScreenCap(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	var target_file = "/sdcard/Download/screencap.png"

	device := adbclient.NewDevice(client)
	result, err := device.SaveScreenCap(target_file)
	assert.Nil(t, err)

	exists := client.Shell.Exists(target_file)
	assert.True(t, exists)

	if err != nil {
		logging.Log.Error(err.Error())
		logging.Log.Error(result.Error())
	}

	value := client.Shell.Exists(target_file)
	assert.True(t, value)

	value = client.Shell.IsFile(target_file)
	assert.True(t, value)

	value = client.Shell.IsDir(target_file)
	assert.False(t, value)

	value = client.Shell.IsSymlink(target_file)
	assert.False(t, value)

	value, err = client.Shell.Remove(target_file, false)
	assert.Nil(t, err)
	assert.True(t, value, "file not removed")
}

func TestListPackages(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	defer client.Disconnect()

	device := adbclient.NewDevice(client)
	pm := device.PackageManager()

	// system apps
	packages, err := pm.ListPackages(packagemanager.PackageOptions{
		ShowOnlyEnabed: true,
		ShowOnlySystem: true,
	})
	assert.Nil(t, err)

	for _, p := range packages {
		logging.Log.Debugf("%s, uid:%s", p.Name, p.UID)
		assert.True(t, p.Filename != "")
		assert.True(t, p.Name != "")
		assert.True(t, p.VersionCode != "")
		assert.True(t, p.UID != "")
	}
}

func TestFindPackages(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	defer client.Disconnect()

	device := adbclient.NewDevice(client)
	pm := device.PackageManager()

	packages, err := pm.ListPackagesWithFilter(packagemanager.PackageOptions{ShowOnlySystem: true}, "com.google")
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(packages), 1)

	for _, p := range packages {
		assert.True(t, p.Filename != "")
		assert.True(t, strings.HasPrefix(p.Name, "com.google"))
		assert.True(t, p.VersionCode != "")
		assert.True(t, p.UID != "")
		assert.True(t, p.MaybeIsSystem())
		logging.Log.Debugf("%s, uid:%s", p.Name, p.UID)
	}
}

func TestIsSystem(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	device := adbclient.NewDevice(client)
	pm := device.PackageManager()

	is_system, _ := pm.IsSystem("com.android.tv.settings")
	assert.True(t, is_system)
}

func TestSendKey(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)
	result, err := client.Shell.SendKeyEvent(input.DPAD, nil, input.KEYCODE_DPAD_DOWN)
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	fmt.Println(result.String())
}

func TestSendText(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		result, err := client.Shell.SendString("Alessandro")
		assert.Nil(t, err)
		assert.True(t, result.IsOk())
		fmt.Println(result.String())
		wg.Done()
	}()
	wg.Wait()
}

func TestSendKeys(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		result, err := client.Shell.SendKeyEvents(input.DPAD, nil, input.KEYCODE_DEL, input.KEYCODE_DEL, input.KEYCODE_DEL, input.KEYCODE_DEL, input.KEYCODE_DEL)
		assert.Nil(t, err)
		assert.True(t, result.IsOk())
		fmt.Println(result.String())

		wg.Done()
	}()

	wg.Wait()
}

func TestGetEvents(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)
	result, err := client.Shell.GetEvents()
	assert.Nil(t, err)
	assert.True(t, len(result) > 0)
	repr.Println(result)
}

func TestIsScreenOn(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)
	device := adbclient.NewDevice(client)
	result, err := device.IsScreenOn()
	assert.Nil(t, err)
	assert.True(t, result)

	fmt.Printf("screen is on = %t\n", result)
}

func TestPowerOffOn(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)
	device := adbclient.NewDevice(client)

	result, err := device.IsScreenOn()
	assert.Nil(t, err)
	assert.True(t, result)

	fmt.Printf("screen is on = %t\n", result)

	if result {
		result, err = device.PowerOff()
	} else {
		result, err = device.PowerOn()
	}
	assert.Nil(t, err)
	assert.True(t, result)
}

func TestInstall(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	options := adbclient.InstallOptions{
		GrantPermissions:  true,
		AllowTestPackages: true,
		AllowDowngrade:    true,
		KeepData:          true,
	}

	result, err := client.Install("/Users/alessandro/Downloads/AndroidTV/Swisscom_2022-11-09/ArcCustomizeSettings.apk", &options)

	assert.Nil(t, err)
	assert.True(t, result.IsOk())
	fmt.Println(result.String())
}

func TestIsInstalled(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	device := adbclient.NewDevice(client)
	result, err := device.PackageManager().IsInstalled("com.swisscom.aot.library.sample", "")
	assert.Nil(t, err)
	assert.True(t, result)
}

func TestUninstall(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)
	device := adbclient.NewDevice(client)
	isInstalled, err := device.PackageManager().IsInstalled("com.swisscom.aot.library.sample", "")
	assert.Nil(t, err)
	assert.True(t, isInstalled)

	result, err := client.Uninstall("com.swisscom.aot.library.sample")
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	fmt.Println(result.String())
}

func TestGrantPermissions(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)
	device := adbclient.NewDevice(client)

	result, err := device.PackageManager().GrantPermission("com.swisscom.aot.library.sample", "android.permission.ACCESS_NETWORK_STATE")
	assert.Nil(t, err)
	assert.True(t, result.IsOk())
}

func TestRevokePermissions(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)
	device := adbclient.NewDevice(client)

	result, err := device.PackageManager().RevokePermission("com.swisscom.aot.library.sample", "android.permission.ACCESS_NETWORK_STATE")
	assert.Nil(t, err)
	assert.True(t, result.IsOk())
}

func TestEnablePackage(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	client.Root()
	device := adbclient.NewDevice(client)
	result, err := device.PackageManager().Enable("com.google.android.tvlauncher")
	assert.Nil(t, err)
	assert.True(t, result.IsOk())
}

func TestDisablePackage(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	client.Root()
	device := adbclient.NewDevice(client)
	result, err := device.PackageManager().Disable("com.google.android.tvlauncher")
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	fmt.Println(result.String())
}

func TestPackageManagerGetPath(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)
	device := adbclient.NewDevice(client)

	result, err := device.PackageManager().Path("com.android.systemui", "")
	assert.Nil(t, err)
	assert.Equal(t, "/system_ext/priv-app/SystemUI/SystemUI.apk", result)
}

func TestPmUninstall(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	device := adbclient.NewDevice(client)
	packageName := "..."

	packages, err := device.PackageManager().ListPackagesWithFilter(packagemanager.PackageOptions{}, packageName)
	assert.Nil(t, err)

	installed := streams.Any(packages, func(p packagemanager.Package) bool {
		logging.Log.Debugf("Checking package %s", p.Name)
		return p.Name == packageName
	})

	if !installed {
		logging.Log.Warn("Not installed. Skipping test")
		return
	}

	options := packagemanager.UninstallOptions{
		KeepData:    true,
		User:        "0",
		VersionCode: "",
	}
	result, err := device.PackageManager().Uninstall(packageName, &options)
	assert.Nil(t, err)
	assert.True(t, result.IsOk())
	fmt.Println(result.String())
}

func TestPmInstall(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	device := adbclient.NewDevice(client)

	result, err := client.Push(local_apk, "/data/local/tmp/")
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	remote_apk := fmt.Sprintf("/data/local/tmp/%s", path.Base(local_apk))
	fmt.Printf("remote apk name: %s\n", remote_apk)

	result, err = device.PackageManager().Install(remote_apk, &packagemanager.InstallOptions{
		User:                "0",
		RestrictPermissions: false,
		Pkg:                 "",
		InstallLocation:     1,
		GrantPermissions:    true,
	})

	assert.Nil(t, err)
	assert.True(t, result.IsOk())
	fmt.Println(result.String())

}

func TestDump(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	pkg := "com.android.tv.settings"

	device := adbclient.NewDevice(client)
	result, err := device.PackageManager().Dump(pkg)
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	parser := packagemanager.NewPackageReader(result.Output())
	assert.NotNil(t, parser)

	if parser == nil {
		fmt.Printf("%s\n", result.String())
		return
	}

	assert.Equal(t, pkg, parser.PackageName())
	assert.Equal(t, "1000", parser.UserID())
	assert.Equal(t, "1.0", parser.VersionName())
	assert.Equal(t, "1", parser.VersionCode())
	assert.True(t, len(parser.CodePath()) > 1)
	assert.True(t, len(parser.TimeStamp()) > 1)
	assert.True(t, len(parser.LastUpdateTime()) > 1)
	assert.True(t, len(parser.FirstInstallTime()) > 1)

	flags := parser.Flags()
	assert.True(t, len(flags) > 0)

	for _, v := range flags {
		assert.True(t, len(v) > 1)
	}

}

func TestRuntimePermissions(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	device := adbclient.NewDevice(client)
	result, err := device.PackageManager().RuntimePermissions("com.netflix.ninja")
	assert.Nil(t, err)
	assert.True(t, len(result) > 0)

	for _, v := range result {
		logging.Log.Debug(v.String())
	}
}

func TestInstallPermissions(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	device := adbclient.NewDevice(client)
	result, err := device.PackageManager().InstallPermissions("com.netflix.ninja")
	assert.Nil(t, err)
	assert.True(t, len(result) > 0)

	for _, v := range result {
		logging.Log.Debug(v.String())
	}
}

func TestRequestedPermissions(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	device := adbclient.NewDevice(client)
	result, err := device.PackageManager().RequestedPermissions("com.netflix.ninja")
	assert.Nil(t, err)
	assert.True(t, len(result) > 0)

	for _, v := range result {
		logging.Log.Debug(v.String())
	}
}

func TestClearPackage(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)
	device := adbclient.NewDevice(client)
	result, err := device.PackageManager().Clear("com.netflix.ninja")
	assert.Nil(t, err)
	assert.True(t, result.IsOk())
}

func TestListSettings(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)
	err := client.Root()
	assert.Nil(t, err)

	settings, err := client.Shell.ListSettings(types.SettingsGlobal)
	assert.Nil(t, err)
	assert.True(t, settings.Len() > 0)
	logging.Log.Debugf("loaded %d global settings", settings.Len())

	settings, err = client.Shell.ListSettings(types.SettingsSystem)
	assert.Nil(t, err)
	assert.True(t, settings.Len() > 0)
	logging.Log.Debugf("loaded %d system settings", settings.Len())

	settings, err = client.Shell.ListSettings(types.SettingsSecure)
	assert.Nil(t, err)
	assert.True(t, settings.Len() > 0)
	logging.Log.Debugf("loaded %d secure settings", settings.Len())
}

func TestGetSettings(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)
	err := client.Root()
	assert.Nil(t, err)

	settings, err := client.Shell.GetSetting("user_rotation", types.SettingsSystem)
	assert.Nil(t, err)
	assert.NotNil(t, settings)
	assert.Equal(t, "0", *settings)
	logging.Log.Debugf("user_rotation = %s", *settings)

	settings, err = client.Shell.GetSetting("transition_animation_scale", types.SettingsGlobal)
	assert.Nil(t, err)
	assert.NotNil(t, settings)
	assert.True(t, len(*settings) > 0)
	logging.Log.Debugf("transition_animation_scale = %s", *settings)
}

func TestPutSettings(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)
	err := client.Root()
	assert.Nil(t, err)

	err = client.Shell.PutSetting("transition_animation_scale", "1.1", types.SettingsGlobal)
	assert.Nil(t, err)

	settings, err := client.Shell.GetSetting("transition_animation_scale", types.SettingsGlobal)
	assert.Nil(t, err)
	assert.NotNil(t, settings)
	assert.Equal(t, "1.1", *settings)

	err = client.Shell.PutSetting("transition_animation_scale", "1.0", types.SettingsGlobal)
	assert.Nil(t, err)

	settings, err = client.Shell.GetSetting("transition_animation_scale", types.SettingsGlobal)
	assert.Nil(t, err)
	assert.NotNil(t, settings)
	assert.Equal(t, "1.0", *settings)
}

func TestDeleteSettings(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)
	err := client.Root()
	assert.Nil(t, err)

	err = client.Shell.PutSetting("a_custom_setting", "1", types.SettingsGlobal)
	assert.Nil(t, err)

	settings, err := client.Shell.GetSetting("a_custom_setting", types.SettingsGlobal)
	assert.Nil(t, err)
	assert.NotNil(t, settings)
	assert.Equal(t, "1", *settings)

	err = client.Shell.DeleteSetting("a_custom_setting", types.SettingsGlobal)
	assert.Nil(t, err)

	settings, err = client.Shell.GetSetting("a_custom_setting", types.SettingsGlobal)
	assert.Nil(t, err)
	assert.Nil(t, settings)
}

func TestScan(t *testing.T) {
	sc := scanner.NewScanner()

	logging.Log.Debugf("Scanning for devices...")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for remoteAddr := range sc.Results {
			if remoteAddr != nil {
				logging.Log.Infof("Device found: %s (name=%s, mac=%s)", remoteAddr.GetSerialAddress(), remoteAddr.Name(), remoteAddr.MacAddress().String())
			}
		}
	}()
	sc.Scan()
	wg.Wait()
	logging.Log.Infof("Done")

	logging.Log.Debugf("Scanning mdns services...")

	client := adbclient.NullClient(true)
	services, _ := client.Mdns.Services()

	for _, service := range services {
		logging.Log.Infof("Mdns found: %s (name=%s)", service.Address().GetSerialAddress(), service.Name())
	}
	logging.Log.Infof("Done")
}

func TestLogcatSimple(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)
	since := time.Now().Add(-3 * time.Hour)

	result, err := client.Logcat(types.LogcatOptions{
		Expr:     "swisscom",
		Dump:     true,
		Filename: "",
		Tags:     nil,
		Format:   "",
		Since:    &since,
		Pids:     nil,
		Timeout:  10 * time.Second,
	})

	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	for _, line := range result.OutputLines(true) {
		logging.Log.Debugf(line)
	}
}

func TestLogcatToDeviceFile(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)
	since := time.Now().Add(-3 * time.Hour)

	result, err := client.Logcat(types.LogcatOptions{
		Expr:     "swisscom",
		Dump:     true,
		Filename: "/sdcard/Download/logcat.txt",
		Tags:     nil,
		Format:   "",
		Since:    &since,
		Pids:     nil,
		Timeout:  10 * time.Second,
	})

	assert.Nil(t, err)
	assert.True(t, result.IsOk())
	fmt.Println(result.String())
}

func TestLogcatToLocalFile(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)
	since := time.Now().Add(-3 * time.Hour)

	localFile, _ := os.Create("logcat.txt")

	result, err := client.Logcat(types.LogcatOptions{
		Expr:    "swisscom",
		Dump:    true,
		File:    localFile,
		Tags:    nil,
		Format:  "",
		Since:   &since,
		Pids:    nil,
		Timeout: 10 * time.Second,
	})

	assert.Nil(t, err)
	assert.True(t, result.IsOk())
	fmt.Println(result.String())
}

func TestLogcatProcessPipe(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	defer close(c)

	since := time.Now().Add(-2 * time.Hour)

	pb, err := client.LogcatPipe(types.LogcatOptions{
		Format: "pid",
		Since:  &since,
		Tags: []types.LogcatTag{
			{
				Name:  "tvlib.RestClient",
				Level: types.LogcatVerbose,
			},
		},
		Timeout: 20 * time.Second,
	})

	if err != nil {
		panic(err)
	}

	pipeOutput := pb.StdoutPipe
	fmt.Printf("pipeOutput: %s\n", pipeOutput)

	err = processbuilder.Start(pb)
	assert.Nil(t, err)

	fmt.Println("Now starting the scanner...")

	scanner := bufio.NewScanner(pipeOutput)
	scanned := 0
	for scanner.Scan() {
		text := scanner.Text()
		fmt.Println(text)

		scanned += 1
		if scanned > 5 {
			// c <- os.Interrupt
			break
		}
	}

	exit, _, err := processbuilder.Wait(pb)
	assert.Nil(t, err)
	assert.Equal(t, 0, exit)

	fmt.Println("ok done")
}

func TestClearLogcat(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	err := client.ClearLogcat()
	assert.Nil(t, err)
}

func TestListDumpsys(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	result, err := client.Shell.ListDumpSys()
	assert.Nil(t, err)
	assert.True(t, len(result) > 0)

	for index, line := range result {
		logging.Log.Debugf("%d = %s", index, line)
	}
}

func TestDumpsys(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	result, err := client.Shell.DumpSys("bluetooth_manager")
	assert.Nil(t, err)
	assert.True(t, result.IsOk())
	assert.True(t, len(result.OutputLines(true)) > 0)

	parser := types.DumpsysParser{Lines: result.OutputLines(true)}
	sections := parser.FindSections()

	for _, line := range sections {
		fmt.Println(line)
	}

	fmt.Println("")
	fmt.Println("")

	section := parser.FindSection("AdapterProperties")
	assert.NotNil(t, section)

	fmt.Println("AdapterProperties:")
	for _, line := range section.Lines {
		fmt.Println(line)
	}

	fmt.Println("")
	fmt.Println("")

	section = parser.FindSection("Bluetooth Status")
	assert.NotNil(t, section)

	fmt.Println("Bluetooth Status:")
	for _, line := range section.Lines {
		fmt.Println(line)
	}
}

func TestFactoryReset(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)
	assert.True(t, client.MustRoot())

	intent := types.NewIntent()
	intent.Action = "android.intent.action.FACTORY_RESET"
	intent.Package = "android"
	intent.ReceiverForeground = true
	intent.Wait = true

	device := adbclient.NewDevice(client)
	result, err := device.ActivityManager().Broadcast(intent)
	assert.Nil(t, err)
	assert.True(t, result.IsOk())
}

func TestScreenMirror(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)
	assert.True(t, client.MustRoot())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	cmd1 := client.Shell.NewCommand().WithArgs("while true; do screenrecord --output-format=h264 -; done").ToCommand()
	cmd2 := processbuilder.NewCommand("ffplay", "-framerate", "60", "-probesize", "32", "-sync", "video", "-")

	o, e, code, _, err := processbuilder.Output(processbuilder.Option{Timeout: 10 * time.Second}, cmd1, cmd2)
	assert.Nil(t, err)
	assert.Equal(t, 0, code)

	fmt.Printf("output %s\n", o.String())
	fmt.Printf("err %s\n", e.String())

	if err != nil {
		println(err.Error())
	}
}

func TestDebugWorkManager(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	var device = adbclient.NewDevice(client)

	// clear logcat
	err := client.ClearLogcat()
	assert.Nil(t, err)

	var intent = types.NewIntent()
	intent.Action = "androidx.work.diagnostics.REQUEST_DIAGNOSTICS"
	intent.Package = "com.swisscom.aot.library.standalone"
	intent.Wait = true

	_, err = device.ActivityManager().Broadcast(intent)
	assert.Nil(t, err)

	since := time.Now().Add(-10 * time.Second)

	result, err := client.Logcat(types.LogcatOptions{
		Format:  "brief",
		Since:   &since,
		Dump:    false,
		Timeout: 2 * time.Second,
		Tags: []types.LogcatTag{
			{
				Name:  "WM-DiagnosticsWrkr",
				Level: types.LogcatVerbose,
			},
		},
	})

	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("exitCode: %d\n", result.ExitCode)

	f := regexp.MustCompile(`[\w]{8}-[\w]{4}-[\w]{4}-[\w]{4}-[\w]{12}\s+(?P<classname>[\w\.]+)\s+(?P<jobid>[\w]+)\s+(?P<status>[\w]+)\s+(?P<name>[\w\.]+)\s+(?P<tags>[\w\.,]+)`)

	for _, line := range result.OutputLines(true) {
		match := f.FindStringSubmatch(line)
		if len(match) > 0 {
			groups := make(map[string]string)
			for i, name := range f.SubexpNames() {
				if i != 0 && name != "" {
					groups[name] = match[i]
				}
			}
			fmt.Printf("[%s] %s => %s\n", groups["jobid"], groups["classname"], groups["status"])
		}
	}
}

func TestMotionEvents(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	client.Shell.MotionEvent(input.MOUSE, input.DOWN, types.Pair[int, int]{First: 500, Second: 500})
	client.Shell.MotionEvent(input.MOUSE, input.MOVE, types.Pair[int, int]{First: 500, Second: 500})
	client.Shell.MotionEvent(input.MOUSE, input.UP, types.Pair[int, int]{First: 300, Second: 700})

	client.Shell.Swipe(input.TOUCHSCREEEN, 2000, types.Pair[int, int]{First: 100, Second: 100}, types.Pair[int, int]{First: 600, Second: 600})

	client.Shell.Tap(input.TOUCHSCREEEN, types.Pair[int, int]{First: 100, Second: 100})
	client.Shell.Tap(input.TOUCHSCREEEN, types.Pair[int, int]{First: 200, Second: 200})
}

func TestInputPress(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	client.Shell.Press(input.TRACKBALL)
	client.Shell.Roll(input.TRACKBALL, types.Pair[int, int]{First: 200, Second: 200})
	client.Shell.Press(input.TRACKBALL)
}

func TestStartService(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	intent := types.NewIntent()
	intent.Action = "swisscom.android.tv.action.FIRMWARE_PASSIVE_CHECK"
	intent.Component = "com.swisscom.aot.library.standalone/.service.SystemService"
	intent.Wait = true

	var device = adbclient.NewDevice(client)
	_, err := device.ActivityManager().StartService(intent)
	assert.Nil(t, err)
}

package adbclient_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/sephiroth74/go_adb_client/scanner"
	"github.com/sephiroth74/go_adb_client/shell"
	"github.com/sephiroth74/go_adb_client/transport"
	streams "github.com/sephiroth74/go_streams"
	"golang.org/x/sys/unix"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"pkg.re/essentialkaos/ek.v12/path"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/alecthomas/repr"
	"github.com/magiconair/properties"
	"github.com/reactivex/rxgo/v2"
	"github.com/stretchr/testify/assert"

	"github.com/sephiroth74/go_adb_client"
	"github.com/sephiroth74/go_adb_client/connection"
	"github.com/sephiroth74/go_adb_client/input"
	"github.com/sephiroth74/go_adb_client/logging"
	"github.com/sephiroth74/go_adb_client/mdns"
	"github.com/sephiroth74/go_adb_client/packagemanager"
	"github.com/sephiroth74/go_adb_client/types"
)

var device_ip1 = net.IPv4(192, 168, 1, 105)
var device_ip2 = net.IPv4(192, 168, 1, 3)
var device_ip = device_ip2

var local_apk = ""

func init() {
}

func NewClient() *adbclient.Client {
	return adbclient.NewClient(types.ClientAddr{IP: device_ip, Port: 5555}, true)
}

func AssertClientConnected(t *testing.T, client *adbclient.Client) {
	result, err := client.Connect(5 * time.Second)
	assert.Nil(t, err, "Error connecting to %s", client.Address.String())
	assert.True(t, result.IsOk(), "Error: %s", result.Error())
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

	result, err := client.Shell.NewProcess().WithArgs("nproc").WithStdout(&w).Invoke()
	assert.Nil(t, err)
	assert.True(t, result.IsOk())
	assert.True(t, result.Output() == "")
	assert.True(t, result.Error() == "")
	assert.True(t, s.Len() > 0)
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

func TestDirWalk(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	var device = adbclient.NewDevice(client)
	root, err := device.Client.Root()
	assert.Nil(t, err)
	assert.True(t, root.IsOk())

	dirname := "/sdcard/"
	result, err := device.Client.Shell.ListDir(dirname)
	assert.Nil(t, err)

	logging.Log.Info().Msgf("result of listdir: %s", dirname)
	for _, line := range result {
		logging.Log.Info().Msg(line.String())
	}

}

func TestRecordScreen(t *testing.T) {
	var client = NewClient()
	defer client.Disconnect()
	AssertClientConnected(t, client)

	var device = adbclient.NewDevice(client)

	var result transport.Result
	var err error

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	defer close(c)

	result, err = device.Client.Shell.ScreenRecord(shell.ScreenRecordOptions{
		Bitrate:   8000000,
		Timelimit: 5,
		Rotate:    false,
		BugReport: false,
		Size:      &types.Size{Width: 1920, Height: 1080},
		Verbose:   false,
	}, c, "/sdcard/Download/screenrecord.mp4")

	if err != nil {
		assert.Equal(t, int(unix.SIGINT), result.ExitCode)
	}

	if result.ExitCode != int(unix.SIGINT) {
		assert.True(t, result.IsOk())
		println(result.Output())
		println(result.Error())
	}
}

func TestConnect(t *testing.T) {
	var client = NewClient()

	result, err := client.Connect(5 * time.Second)
	logging.Log.Debug().Msgf("result=%s", result.ToString())
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

	result, err = client.WaitForDeviceWithTimeout(connection.WaitForDeviceTimeout)
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
	client := NewClient()
	AssertClientConnected(t, client)

	result, err := client.Root()
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
	var client = NewClient()
	AssertClientConnected(t, client)
	list, err := client.ListDevices()

	if err == nil {
		for x := 0; x < len(list); x++ {
			logging.Log.Debug().Msgf("Device: %#v\n", list[x])
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

	result, err = client.WaitForDeviceWithTimeout(time.Duration(2) * time.Minute)
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

	result, err := client.Root()
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
		logging.Log.Warn().Msgf("error=%#v\n", err.Error())
	}
	assert.Nil(t, err)
	assert.True(t, result.IsOk())
}

func TestGetVersion(t *testing.T) {
	conn := connection.NewConnection(true)
	result, err := conn.Version()
	assert.Nil(t, err)
	assert.NotEmpty(t, result)
	logging.Log.Debug().Msgf("adb version=%s", result)
}

func TestMdns(t *testing.T) {
	var mdns = mdns.NewMdns(connection.NewConnection(true))
	result, err := mdns.Check()
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	devices, err := mdns.Services()
	assert.Nil(t, err)

	logging.Log.Debug().Msgf("Found %d devices", len(devices))

	for i := 0; i < len(devices); i++ {
		logging.Log.Debug().Msgf("device: %#v", devices[i])
	}

	assert.True(t, len(devices) > 0)

	client2 := adbclient.NewClient(devices[1], true)
	result, err = client2.Connect(5 * time.Second)
	assert.Nil(t, err)
	logging.Log.Debug().Msg(result.String())

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

	result, err := client.Pull("/data/data/com.library.sample", path)

	logging.Log.Debug().Msgf("output: %s", result.Output())
	logging.Log.Debug().Msgf("error: %s", result.Error())

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

	result, err := client.Shell.Which("which")
	assert.Nil(t, err)

	println(result.Repr())
}

func TestRx(t *testing.T) {
	var client = NewClient()

	observable := rxgo.FromEventSource(client.Channel)

	observable.DoOnNext(func(i interface{}) {
		logging.Log.Info().Msgf("onNext: %s", repr.String(i))
	})

	observable.DoOnCompleted(func() {
		logging.Log.Info().Msg("onComplete")
	})

	observable.DoOnError(func(err error) {
		logging.Log.Info().Msgf("onError: %s", err.Error())
	})

	client.Disconnect()
	client.Connect(5 * time.Second)
	client.IsConnected()
	// client.Disconnect()

	client = nil
	runtime.GC()

	time.Sleep(time.Duration(2) * time.Second)
}

func TestActivityManager(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	var device = adbclient.NewDevice(client)

	var intent = types.NewIntent()
	intent.Action = "android.action.View"
	intent.Extra.Es["key1"] = "string1"
	intent.Extra.Es["key2"] = "string2"
	intent.Extra.Eia["key_eia1"] = []int{1, 2, 3}

	device.ActivityManager().Broadcast(intent)
}

func TestShellCat(t *testing.T) {
	var client = NewClient()
	AssertClientConnected(t, client)

	assert.True(t, client.MustRoot())

	result, err := client.Shell.Whoami()
	assert.Nil(t, err)
	assert.Equal(t, "root", result.Output())

	result, err = client.Shell.Cat("/system/build.prop")
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	props, err := properties.Load(result.Stdout, properties.UTF8)
	assert.Nil(t, err)

	for _, k := range props.Keys() {
		v, ok := props.Get(k)

		if ok {
			logging.Log.Debug().Msgf("%s = %s", k, v)
		} else {
			logging.Log.Warn().Msgf("Error reading key %s", k)
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

	logging.Log.Debug().Msgf("ro.build.product -> %s\n", *prop)

	prop = shell.GetProp("invalid.key.string")
	assert.Nil(t, prop)
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
		logging.Log.Debug().Msgf("%s=%s", v, value)
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
			logging.Log.Debug().Msgf("%s=%s", v, *pt)
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

	logging.Log.Info().Msgf("device name: %s", *deviceName)
	logging.Log.Info().Msgf("device api level: %s", *apiLevel)
	logging.Log.Info().Msgf("device version release: %s", *version)
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
}

func TestWriteScreenCap(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	target_file, err := filepath.Abs("./exports/screencap.png")
	assert.Nil(t, err)

	var target_dir = filepath.Dir(target_file)

	logging.Log.Info().Msgf("target file: %s", target_file)
	logging.Log.Info().Msgf("target dir: %s", target_dir)

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
		logging.Log.Error().Msgf(err.Error())
		logging.Log.Error().Msgf(result.Error())
	}

	assert.True(t, result.IsOk())
	if _, err := os.Stat(target_file); errors.Is(err, os.ErrNotExist) {
		assert.Fail(t, "Screencap not exported")
	}

	os.RemoveAll(target_dir)
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
		logging.Log.Error().Msgf(err.Error())
		logging.Log.Error().Msgf(result.Error())
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
		logging.Log.Debug().Msgf("%s, uid:%s", p.Name, p.UID)
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
		logging.Log.Debug().Msgf("%s, uid:%s", p.Name, p.UID)
	}
}

func TestIsSystem(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	//defer client.Disconnect()

	device := adbclient.NewDevice(client)
	pm := device.PackageManager()

	is_system, _ := pm.IsSystem("com.google.youtube")
	assert.True(t, is_system)

	is_system, _ = pm.IsSystem("com.google.youtube")
	assert.False(t, is_system)
}

func TestSendKey(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)
	result, err := client.Shell.SendKeyEvent(input.KEYCODE_DPAD_DOWN)
	assert.Nil(t, err)
	assert.True(t, result.IsOk())
}

func TestSendText(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	// defer client.Disconnect()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		client.Shell.SendString("Alessandro")
		wg.Done()
	}()

	wg.Wait()
}

func TestSendKeys(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	// defer client.Disconnect()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		client.Shell.SendKeyEvents(input.KEYCODE_DEL, input.KEYCODE_DEL, input.KEYCODE_DEL, input.KEYCODE_DEL, input.KEYCODE_DEL)
		wg.Done()
	}()

	wg.Wait()
}

func TestGetEvents(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	// defer client.Disconnect()

	result, err := client.Shell.GetEvents()
	assert.Nil(t, err)

	repr.Println(result)
}

func TestIsScreenOn(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	// defer client.Disconnect()

	device := adbclient.NewDevice(client)

	result, err := device.IsScreenOn()
	assert.Nil(t, err)
	assert.True(t, result)

	if result {
		result, err = device.PowerOff()
	} else {
		result, err = device.PowerOn()
	}
	assert.Nil(t, err)
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

	result, err := client.Install(local_apk, &options)

	assert.Nil(t, err)
	assert.True(t, result.IsOk())
}

func TestPmUninstall(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	device := adbclient.NewDevice(client)
	packageName := "..."

	packages, err := device.PackageManager().ListPackagesWithFilter(packagemanager.PackageOptions{}, packageName)
	assert.Nil(t, err)

	installed := streams.Any(packages, func(p packagemanager.Package) bool {
		logging.Log.Debug().Msgf("Checking package %s", p.Name)
		return p.Name == packageName
	})

	if !installed {
		logging.Log.Warn().Msgf("Not installed. Skipping test")
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
}

func TestPmInstall(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	device := adbclient.NewDevice(client)

	result, err := client.Push(local_apk, "/data/local/tmp/")
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	remote_apk := fmt.Sprintf("/data/local/tmp/%s", path.Base(local_apk))
	result, err = device.PackageManager().Install(remote_apk, &packagemanager.InstallOptions{
		User:                "0",
		RestrictPermissions: false,
		Pkg:                 "",
		InstallLocation:     1,
		GrantPermissions:    true,
	})

	assert.Nil(t, err)
	assert.True(t, result.IsOk())

}

func TestDump(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	pkg := "com.netflix.ninja"

	device := adbclient.NewDevice(client)
	result, err := device.PackageManager().Dump(pkg)
	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	parser := packagemanager.NewPackageReader(result.Output())
	assert.NotNil(t, parser)
	assert.Equal(t, pkg, parser.PackageName())
	assert.Equal(t, "1000", parser.UserID())
	assert.Equal(t, "1.0.41", parser.VersionName())
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
		logging.Log.Debug().Msg(v.String())
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
		logging.Log.Debug().Msg(v.String())
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
		logging.Log.Debug().Msg(v.String())
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
	_, err := client.Root()
	assert.Nil(t, err)

	settings, err := client.Shell.ListSettings(types.SettingsGlobal)
	assert.Nil(t, err)
	assert.True(t, settings.Len() > 0)
	logging.Log.Debug().Msgf("loaded %d global settings", settings.Len())

	settings, err = client.Shell.ListSettings(types.SettingsSystem)
	assert.Nil(t, err)
	assert.True(t, settings.Len() > 0)
	logging.Log.Debug().Msgf("loaded %d system settings", settings.Len())

	settings, err = client.Shell.ListSettings(types.SettingsSecure)
	assert.Nil(t, err)
	assert.True(t, settings.Len() > 0)
	logging.Log.Debug().Msgf("loaded %d secure settings", settings.Len())
}

func TestGetSettings(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)
	_, err := client.Root()
	assert.Nil(t, err)

	settings, err := client.Shell.GetSetting("system_locales", types.SettingsSystem)
	assert.Nil(t, err)
	assert.NotNil(t, settings)
	assert.True(t, len(*settings) > 0)
	logging.Log.Debug().Msgf("system_locales = %s", *settings)

	settings, err = client.Shell.GetSetting("transition_animation_scale", types.SettingsGlobal)
	assert.Nil(t, err)
	assert.NotNil(t, settings)
	assert.True(t, len(*settings) > 0)
	logging.Log.Debug().Msgf("transition_animation_scale = %s", *settings)
}

func TestPutSettings(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)
	_, err := client.Root()
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
	_, err := client.Root()
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

	logging.Log.Debug().Msgf("Scanning for devices...")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for remoteAddr := range sc.Results {
			if remoteAddr != nil {
				logging.Log.Info().Msgf("Device found: %s", *remoteAddr)
			}
		}
	}()
	sc.Scan()
	wg.Wait()
	logging.Log.Info().Msgf("Done")

	logging.Log.Debug().Msgf("Scanning mdns services...")

	client := adbclient.NullClient(true)
	services, _ := client.Mdns.Services()

	for _, service := range services {
		logging.Log.Info().Msgf("Mdns found: %s", service.Address.GetSerialAddress())
	}
	logging.Log.Info().Msgf("Done")
}

func TestLogcat(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	since := time.Now().Add(-3 * time.Hour)

	result, err := client.Logcat(types.LogcatOptions{
		Expr:     "Authorization: Bearer ([0-9a-zA-Z-]+)",
		Dump:     true,
		Filename: "",
		Tags:     nil,
		Format:   "",
		Since:    &since,
		Pids:     nil,
		Timeout:  0 * time.Second,
	})

	//
	//result, err := client.Logcat(adbclient.LogcatOptions{
	//	Expr:     "",
	//	Dump:     true,
	//	Filename: "",
	//	Tags: []adbclient.LogcatTag{
	//		{
	//			Name:  "PropertiesReceiver",
	//			Level: adbclient.LogcatDebug,
	//		},
	//		{
	//			Name:  "MY_CUSTOM_TAG",
	//			Level: adbclient.LogcatVerbose,
	//		},
	//	},
	//	Format: "tag",
	//	Since:  "",
	//	Pids:   nil,
	//})

	assert.Nil(t, err)
	assert.True(t, result.IsOk())

	if err != nil {
		logging.Log.Warn().Msgf("Error: %s", err.Error())
		logging.Log.Warn().Msgf(result.Error())
	}

	for _, line := range result.OutputLines() {
		logging.Log.Debug().Msgf(line)
	}
}

package adbclient_test

import (
	"errors"
	"fmt"
	"it.sephiroth/adbclient/util"
	"net"
	"os"
	"path/filepath"
	"pkg.re/essentialkaos/ek.v12/path"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alecthomas/repr"
	"github.com/magiconair/properties"
	goLogging "github.com/op/go-logging"
	"github.com/reactivex/rxgo/v2"
	"github.com/stretchr/testify/assert"

	"it.sephiroth/adbclient"
	"it.sephiroth/adbclient/connection"
	"it.sephiroth/adbclient/input"
	"it.sephiroth/adbclient/logging"
	"it.sephiroth/adbclient/mdns"
	"it.sephiroth/adbclient/packagemanager"
	"it.sephiroth/adbclient/types"
)

var device_ip1 = net.IPv4(192, 168, 1, 105)
var device_ip2 = net.IPv4(192, 168, 1, 110)
var device_ip = device_ip2

var local_apk = ""

var log = logging.GetLogger("test")

func init() {
	logging.SetLevel(goLogging.DEBUG)
}

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

	result, err := client.Pull("/data/data/com.library.sample", path)

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

	result, err := client.Shell.Which("which")
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

	assert.True(t, client.TryRoot())

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

	shell := client.Shell
	prop := shell.GetProp("wlan.driver.status")
	assert.NotNil(t, prop)
	assert.Equal(t, "ok", *prop)

	prop = shell.GetProp("invalid.key.string")
	assert.Nil(t, prop)
}

func TestShellGetProps(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	shell := client.Shell
	props, err := shell.GetProps()
	assert.Nil(t, err)
	assert.True(t, len(props) > 0)

	for _, v := range props {
		log.Debugf("%s=%s", v.First, v.Second)
	}
}

func TestShellGetPropsType(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	shell := client.Shell
	props, err := shell.GetProps()
	assert.Nil(t, err)
	assert.True(t, len(props) > 0)

	for _, v := range props {
		pt, ok := shell.GetPropType(v.First)
		assert.Truef(t, ok, "Error getting type of key %s", v.First)
		if ok {
			log.Debugf("%s=%s", v.First, *pt)
		}
	}
}

func TestDevice(t *testing.T) {
	client := NewClient()
	AssertClientConnected(t, client)

	device := adbclient.Device[types.ClientAddr]{Client: client}
	deviceName := device.Name()
	apiLevel := device.ApiLevel()
	version := device.Version()

	assert.NotNil(t, deviceName)
	assert.NotNil(t, apiLevel)
	assert.NotNil(t, version)

	log.Infof("device name: %s", *deviceName)
	log.Infof("device api level: %s", *apiLevel)
	log.Infof("device version release: %s", *version)
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

	var target_file = "./exports/screencap.png"
	var target_dir = filepath.Dir(target_file)

	log.Infof("target file: %s", target_file)
	log.Infof("target dir: %s", target_dir)

	os.RemoveAll(target_dir)
	os.MkdirAll(target_dir, 0755)

	_, err := os.Stat(target_dir)
	assert.Nil(t, err)

	f, err := os.Create(target_file)
	assert.Nil(t, err)
	os.Chmod(target_file, 0755)

	log.Infof("f: %v", f)

	device := adbclient.NewDevice(client)
	result, err := device.WriteScreenCap(f)
	assert.Nil(t, err)

	if err != nil {
		log.Error(err.Error())
		log.Error(result.Error())
	}

	assert.True(t, result.IsOk())
	if _, err := os.Stat("./exports/screencap.png"); errors.Is(err, os.ErrNotExist) {
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
		log.Error(err.Error())
		log.Error(result.Error())
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
	packages, err := pm.ListPackages(&packagemanager.PackageOptions{
		ShowOnlyEnabed: true,
		ShowOnlySystem: true,
	})
	assert.Nil(t, err)

	for _, p := range packages {
		log.Debugf("%s, uid:%s", p.Name, p.UID)
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

	packages, err := pm.ListPackagesWithFilter(&packagemanager.PackageOptions{ShowOnlySystem: true}, "com.google")
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(packages), 1)

	for _, p := range packages {
		assert.True(t, p.Filename != "")
		assert.True(t, strings.HasPrefix(p.Name, "com.google"))
		assert.True(t, p.VersionCode != "")
		assert.True(t, p.UID != "")
		assert.True(t, p.MaybeIsSystem())
		log.Debugf("%s, uid:%s", p.Name, p.UID)
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

	packages, err := device.PackageManager().ListPackagesWithFilter(nil, packageName)
	assert.Nil(t, err)

	installed := util.Any(packages, func(p packagemanager.Package) bool {
		log.Debugf("Checking package %s", p.Name)
		return p.Name == packageName
	})

	if !installed {
		log.Warning("Not installed. Skipping test")
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
		log.Debug(v)
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
		log.Debug(v)
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
		log.Debug(v)
	}
}

func TestScan(t *testing.T) {
	// client := NewClient()
	// AssertClientConnected(t, client)

	// device := adbclient.NewDevice(client)
	// pm := device.PackageManager()

	// pm.List(packagemanager.PACKAGES)

	// conn, err := net.DialTimeout("tcp", "192.168.1.122:5555", time.Duration(1)*time.Second)
	// if err != nil {
	// 	log.Warningf("Failed to connect to host")
	// 	return
	// }

	// defer conn.Close()

	// log.Debugf("addr: %v", conn.RemoteAddr().String())

	scanner := NewScanner()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for remoteAddr := range scanner.Results {
			if remoteAddr != nil {
				log.Infof("Device found: %s", *remoteAddr)
			}
		}
	}()

	scanner.Scan()
	wg.Wait()

	log.Info("Done")
}

type Scanner struct {
	Results chan *string
}

func NewScanner() *Scanner {
	return &Scanner{
		Results: make(chan *string),
	}
}

func (s *Scanner) Scan() {
	go func() {
		wg := new(sync.WaitGroup)
		baseHost := "192.168.1.%d:5555"
		// Adding routines to workgroup and running then
		for i := 1; i < 255; i++ {
			host := fmt.Sprintf(baseHost, i)
			wg.Add(1)
			go worker(i, host, s.Results, wg)
		}
		wg.Wait()
		close(s.Results)
	}()
}

func worker(index int, host string, ch chan *string, wg *sync.WaitGroup) {
	// Decreasing internal counter for wait-group as soon as goroutine finishes
	defer wg.Done()
	log.Debugf("[%d] Trying to connect to %s", index, host)
	conn, err := net.DialTimeout("tcp", host, time.Duration(1)*time.Second)
	if err != nil {
		ch <- nil
		return
	}

	defer conn.Close()

	var remoteAddr = conn.RemoteAddr().String()
	ch <- &remoteAddr
}

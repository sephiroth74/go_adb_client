package packagemanager

import (
	"fmt"
	"it.sephiroth/adbclient/transport"
	"regexp"
	"strconv"
	"strings"

	"github.com/alecthomas/repr"
	"it.sephiroth/adbclient/logging"
	"it.sephiroth/adbclient/shell"
	"it.sephiroth/adbclient/types"
)

var log = logging.GetLogger("pm")

type PackageManager[T types.Serial] struct {
	Shell *shell.Shell[T]
}

func (p PackageManager[T]) Path(packagename string) (string, error) {
	result, err := p.Shell.Execute("pm", 0, "path", packagename)
	if err != nil {
		return "", err
	}

	if result.IsOk() {
		return result.Output(), nil
	} else {
		return "", result.NewError()
	}
}

func (p PackageManager[T]) ListPackages(options *PackageOptions) ([]Package, error) {
	return p.ListPackagesWithFilter(options, "")
}

func (p PackageManager[T]) IsSystem(name string) (bool, error) {
	result, err := p.Shell.Execute(
		fmt.Sprintf("pm dump %s | egrep '^ {1,}flags=' | egrep ' {1,}SYSTEM {1,}'", name), 0)

	if err != nil {
		return false, err
	}

	return result.IsOk(), nil
}

func (p PackageManager[T]) Dump(name string) (transport.Result, error) {
	return p.Shell.Executef("pm dump %s", 0, name)
}

// ListPackagesWithFilter List packages on the device
func (p PackageManager[T]) ListPackagesWithFilter(options *PackageOptions, filter string) ([]Package, error) {
	//	list packages [-f] [-d] [-e] [-s] [-3] [-i] [-l] [-u] [-U]
	//	[--show-versioncode] [--apex-only] [--uid UID] [--user USER_ID] [FILTER]
	//  Prints all packages; optionally only those whose name contains
	//  the text in FILTER.  Options are:
	//	-f: see their associated file
	//	-a: all known packages (but excluding APEXes)
	//	-d: filter to only show disabled packages
	//	-e: filter to only show enabled packages
	//	-s: filter to only show system packages
	//	-3: filter to only show third party packages
	//	-i: see the installer for the packages
	//	-l: ignored (used for compatibility with older releases)
	//	-U: also show the package UID
	//	-u: also include uninstalled packages
	//	--show-versioncode: also show the version code
	//	--apex-only: only show APEX packages
	//	--uid UID: filter to only show packages with the given UID
	//	--user USER_ID: only list packages belonging to the given user
	args := []string{"list", "packages", "-f", "-U", "--show-versioncode"}

	if options != nil {
		if options.ShowOnly3rdParty {
			args = append(args, "-3")
		}
		if options.ShowOnlyDisabled {
			args = append(args, "-d")
		}
		if options.ShowOnlyEnabed {
			args = append(args, "-e")
		}
		if options.ShowOnlySystem {
			args = append(args, "-s")
		}
	}

	if filter != "" {
		args = append(args, filter)
	}

	result, err := p.Shell.Execute("pm", 0, args...)
	if err != nil {
		return nil, err
	}

	if result.IsOk() {
		f := regexp.MustCompile(`package:(.*\.apk)=([^\s]+)\s*(versionCode|uid):([^\s]+)\s+(versionCode|uid):([^\s]+)$`)
		var packages = []Package{}

		for _, line := range result.OutputLines() {
			m := f.FindStringSubmatch(line)
			if len(m) == 7 {
				var versionCode string
				var uid string

				if m[3] == "uid" {
					versionCode = m[6]
					uid = m[4]
				} else {
					uid = m[6]
					versionCode = m[4]
				}

				packages = append(packages, Package{
					Filename:    m[1],
					Name:        m[2],
					VersionCode: versionCode,
					UID:         uid,
				})
			}
		}

		return packages, nil
	} else {
		return nil, result.NewError()
	}
}

func (p PackageManager[T]) Install(src string, options *InstallOptions) (transport.Result, error) {
	var args []string
	if options != nil {
		if options.RestrictPermissions {
			args = append(args, "--restrict-permissions")
		}
		if options.User != "" {
			args = append(args, "--user", options.User)
		}
		if options.Pkg != "" {
			args = append(args, "--pkg", options.Pkg)
		}
		if options.InstallLocation > 0 {
			args = append(args, "--install-location", fmt.Sprintf("%d", options.InstallLocation))
		}
		if options.GrantPermissions {
			args = append(args, "-g")
		}
	}
	args = append(args, src)
	return p.Shell.Execute("cmd package install", 0, args...)
}

// Uninstall
// uninstall [-k] [--user USER_ID] [--versionCode VERSION_CODE]
// PACKAGE [SPLIT...]
// Remove the given package name from the system.  May remove an entire app
// if no SPLIT names specified, otherwise will remove only the splits of the
// given app.  Options are:
// -k: keep the data and cache directories around after package removal.
// --user: remove the app from the given user.
// --versionCode: only uninstall if the app has the given version code.
func (p PackageManager[T]) Uninstall(packageName string, options *UninstallOptions) (transport.Result, error) {
	var args []string
	if options != nil {
		if options.KeepData {
			args = append(args, "-k")
		}
		if options.User != "" {
			args = append(args, "--user", options.User)
		}
		if options.VersionCode != "" {
			args = append(args, "--versionCode", options.VersionCode)
		}
	}
	args = append(args, packageName)
	return p.Shell.Execute("cmd package uninstall", 0, args...)
}

func (p PackageManager[T]) RuntimePermissions(packageName string) ([]types.PackagePermission, error) {
	result, err := p.Dump(packageName)
	if err != nil {
		return nil, err
	}

	var runtimePermissions []types.PackagePermission

	data := result.Output()
	f := regexp.MustCompile(`(?m)^\s{3,}runtime permissions:\s+`)
	m := f.FindStringIndex(data)

	if m != nil {
		data = data[m[1]:]
		m := regexp.MustCompile("(?m)^$").FindStringIndex(data)
		if m != nil {
			data = data[:m[1]]

			f = regexp.MustCompile(`(?m)^\s*([^:]+):\s+granted=(false|true),\s+flags=\[\s*([^\]]+)\]$`)
			match := f.FindAllStringSubmatch(data, -1)
			if match != nil {
				for _, v := range match {
					name := v[1]
					granted := v[2]
					flags := strings.Split(strings.TrimSpace(v[3]), "|")
					boolValue, _ := strconv.ParseBool(granted)

					runtimePermissions = append(runtimePermissions, types.PackagePermission{
						Permission: types.Permission{Name: name},
						Granted:    boolValue,
						Flags:      flags,
					})
				}
			}
		}
	}
	return runtimePermissions, nil
}

func (p PackageManager[T]) InstallPermissions(packageName string) ([]types.PackagePermission, error) {
	result, err := p.Dump(packageName)
	if err != nil {
		return nil, err
	}

	if !result.IsOk() {
		return nil, result.NewError()
	}

	var parser = NewPackageReader(result.Output())
	return parser.InstallPermissions(), nil
}

func (p PackageManager[T]) RequestedPermissions(packageName string) ([]types.RequestedPermission, error) {
	result, err := p.Dump(packageName)
	if err != nil {
		return nil, err
	}

	if !result.IsOk() {
		return nil, result.NewError()
	}

	var parser = NewPackageReader(result.Output())
	return parser.RequestedPermissions(), nil
}

// Clear executes a "pm clear packageName" on the connected device
func (p PackageManager[T]) Clear(packageName string) (transport.Result, error) {
	return p.Shell.Executef("pm clear", 0, packageName)
}

func (p PackageManager[T]) GrantPermission(packageName string, permission string) (transport.Result, error) {
	return p.Shell.Executef("pm grant %s %s", 0, packageName, permission)
}

func (p PackageManager[T]) RevokePermission(packageName string, permission string) (transport.Result, error) {
	return p.Shell.Executef("pm revoke %s %s", 0, packageName, permission)
}

type UninstallOptions struct {
	// -k
	KeepData bool
	// --user
	User string
	// --versionCode
	VersionCode string
}

type InstallOptions struct {
	// --user: install under the given user.
	User string
	// --dont-kill: installing a new feature split, don't kill running app
	DontKill bool
	// --restrict-permissions: don't whitelist restricted permissions at install
	RestrictPermissions bool
	// --pkg: specify expected package name of app being installed
	Pkg string
	// --install-location: force the install location:
	// 0=auto, 1=internal only, 2=prefer external
	InstallLocation int
	// -g: grant all runtime permissions
	GrantPermissions bool
}

type PackageOptions struct {
	// -d: filter to only show disabled packages
	ShowOnlyDisabled bool
	// -e: filter to only show enabled packages
	ShowOnlyEnabed bool
	// -s: filter to only show system packages
	ShowOnlySystem bool
	// -3: filter to only show third party packages
	ShowOnly3rdParty bool
}

type Package struct {
	Filename    string
	Name        string
	VersionCode string
	UID         string
}

func (p Package) String() string {
	return repr.String(p)
}

// MaybeIsSystem Check if the package is a system app.
// Sometimes the apk file path or the uid indicate if the package is a system app.
// if false is returned, the package can still be a system app, but it should be checked
// against the dumpsys flags
func (p Package) MaybeIsSystem() bool {
	return strings.HasPrefix(p.Filename, "/system/") || strings.HasPrefix(p.Filename, "/product/") || strings.HasPrefix(p.Filename, "/system_ext/") ||
		strings.HasPrefix(p.Filename, "/vendor/") || strings.HasPrefix(p.Filename, "/apex/") ||
		p.UID == "1000"
}

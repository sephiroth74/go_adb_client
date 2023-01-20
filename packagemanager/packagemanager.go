package packagemanager

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/sephiroth74/go_adb_client/process"

	"github.com/alecthomas/repr"
	"github.com/sephiroth74/go_adb_client/shell"
	"github.com/sephiroth74/go_adb_client/types"
)

type PackageManager struct {
	Shell *shell.Shell
}

func (p PackageManager) Path(packageName string, user string) (string, error) {
	cmd := p.Shell.NewCommand().WithArgs("pm", "path")
	if user != "" {
		cmd.AddArgs("--user", user)
	}

	cmd.AddArgs(packageName)

	result, err := process.SimpleOutput(cmd, p.Shell.Conn.Verbose)

	if err != nil {
		return "", err
	}

	if result.IsOk() {
		f := regexp.MustCompile(`package:(.*)`)
		m := f.FindStringSubmatch(result.Output())
		if m != nil && len(m) == 2 {
			return m[1], nil
		}
		return "", errors.New("path not found")
	} else {
		return "", result.NewError()
	}
}

func (p PackageManager) ListPackages(options PackageOptions) ([]Package, error) {
	return p.ListPackagesWithFilter(options, "")
}

func (p PackageManager) IsSystem(name string) (bool, error) {
	cmd := p.Shell.NewCommand().WithArgs(fmt.Sprintf("pm dump %s | egrep '^ {1,}flags=' | egrep ' {1,}SYSTEM {1,}'", name))
	result, err := process.SimpleOutput(cmd, p.Shell.Conn.Verbose)

	if err != nil {
		return false, err
	}

	return result.IsOk(), nil
}

func (p PackageManager) Dump(name string) (process.OutputResult, error) {
	cmd := p.Shell.NewCommand().WithArgs(fmt.Sprintf("pm dump %s", name))
	return process.SimpleOutput(cmd, p.Shell.Conn.Verbose)
}

func (p PackageManager) IsInstalled(packagename string, user string) (bool, error) {
	pkg, err := p.Path(packagename, user)
	if err != nil {
		return false, err
	}
	return len(pkg) > 0, nil
}

// ListPackagesWithFilter List packages on the device
func (p PackageManager) ListPackagesWithFilter(options PackageOptions, filter string) ([]Package, error) {
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

	if filter != "" {
		args = append(args, filter)
	}

	cmd := p.Shell.NewCommand().WithArgs("pm").AddArgs(args...)
	result, err := process.SimpleOutput(cmd, p.Shell.Conn.Verbose)
	// result, err := p.Shell.ExecuteWithTimeout("pm", 0, args...)
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

func (p PackageManager) Install(src string, options *InstallOptions) (process.OutputResult, error) {
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
	cmd := p.Shell.NewCommand().WithArgs("cmd package install").AddArgs(args...)
	return process.SimpleOutput(cmd, p.Shell.Conn.Verbose)
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
func (p PackageManager) Uninstall(packageName string, options *UninstallOptions) (process.OutputResult, error) {
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
	cmd := p.Shell.NewCommand().WithArgs("cmd package uninstall").AddArgs(args...)
	return process.SimpleOutput(cmd, p.Shell.Conn.Verbose)
	// return p.Shell.ExecuteWithTimeout("cmd package uninstall", 0, args...)
}

func (p PackageManager) RuntimePermissions(packageName string) ([]types.PackagePermission, error) {
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

func (p PackageManager) InstallPermissions(packageName string) ([]types.PackagePermission, error) {
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

func (p PackageManager) RequestedPermissions(packageName string) ([]types.RequestedPermission, error) {
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

func (p PackageManager) DumpPackage(packageName string) (*SimplePackageReader, error) {
	result, err := p.Dump(packageName)
	if err != nil {
		return nil, err
	}

	if !result.IsOk() {
		return nil, result.NewError()
	}

	var parser = NewPackageReader(result.Output())
	return parser, nil
}

// Clear executes a "pm clear packageName" on the connected device
func (p PackageManager) Clear(packageName string) (process.OutputResult, error) {
	cmd := p.Shell.NewCommand().WithArgs(fmt.Sprintf("pm clear %s", packageName))
	return process.SimpleOutput(cmd, p.Shell.Conn.Verbose)
}

func (p PackageManager) GrantPermission(packageName string, permission string) (process.OutputResult, error) {
	return process.SimpleOutput(p.Shell.NewCommand().WithArgs(fmt.Sprintf("pm grant %s %s", packageName, permission)), p.Shell.Conn.Verbose)
}

func (p PackageManager) RevokePermission(packageName string, permission string) (process.OutputResult, error) {
	return process.SimpleOutput(p.Shell.NewCommand().WithArgs(fmt.Sprintf("pm revoke %s %s", packageName, permission)), p.Shell.Conn.Verbose)
}

// Enable enable a package
func (p PackageManager) Enable(packageName string) (process.OutputResult, error) {
	return process.SimpleOutput(p.Shell.NewCommand().WithArgs(fmt.Sprintf("pm enable %s", packageName)), p.Shell.Conn.Verbose)
}

// Disable disable a package
func (p PackageManager) Disable(packageName string) (process.OutputResult, error) {
	return process.SimpleOutput(p.Shell.NewCommand().WithArgs(fmt.Sprintf("pm disable %s", packageName)), p.Shell.Conn.Verbose)
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

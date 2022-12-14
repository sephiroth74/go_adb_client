package packagemanager

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/alecthomas/repr"
	"it.sephiroth/adbclient/logging"
	"it.sephiroth/adbclient/shell"
	"it.sephiroth/adbclient/types"
)

const (
// 	Prints all known permissions, optionally only those in group. Options:
// -g: Organize by group.
// -f: Print all information.
// -s: Short summary.
// -d: Only list dangerous permissions.
// -u: List only the permissions users will see.

// Print the path to the APK of the given package.
// PATH ListType = "path"
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
		fmt.Sprintf("dumpsys package %s | egrep '^ {1,}flags=' | egrep ' {1,}SYSTEM {1,}'", name), 0)

	if err != nil {
		return false, err
	}
	
	return result.IsOk(), nil
}

// List packages on the device
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

// Check if the package is a system app.
// Sometimes the apk file path or the uid indicate if the package is a system app.
// if false is returned, the package can still be a system app, but it should be checked
// against the dumpsys flags
func (p Package) MaybeIsSystem() bool {
	return strings.HasPrefix(p.Filename, "/system/") || strings.HasPrefix(p.Filename, "/product/") || strings.HasPrefix(p.Filename, "/system_ext/") ||
		strings.HasPrefix(p.Filename, "/vendor/") || strings.HasPrefix(p.Filename, "/apex/") ||
		p.UID == "1000"
}


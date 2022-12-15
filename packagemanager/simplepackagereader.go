package packagemanager

import (
	"errors"
	"fmt"
	"github.com/sephiroth74/go_adb_client/types"
	"regexp"
	"strconv"
	"strings"
)

type SimplePackageReader struct {
	Data string
}

// NewPackageReader Returns an instance of SimplePackageReader if data is successfully parsed.
// data is the output from "adb shell dump package..."
func NewPackageReader(data string) *SimplePackageReader {
	r := regexp.MustCompile("(?m)^Packages:\n")
	m := r.FindStringIndex(data)
	if m != nil {
		data = data[m[0]:]
		m2 := regexp.MustCompile("(?m)^$").FindStringIndex(data)
		if m2 != nil {
			data = data[:m2[1]]
			return &SimplePackageReader{
				Data: data,
			}
		}
	}
	return nil
}

func (s SimplePackageReader) Keys() []string {
	return []string{
		"package_name",
		"version_name",
		"code_path",
		"first_install_time",
		"last_update_time",
		"is_system_app",
		"is_debuggable",
		"user_id",
		"data_dir"}
}

func (s SimplePackageReader) PackageName() string {
	result, _ := s.parse(`(?m)^\s+Package\s+\[([^\]]+)\]\s+\([^\\)]+\):`)
	return result
}

func (s SimplePackageReader) Flags() []string {
	result, _ := s.parse(`(?m)^\s{3,}flags=\[\s*([^\]]+)\s*\]$`)
	return regexp.MustCompile(`\s+`).Split(strings.TrimSpace(result), -1)
}

func (s SimplePackageReader) FirstInstallTime() string {
	result, _ := s.getItem("firstInstallTime")
	return result
}

func (s SimplePackageReader) LastUpdateTime() string {
	result, _ := s.getItem("lastUpdateTime")
	return result
}

func (s SimplePackageReader) TimeStamp() string {
	result, _ := s.getItem("timeStamp")
	return result
}

func (s SimplePackageReader) DataDir() string {
	result, _ := s.getItem("dataDir")
	return result
}

func (s SimplePackageReader) UserID() string {
	result, _ := s.getItem(`userId`)
	return result
}

func (s SimplePackageReader) CodePath() string {
	result, _ := s.getItem("codePath")
	return result
}

func (s SimplePackageReader) ResourcePath() string {
	result, _ := s.getItem("resourcePath")
	return result
}

func (s SimplePackageReader) VersionName() string {
	result, _ := s.getItem("versionName")
	return result
}

func (s SimplePackageReader) InstallPermissions() []types.PackagePermission {
	f := regexp.MustCompile(`(?m)^\s{3,}install permissions:\n(?P<permissions>(\s{4,}[^\:]+:\s+granted=(true|false)\n)+)`)
	m := f.FindStringSubmatch(s.Data)
	var result []types.PackagePermission
	if m != nil {
		data := m[1]

		f = regexp.MustCompile(`(?m)^\s{4,}(?P<name>[^\:]+):\s+granted=(?P<granted>true|false)$`)
		match := f.FindAllStringSubmatch(data, -1)
		if match != nil {
			for _, v := range match {
				name := v[1]
				granted := v[2]
				boolValue, _ := strconv.ParseBool(granted)
				result = append(result, types.PackagePermission{
					Permission: types.Permission{Name: name},
					Granted:    boolValue,
					Flags:      nil,
				})
			}
		}
	}
	return result
}

func (s SimplePackageReader) RequestedPermissions() []types.RequestedPermission {
	f := regexp.MustCompile(`(?m)^\s{3,}requested permissions:\n((\s{4,}[\w\.]+$)+)`)
	m := f.FindStringSubmatch(s.Data)
	var result []types.RequestedPermission
	if m != nil {
		data := m[1]

		f = regexp.MustCompile(`(?m)^\s{4,}([\w\.]+)$`)
		match := f.FindAllStringSubmatch(data, -1)
		if match != nil {
			for _, v := range match {
				name := v[1]
				result = append(result, types.RequestedPermission{
					Permission: types.Permission{
						Name: name,
					},
				})
			}
		}
	}
	return result
}

func (s SimplePackageReader) getItem(name string) (string, error) {
	match := fmt.Sprintf(`(?m)^\s{3,}%s=(.*)$`, name)
	return s.parse(match)
}

func (s SimplePackageReader) parse(match string) (string, error) {
	r := regexp.MustCompile(match)
	m := r.FindStringSubmatch(s.Data)
	if len(m) == 2 {
		return m[1], nil
	}
	return "", errors.New(fmt.Sprintf("failed to find %s", match))
}

package packagemanager

import (
	"errors"
	"fmt"
	"regexp"
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

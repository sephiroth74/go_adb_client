package types

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var (
	deviceFileRegexp = regexp.MustCompile(`\s+`)
)

type DeviceFile struct {
	Line        string
	Permissions string
	LinkCount   int
	Owner       string
	Group       string
	Size        string
	Date        string
	Time        string
	Name        string
}

func NewDeviceFile(line string) (DeviceFile, error) {
	slice := deviceFileRegexp.Split(line, 8)
	if len(slice) == 8 {
		linkCount, _ := strconv.Atoi(slice[1])
		return DeviceFile{
			Line:        line,
			Permissions: slice[0],
			LinkCount:   linkCount,
			Owner:       slice[2],
			Group:       slice[3],
			Size:        slice[4],
			Date:        slice[5],
			Time:        slice[6],
			Name:        slice[7],
		}, nil
	} else {
		return DeviceFile{}, errors.New("not a valid file")
	}
}

func (d DeviceFile) String() string {
	return d.Line
}

func (d DeviceFile) IsSymlink() bool {
	return d.Permissions[0] == 'l'
}

func (d DeviceFile) IsDir() bool {
	return d.Permissions[0] == 'd'
}

func (d DeviceFile) Symlink() string {
	if d.IsSymlink() {
		s := strings.Split(d.Name, "->")
		return strings.TrimSpace(s[1])
	} else {
		return ""
	}
}

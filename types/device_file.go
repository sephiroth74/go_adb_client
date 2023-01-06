package types

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	deviceFileRegexp            = regexp.MustCompile(`\s+`)
	deviceFileDefaultTimeLayout = "2006-01-02 15:04"
)

type DeviceFile struct {
	Line        string
	Permissions string
	LinkCount   string
	Owner       string
	Group       string
	Size        string
	DateTime    time.Time
	Name        string
	Parent      string
}

type DeviceFileParser interface {
	Parse(string, string, string) (*DeviceFile, error)
}

type DefaultDeviceFileParser struct{}

func (d DefaultDeviceFileParser) Parse(parent string, line string, name string) (DeviceFile, error) {
	slice := deviceFileRegexp.Split(line, 8)
	if len(slice) == 8 {
		date, err := time.Parse(deviceFileDefaultTimeLayout, slice[5]+" "+slice[6])
		if err != nil {
			return DeviceFile{}, err
		}

		newname := slice[7]
		if name != "" {
			newname = name
		}

		return DeviceFile{
			Parent:      parent,
			Line:        line,
			Permissions: slice[0],
			LinkCount:   slice[1],
			Owner:       slice[2],
			Group:       slice[3],
			Size:        slice[4],
			DateTime:    date,
			Name:        newname,
		}, nil
	} else {
		return DeviceFile{}, errors.New("not a valid format")
	}
}

type StatDeviceFileParser struct{}

func (d StatDeviceFileParser) Parse(parent string, line string, name string) (DeviceFile, error) {
	slice := deviceFileRegexp.Split(line, 7)
	if len(slice) == 7 {
		i, err := strconv.ParseInt(slice[5], 10, 64)
		if err != nil {
			return DeviceFile{}, err
		}
		date := time.Unix(i, 0)

		newname := slice[6]
		if name != "" {
			newname = name
		}

		return DeviceFile{
			Parent:      parent,
			Line:        line,
			Permissions: slice[0],
			LinkCount:   slice[1],
			Owner:       slice[2],
			Group:       slice[3],
			Size:        slice[4],
			DateTime:    date,
			Name:        newname,
		}, nil
	} else {
		return DeviceFile{}, errors.New("not a valid format")
	}
}

func (d DeviceFile) String() string {
	datestr := d.DateTime.Format("2006-01-02 15:04")
	return fmt.Sprintf("%s %s %s %s %s %s %s", d.Permissions, d.LinkCount, d.Owner, d.Group, d.Size, datestr, d.Name)
	//return d.Line
}

func (d DeviceFile) IsSymlink() bool {
	return d.Permissions[0] == 'l'
}

func (d DeviceFile) IsDir() bool {
	return d.Permissions[0] == 'd'
}

func (d DeviceFile) Abs() string {
	if d.IsSymlink() {
		return d.Symlink()
	}
	return filepath.Clean(filepath.Join(d.Parent, d.Name))
}

func (d DeviceFile) Symlink() string {
	if d.IsSymlink() {
		s := strings.Split(d.Name, "->")
		return strings.TrimSpace(s[1])
	} else {
		return ""
	}
}

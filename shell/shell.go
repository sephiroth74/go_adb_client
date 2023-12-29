package shell

import (
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/magiconair/properties"
	"github.com/sephiroth74/go-processbuilder"
	"github.com/sephiroth74/go_adb_client/connection"
	"github.com/sephiroth74/go_adb_client/process"
	streams "github.com/sephiroth74/go_streams"

	"github.com/sephiroth74/go_adb_client/input"
	"github.com/sephiroth74/go_adb_client/types"
	"github.com/sephiroth74/go_adb_client/util/constants"
)

type Shell struct {
	Conn    *connection.Connection
	Address types.Serial
}

func NewShell(conn *connection.Connection, serial types.Serial) *Shell {
	var s = Shell{
		Address: serial,
		Conn:    conn,
	}
	return &s
}

func (s Shell) NewCommand() *process.ADBCommand {
	return s.Conn.NewAdbCommand().WithSerialAddr(&s.Address).WithCommand("shell")
}

func (s Shell) Execute(command string, args ...string) (process.OutputResult, error) {
	return process.SimpleOutput(s.NewCommand().WithArgs(command).AddArgs(args...), s.Conn.Verbose)
}

func (s Shell) Cat(filename string) (process.OutputResult, error) {
	return process.SimpleOutput(s.NewCommand().WithArgs("cat", filename), s.Conn.Verbose)
}

func (s Shell) Whoami() (process.OutputResult, error) {
	return process.SimpleOutput(s.NewCommand().WithArgs("whoami"), s.Conn.Verbose)
}

func (s Shell) Which(command string) (string, error) {
	cmd := s.NewCommand().WithArgs("which", command)
	result, err := process.SimpleOutput(cmd, s.Conn.Verbose)
	if err != nil {
		return "", err
	}

	if !result.IsOk() {
		return "", result.NewError()
	}

	return result.Output(), nil
}

func (s Shell) GetCommand(command string) (string, error) {
	cmd := s.NewCommand().WithArgs("command", "-v", command)
	result, err := process.SimpleOutput(cmd, s.Conn.Verbose)
	if err != nil {
		return "", err
	}

	if !result.IsOk() {
		return "", result.NewError()
	}

	return strings.TrimSpace(result.Output()), nil
}

func (s Shell) HasCommand(command string) (bool, error) {
	result, err := s.GetCommand(command)
	if err != nil {
		return false, err
	}

	if len(result) < 1 {
		return false, fmt.Errorf("command %s not found", command)
	} else {
		return true, nil
	}
}

func (s Shell) CheckAvbctl() error {
	status, err := s.HasCommand("avbctl")
	if err != nil {
		return err
	}

	if !status {
		return fmt.Errorf("avbctl not found")
	}

	return nil
}

func (s Shell) GetVerityStatus() (bool, error) {
	err := s.CheckAvbctl()
	if err != nil {
		return false, err
	}

	status, err := process.SimpleOutput(s.NewCommand().WithArgs("avbctl", "get-verity"), s.Conn.Verbose)

	if err != nil {
		return false, err
	}

	if status.IsOk() {
		return strings.Contains(status.Output(), "enabled"), nil
	} else {
		return false, status.NewError()
	}
}

func (s Shell) DisableVerity() error {
	err := s.CheckAvbctl()
	if err != nil {
		return err
	}

	status, err := process.SimpleOutput(s.NewCommand().WithArgs("avbctl", "disable-verity"), s.Conn.Verbose)
	if err != nil {
		return err
	}

	if !status.IsOk() {
		return status.NewError()
	}

	return nil
}

func (s Shell) EnableVerity() error {
	err := s.CheckAvbctl()
	if err != nil {
		return err
	}

	status, err := process.SimpleOutput(s.NewCommand().WithArgs("avbctl", "enable-verity"), s.Conn.Verbose)
	if err != nil {
		return err
	}

	if !status.IsOk() {
		return status.NewError()
	}

	return nil
}

// GetProp ExecuteWithTimeout the command "adb shell getprop key" and returns its value if found, nil otherwise
// Deprecated use GetPropValue instead
func (s Shell) GetProp(key string) *string {
	result, err := process.SimpleOutput(s.NewCommand().WithArgs("getprop", key), s.Conn.Verbose)
	if err != nil {
		return nil
	}

	if result.IsOk() {
		trim := result.Output()
		return &trim
	} else {
		return nil
	}
}

// GetPropValue return the value of the given property key
func (s Shell) GetPropValue(key string) (string, error) {
	result, err := s.Execute("getprop", key)
	if err != nil {
		return "", err
	}

	if result.IsOk() {
		trim := strings.TrimSpace(result.Output())
		return trim, nil
	} else {
		return "", result.NewError()
	}
}

// GetPropType Returns the property type.
// Can be string, int, bool, enum [list string]
func (s Shell) GetPropType(key string) (*string, bool) {
	result, err := process.SimpleOutput(s.NewCommand().WithArgs("getprop", "-T", key), s.Conn.Verbose)
	if err != nil {
		return nil, false
	}

	if result.IsOk() {
		trim := result.Output()
		return &trim, true
	} else {
		return nil, false
	}
}

func (s Shell) GetProps() (*properties.Properties, error) {
	result, err := process.SimpleOutput(s.NewCommand().WithArgs("getprop"), s.Conn.Verbose)
	if err != nil {
		return nil, err
	}

	if result.IsOk() {
		props := properties.NewProperties()
		pairs, err := parsePropLines(result.Output())
		for _, t := range pairs {
			if err != nil {
				return nil, err
			}

			if _, _, err := props.Set(t.First, t.Second); err != nil {
				return nil, err
			}
		}
		return props, nil
	} else {
		return nil, result.NewError()
	}
}

func (s Shell) SetProp(key string, value string) bool {
	newvalue := value
	if newvalue == "" {
		newvalue = "\"\""
	}

	result, err := process.SimpleOutput(s.NewCommand().WithArgs("setprop", key, newvalue).WithTimeout(constants.DEFAULT_TIMEOUT), s.Conn.Verbose)
	if err != nil {
		return false
	}
	return result.IsOk()
}

// ClearProp set the given property key to an empty value.
func (s Shell) ClearProp(key string) bool {
	return s.SetProp(key, "")
}

func (s Shell) Exists(filename string) bool {
	return testFile(s, filename, "e")
}

func (s Shell) IsFile(filename string) bool {
	return testFile(s, filename, "f")
}

func (s Shell) IsDir(filename string) bool {
	return testFile(s, filename, "d")
}

func (s Shell) IsSymlink(filename string) bool {
	return testFile(s, filename, "h")
}

func (s Shell) Remove(filename string, force bool) (bool, error) {
	var command string
	if force {
		command = fmt.Sprintf("rm -f %s", filename)
	} else {
		command = fmt.Sprintf("rm %s", filename)
	}

	cmd := s.NewCommand().WithArgs(command)
	result, err := process.SimpleOutput(cmd, s.Conn.Verbose)
	if err != nil {
		return false, nil
	}
	return result.IsOk(), nil
}

func (s Shell) RemoveDir(filename string, force bool) (bool, error) {
	var command string
	if force {
		command = fmt.Sprintf("rm -fr %s", filename)
	} else {
		command = fmt.Sprintf("rm -r %s", filename)
	}

	cmd := s.NewCommand().WithArgs(command)
	result, err := process.SimpleOutput(cmd, s.Conn.Verbose)
	if err != nil {
		return false, nil
	}
	return result.IsOk(), nil
}

func (s Shell) Chmod(mode os.FileMode, recursive bool, filename string) error {
	var sb []string
	if recursive {
		sb = append(sb, "-R")
	}
	sb = append(sb, mode.String())
	sb = append(sb, filename)

	res, err := s.Execute("chmod", sb...)
	if err != nil {
		return err
	}
	if !res.IsOk() {
		return res.NewError()
	}
	return nil
}

func (s Shell) ChmodString(mode string, recursive bool, filename string) error {
	var sb []string
	if recursive {
		sb = append(sb, "-R")
	}
	sb = append(sb, mode)
	sb = append(sb, filename)

	res, err := s.Execute("chmod", sb...)
	if err != nil {
		return err
	}
	if !res.IsOk() {
		return res.NewError()
	}
	return nil
}

func (s Shell) Stat(filename string) (fs.FileMode, error) {
	res, err := s.Execute("stat", "-L -c '%a'", filename)
	if err != nil {
		return 0, err
	}
	if !res.IsOk() {
		return 0, res.NewError()
	}

	octal := fmt.Sprintf("%04s", res.Output())
	parseInt, err := strconv.ParseInt(octal, 0, 32)
	if err != nil {
		return 0, err
	}
	return fs.FileMode(parseInt), nil
}

func (s Shell) Statf(format string, filename string) (string, error) {
	res, err := s.Execute("stat", fmt.Sprintf("-L -c \"%s\"", format), filename)
	if err != nil {
		return "", err
	}
	if !res.IsOk() {
		return "", res.NewError()
	}
	return res.Output(), nil
}

func (s Shell) Tap(source input.InputSource, pos types.Pair[int, int]) (process.OutputResult, error) {
	cmd := s.NewCommand().WithArgs("input", source.String(), "tap")
	cmd.AddArgs(fmt.Sprintf("%v", pos.First))
	cmd.AddArgs(fmt.Sprintf("%v", pos.Second))
	return process.SimpleOutput(cmd, s.Conn.Verbose)
}

func (s Shell) Roll(source input.InputSource, pos types.Pair[int, int]) (process.OutputResult, error) {
	cmd := s.NewCommand().WithArgs("input", source.String(), "roll")
	cmd.AddArgs(fmt.Sprintf("%v", pos.First))
	cmd.AddArgs(fmt.Sprintf("%v", pos.Second))
	return process.SimpleOutput(cmd, s.Conn.Verbose)
}

func (s Shell) Press(source input.InputSource) (process.OutputResult, error) {
	cmd := s.NewCommand().WithArgs("input", source.String(), "press")
	return process.SimpleOutput(cmd, s.Conn.Verbose)
}

func (s Shell) Swipe(source input.InputSource, duration int32, pos1 types.Pair[int, int], pos2 types.Pair[int, int]) (process.OutputResult, error) {
	cmd := s.NewCommand().WithArgs("input", source.String(), "swipe")
	cmd.AddArgs(fmt.Sprintf("%v", pos1.First))
	cmd.AddArgs(fmt.Sprintf("%v", pos1.Second))
	cmd.AddArgs(fmt.Sprintf("%v", pos2.First))
	cmd.AddArgs(fmt.Sprintf("%v", pos2.Second))
	if duration > 0 {
		cmd.AddArgs(fmt.Sprintf("%v", duration))
	}
	return process.SimpleOutput(cmd, s.Conn.Verbose)
}

func (s Shell) MotionEvent(source input.InputSource, event_type input.MotionEvent, pos types.Pair[int, int]) (process.OutputResult, error) {
	cmd := s.NewCommand().WithArgs("input", source.String(), "motionevent", event_type.String())
	cmd.AddArgs(fmt.Sprintf("%v", pos.First))
	cmd.AddArgs(fmt.Sprintf("%v", pos.Second))
	return process.SimpleOutput(cmd, s.Conn.Verbose)
}

func (s Shell) SendEvent(device string, code_type int, code int, value int) (process.OutputResult, error) {
	cmd := s.NewCommand().WithArgs("sendevent", device, fmt.Sprintf("%d", code_type), fmt.Sprintf("%d", code), fmt.Sprintf("%d", value))
	return process.SimpleOutput(cmd, s.Conn.Verbose)
}

func (s Shell) SendKeyEvent(source input.InputSource, event_type *input.KeyEventType, event input.KeyCode) (process.OutputResult, error) {
	return s.SendKeyEvents(source, event_type, event)
}

func (s Shell) SendKeyEvents(source input.InputSource, event_type *input.KeyEventType, events ...input.KeyCode) (process.OutputResult, error) {
	var format = make([]string, len(events))
	for i, v := range events {
		format[i] = v.String()
	}

	cmd := s.NewCommand().WithArgs("input")
	cmd.AddArgs(source.String())
	cmd.AddArgs("keyevent")

	if event_type != nil {
		cmd.AddArgs(event_type.String())
	}

	cmd.AddArgs(strings.Join(format, " "))
	return process.SimpleOutput(cmd, s.Conn.Verbose)
}

func (s Shell) SendChar(code rune) (process.OutputResult, error) {
	return process.SimpleOutput(s.NewCommand().WithArgs(fmt.Sprintf("input text %c", code)), s.Conn.Verbose)
}

func (s Shell) SendString(value string) (process.OutputResult, error) {
	cmd := s.NewCommand().WithArgs(fmt.Sprintf("input text '%s'", value))
	return process.SimpleOutput(cmd, s.Conn.Verbose)
}

// GetEvents Returns a slice of Pairs each one containing the event type and the event name
func (s Shell) GetEvents() ([]types.Pair[string, string], error) {
	cmd := s.NewCommand().WithArgs("getevent", "-S")
	result, err := process.SimpleOutput(cmd, s.Conn.Verbose)
	if err != nil {
		return nil, err
	}

	arr := parseEvents(result.Output())
	return arr, nil
}

func (s Shell) ScreenRecord(options ScreenRecordOptions, filename string) (*processbuilder.Processbuilder, error) {
	args := []string{"screenrecord"}
	args = append(args, "--bit-rate", fmt.Sprintf("%d", options.Bitrate))
	poption := processbuilder.Option{}

	if options.Timelimit > 0 {
		args = append(args, "--time-limit", fmt.Sprintf("%d", options.Timelimit))
		poption.Timeout = time.Second * time.Duration(options.Timelimit)
	}

	if options.Rotate {
		args = append(args, "--rotate")
	}

	if options.BugReport {
		args = append(args, "--bugreport")
	}

	if options.Verbose {
		args = append(args, "--verbose")
	}

	if options.Size != nil {
		args = append(args, "--size", options.Size.String())
	}

	args = append(args, filename)

	return processbuilder.Create(
		poption,
		s.NewCommand().WithArgs(args...).ToCommand(),
	)
}

func (s Shell) ListDir(dirname string) ([]types.DeviceFile, error) {
	var emptyList []types.DeviceFile

	if !s.IsDir(dirname) {
		return emptyList, os.ErrNotExist
	}

	result, err := process.SimpleOutput(s.NewCommand().WithArgs("ls -lLHap --color=none", dirname), s.Conn.Verbose)

	if err != nil {
		return emptyList, err
	}
	if !result.IsOk() {
		return emptyList, err
	}

	parser := types.DefaultDeviceFileParser{}

	deviceFiles := streams.MapNotNull(result.OutputLines(false), func(line string) (types.DeviceFile, error) {
		return parser.Parse(dirname, line, "")
	})

	//statsParser := types.StatDeviceFileParser{}
	//
	//if streams.IndexOf(deviceFiles, func(file types.DeviceFile) bool {
	//	return file.Name == "./"
	//}) == -1 {
	//	statf, err := s.Statf("%A %h %U %G %b %Y %n", dirname)
	//	if err == nil {
	//		file, err := statsParser.Parse(filepath.Dir(dirname), statf, "./")
	//		if err == nil {
	//			deviceFiles = slices.Insert(deviceFiles, 0, file)
	//		}
	//	}
	//}
	//
	//if streams.IndexOf(deviceFiles, func(file types.DeviceFile) bool {
	//	return file.Name == "../"
	//}) == -1 {
	//	statf2, err := s.Statf("%A %h %U %G %b %Y %n", filepath.Dir(dirname))
	//	if err == nil {
	//		file, err := statsParser.Parse(filepath.Dir(filepath.Dir(dirname)), statf2, "../")
	//		if err == nil {
	//			if len(deviceFiles) > 1 {
	//				deviceFiles = slices.Insert(deviceFiles, 1, file)
	//			} else if len(deviceFiles) == 1 {
	//				deviceFiles = append(deviceFiles, file)
	//			}
	//		}
	//	}
	//}

	return deviceFiles, nil
}

func (s Shell) ListSettings(namespace types.SettingsNamespace) (*properties.Properties, error) {
	cmd := s.NewCommand().WithArgs(fmt.Sprintf("settings list %s", namespace))
	result, err := process.SimpleOutput(cmd, s.Conn.Verbose)
	if err != nil {
		return nil, err
	}
	p, err := properties.LoadString(result.Output())
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s Shell) GetSetting(key string, namespace types.SettingsNamespace) (*string, error) {
	cmd := s.NewCommand().WithArgs(fmt.Sprintf("settings get %s %s", namespace, key))
	result, err := process.SimpleOutput(cmd, s.Conn.Verbose)

	if err != nil {
		return nil, err
	}

	if !result.IsOk() {
		return nil, result.NewError()
	}

	value := result.Output()
	if strings.EqualFold("null", value) {
		return nil, nil
	}
	return &value, nil
}

func (s Shell) PutSetting(key string, value string, namespace types.SettingsNamespace) error {
	cmd := s.NewCommand().WithArgs(fmt.Sprintf("settings put %s %s %s", namespace, key, value))
	result, err := process.SimpleOutput(cmd, s.Conn.Verbose)

	if err != nil {
		return err
	}

	if !result.IsOk() {
		return result.NewError()
	}

	return nil
}

func (s Shell) DeleteSetting(key string, namespace types.SettingsNamespace) error {
	cmd := s.NewCommand().WithArgs(fmt.Sprintf("settings delete %s %s", namespace, key))
	result, err := process.SimpleOutput(cmd, s.Conn.Verbose)

	if err != nil {
		return err
	}

	if !result.IsOk() {
		return result.NewError()
	}

	return nil
}

// DumpSys is a tool that runs on Android devices and provides information about system services.
// For a complete list of services available use ListDumpSys
func (s Shell) DumpSys(name string) (process.OutputResult, error) {
	return process.SimpleOutput(s.NewCommand().WithArgs("dumpsys", name), s.Conn.Verbose)
}

// ListDumpSys return the complete list of system services that can be used with dumpsys
func (s Shell) ListDumpSys() ([]string, error) {
	result, err := s.Execute("dumpsys", "-l")
	var emptylist []string
	if err != nil {
		return emptylist, err
	}
	if !result.IsOk() {
		return emptylist, result.NewError()
	}

	return result.OutputLines(false), nil
}

//

func parseEvents(text string) []types.Pair[string, string] {
	arr := []types.Pair[string, string]{}
	f := regexp.MustCompile(`^add\s+device\s+[0-9]+:\s(?P<event>[^\n]+)\s*name:\s*"(?P<name>[^"]+)"\n?`)
	for {
		m := f.FindStringSubmatchIndex(text)
		if len(m) == 6 {
			event := text[m[2]:m[3]]
			name := text[m[4]:m[5]]
			arr = append(arr, types.Pair[string, string]{
				First:  event,
				Second: name,
			})
		} else {
			break
		}
		text = text[m[1]:]
	}
	return arr
}

func parsePropLines(text string) ([]types.Pair[string, string], error) {
	f := regexp.MustCompile(`(?m)^\[(.*)\]\s*:\s*\[([^\]]*)\]$`)
	m := f.FindAllStringSubmatch(text, -1)
	return streams.Map(m, func(match []string) types.Pair[string, string] {
		return types.Pair[string, string]{
			First:  match[1],
			Second: match[2],
		}
	}), nil

	//m := f.FindStringSubmatch(line)
	//if len(m) == 3 {
	//	return types.Pair[string, string]{
	//		First:  m[1],
	//		Second: m[2],
	//	}, nil
	//} else {
	//	return types.Pair[string, string]{}, errors.New("parse exception. cannot find submatches on the fiven line")
	//}
}

func testFile(shell Shell, filename string, mode string) bool {
	cmd := shell.NewCommand().WithArgs(fmt.Sprintf("test -%s %s && echo 1 || echo 0", mode, filename))
	result, err := process.SimpleOutput(cmd, shell.Conn.Verbose)
	// result, err := shell.ExecuteWithTimeout(fmt.Sprintf("test -%s %s && echo 1 || echo 0", mode, filename), 0)
	if err != nil || !result.IsOk() {
		return false
	}
	return result.Output() == "1"
}

// types

type ScreenRecordOptions struct {
	// --bit-rate 4000000
	// Set the video bit rate, in bits per second. Value may be specified as bits or megabits, e.g. '4000000' is equivalent to '4M'.
	// Default 20Mbps.
	Bitrate uint64

	// --time-limit=120 (in seconds)
	// Set the maximum recording time, in seconds. Default / maximum is 180
	Timelimit uint

	// --rotate
	// Rotates the output 90 degrees. This feature is experimental.
	Rotate bool

	// --bugreport
	// Add additional information, such as a timestamp overlay, that is helpful in videos captured to illustrate bugs.
	BugReport bool

	// --size 1280x720
	// Set the video size, e.g. "1280x720". Default is the device's main display resolution (if supported), 1280x720 if not.
	// For best results, use a size supported by the AVC encoder.
	Size *types.Size

	// --verbose
	// Display interesting information on stdout
	Verbose bool
}

func NewScreenRecordOptions() ScreenRecordOptions {
	return ScreenRecordOptions{
		Bitrate:   20000000,
		Timelimit: 180,
		Rotate:    false,
		BugReport: false,
		Size:      nil,
		Verbose:   false,
	}
}

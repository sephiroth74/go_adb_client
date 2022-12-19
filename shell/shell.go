package shell

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sephiroth74/go_adb_client/input"
	"github.com/sephiroth74/go_adb_client/transport"
	"github.com/sephiroth74/go_adb_client/types"
	"github.com/sephiroth74/go_adb_client/util"
	"github.com/sephiroth74/go_adb_client/util/constants"
)

type Shell struct {
	Address types.Serial
	Adb     *string
}

func NewShell(adb *string, serial types.Serial) *Shell {
	var s = Shell{
		Address: serial,
		Adb:     adb,
	}
	return &s
}

func (s Shell) Execute(command string, timeout time.Duration, args ...string) (transport.Result, error) {
	return s.NewProcess().WithTimeout(timeout).WithArgs(command).WithArgs(args...).Verbose(false).Invoke()
}

func (s Shell) Executef(format string, timeout time.Duration, v ...any) (transport.Result, error) {
	return s.NewProcess().WithTimeout(timeout).WithArgs(fmt.Sprintf(format, v...)).Verbose(false).Invoke()
}

func (s Shell) NewProcess() *transport.ProcessBuilder {
	return transport.NewProcessBuilder().WithSerial(&s.Address).WithPath(s.Adb).WithCommand("shell")
}

func (s Shell) Cat(filename string) (transport.Result, error) {
	return s.Execute("cat", 0, filename)
}

func (s Shell) Whoami() (transport.Result, error) {
	return s.Execute("whoami", 0)
}

func (s Shell) Which(command string) (transport.Result, error) {
	return s.Execute("which", 0, command)
}

// GetProp Execute the command "adb shell getprop key" and returns its value
// if found, nil otherwise
func (s Shell) GetProp(key string) *string {
	result, err := s.Execute("getprop", 0, key)
	if err != nil {
		return nil
	}

	if result.IsOk() {
		trim := strings.TrimSpace(result.Output())
		return &trim
	} else {
		return nil
	}
}

// GetPropType Returns the property type.
// Can be string, int, bool, enum [list string]
func (s Shell) GetPropType(key string) (*string, bool) {
	result, err := s.Execute("getprop", 0, "-T", key)
	if err != nil {
		return nil, false
	}

	if result.IsOk() {
		trim := strings.TrimSpace(result.Output())
		return &trim, true
	} else {
		return nil, false
	}
}

func (s Shell) GetProps() ([]types.Pair[string, string], error) {
	result, err := s.Execute("getprop", 0)
	if err != nil {
		return nil, err
	}

	if result.IsOk() {
		mapped := util.MapNotNull(result.OutputLines(), func(s string) (types.Pair[string, string], error) {
			t, err := parsePropLine(s)
			return t, err
		})
		return mapped, nil
	} else {
		return []types.Pair[string, string]{}, nil
	}
}

func (s Shell) SetProp(key string, value string) bool {
	result, err := s.Execute("setprop", constants.DEFAULT_TIMEOUT, key, value)
	if err != nil {
		return false
	}
	return result.IsOk()
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
	result, err := s.Execute(command, 0)
	if err != nil {
		return false, nil
	}
	return result.IsOk(), nil
}

func (s Shell) SendKeyEvent(event input.KeyCode) (transport.Result, error) {
	return s.SendKeyEvents(event)
}

func (s Shell) SendKeyEvents(events ...input.KeyCode) (transport.Result, error) {
	var format = make([]string, len(events))
	for i, v := range events {
		format[i] = fmt.Sprintf("%s", v.String())
	}
	return s.Executef("input keyevent %s", 0, strings.Join(format, " "))
}

func (s Shell) SendChar(code rune) (transport.Result, error) {
	return s.Executef("input text %c", 0, code)
}

func (s Shell) SendString(value string) (transport.Result, error) {
	return s.Executef("input text '%s'", 0, value)
}

// Returns a slice of Pairs each one containing the event type and the event name
func (s Shell) GetEvents() ([]types.Pair[string, string], error) {
	result, err := s.Execute("getevent", 0, "-p")
	if err != nil {
		return nil, err
	}

	arr := parseEvents(result.Output())
	return arr, nil
}

//

func parseEvents(text string) []types.Pair[string, string] {
	arr := []types.Pair[string, string]{}
	f := regexp.MustCompile(`add device [0-9]+:\s(?P<event>[^\n]+)\s*name:\s*"(?P<name>[^"]+)"`)
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

func parsePropLine(line string) (types.Pair[string, string], error) {
	f := regexp.MustCompile(`^\[(.*)\]\s*:\s*\[(.*)\]\s*$`)
	m := f.FindStringSubmatch(line)
	if len(m) == 3 {
		return types.Pair[string, string]{
			First:  m[1],
			Second: m[2],
		}, nil
	} else {
		return types.Pair[string, string]{}, errors.New("parse exception. cannot find submatches on the fiven line")
	}
}

func testFile(shell Shell, filename string, mode string) bool {
	result, err := shell.Execute(fmt.Sprintf("test -%s %s && echo 1 || echo 0", mode, filename), 0)
	if err != nil || !result.IsOk() {
		return false
	}
	return result.Output() == "1"
}

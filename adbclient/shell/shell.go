package shell

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"it.sephiroth/adbclient/transport"
	"it.sephiroth/adbclient/types"
	"it.sephiroth/adbclient/util"
	"it.sephiroth/adbclient/util/constants"
)

type Shell[T types.Serial] struct {
	Serial T
	Adb    *string
}

func NewShell[T types.Serial](adb *string, serial T) *Shell[T] {
	var s = new(Shell[T])
	s.Serial = serial
	s.Adb = adb
	return s
}

func (s Shell[T]) Execute(command string, timeout time.Duration, args ...string) (transport.Result, error) {
	return transport.NewProcessBuilder(s.Serial).Path(s.Adb).Command("shell").Timeout(timeout).Args(command).Args(args...).Verbose(false).Invoke()
}

func (s Shell[T]) Cat(filename string) (transport.Result, error) {
	return s.Execute("cat", 0, filename)
}

func (s Shell[T]) Whoami() (transport.Result, error) {
	return s.Execute("whoami", constants.DEFAULT_TIMEOUT)
}

func (s Shell[T]) Which(command string) (transport.Result, error) {
	return s.Execute("which", constants.DEFAULT_TIMEOUT, command)
}

// Execute the command "adb shell getprop key" and returns its value
// if found, nil otherwise
func (s Shell[T]) GetProp(key string) *string {
	result, err := s.Execute("getprop", constants.DEFAULT_TIMEOUT, key)
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

// Returns the property type.
// Can be string, int, bool, enum [list string]
func (s Shell[T]) GetPropType(key string) (*string, bool) {
	result, err := s.Execute("getprop", constants.DEFAULT_TIMEOUT, "-T", key)
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

func (s Shell[T]) GetProps() ([]types.Pair[string, string], error) {
	result, err := s.Execute("getprop", constants.DEFAULT_TIMEOUT)
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

func (s Shell[T]) SetProp(key string, value string) bool {
	result, err := s.Execute("setprop", constants.DEFAULT_TIMEOUT, key, value)
	if err != nil {
		return false
	}
	return result.IsOk()
}

//

func parsePropLine(line string) (types.Pair[string, string], error) {
	f := regexp.MustCompile(`^\[(.*)\]\s*:\s*\[(.*)\]\s*$`)
	m := f.FindStringSubmatch(line)
	if len(m) == 3 {
		return types.Pair[string, string]{
			First:  m[1],
			Second: m[2],
		}, nil
	} else {
		return types.Pair[string, string]{}, errors.New("Parse exception. Cannot find submatches on the fiven line")
	}
}

package process

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/sephiroth74/go-processbuilder"
	"github.com/sephiroth74/go_adb_client/types"
)

type Command struct {
	Command string
	Args    []string
}

func (c *Command) AddArgs(args ...string) {
	c.Args = append(c.Args, args...)
}

func NewCommand(command string, args ...string) *Command {
	return &Command{Command: command, Args: args}
}

type ADBCommand struct {
	ADBPath    string
	ADBCommand string
	Serial     string
	Args       []string
	Timeout    time.Duration
}

func NewADBCommand(path string) *ADBCommand {
	return &ADBCommand{ADBPath: path, Timeout: 0}
}

func (a *ADBCommand) WithCommand(command string) *ADBCommand {
	a.ADBCommand = command
	return a
}

func (a *ADBCommand) WithSerial(serial string) *ADBCommand {
	a.Serial = serial
	return a
}

func (a *ADBCommand) WithSerialAddr(t *types.Serial) *ADBCommand {
	a.Serial = (*t).GetSerialAddress()
	return a
}

func (a *ADBCommand) Withargs(args ...string) *ADBCommand {
	a.Args = args
	return a
}

func (a *ADBCommand) AddArgs(args ...string) *ADBCommand {
	a.Args = append(a.Args, args...)
	return a
}

func (a *ADBCommand) WithTimeout(time time.Duration) *ADBCommand {
	a.Timeout = time
	return a
}

func (a *ADBCommand) FullArgs() []string {
	var args = []string{}
	if a.Serial != "" {
		args = append(args, "-s", a.Serial)
	}

	if a.ADBCommand != "" {
		args = append(args, a.ADBCommand)
	}

	args = append(args, a.Args...)
	return args
}

func NewSuccessOutputResult(message string) OutputResult {
	return OutputResult{ExitCode: 0, StdOut: *bytes.NewBufferString(message)}
}

func NewErrorOutputResult(message string) OutputResult {
	return OutputResult{ExitCode: 1, StdErr: *bytes.NewBufferString(message)}
}

type OutputResult struct {
	ExitCode int
	StdOut   bytes.Buffer
	StdErr   bytes.Buffer
}

func (o OutputResult) IsOk() bool {
	return o.ExitCode == 0
}

func (o OutputResult) NewError() error {
	return fmt.Errorf("invalid exit code: %d\n%s", o.ExitCode, o.Error())
}

func (o OutputResult) Error() string {
	return strings.TrimSpace(o.StdErr.String())
}

func (o OutputResult) HasError() bool {
	return len(o.StdErr.Bytes()) > 0
}

func (o OutputResult) Output() string {
	return strings.TrimSpace(o.StdOut.String())
}

func (o OutputResult) OutputLines() []string {
	return strings.Split(strings.TrimSpace(o.StdOut.String()), "\n")
}

func (o OutputResult) String() string {
	return fmt.Sprintf("OutputResult(isOk=`%t`, Stdout=`%s`, Stderr=`%s`, ExitCode=%d)", o.IsOk(), o.Output(), o.Error(), o.ExitCode)
}

func SimpleOutput(command *ADBCommand, verbose bool) (OutputResult, error) {
	option := processbuilder.Option{
		Timeout: command.Timeout,
	}

	if verbose {
		option.LogLevel = zerolog.DebugLevel
	} else {
		option.LogLevel = zerolog.Disabled
	}

	sout, serr, code, err := processbuilder.Output(
		option,
		processbuilder.Command(command.ADBPath, command.FullArgs()...),
	)

	if sout == nil {
		sout = &bytes.Buffer{}
	}

	if serr == nil {
		serr = &bytes.Buffer{}
	}

	result := OutputResult{
		ExitCode: code,
		StdOut:   *sout,
		StdErr:   *serr,
	}

	return result, err
}

package process

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/sephiroth74/go-processbuilder"
	"github.com/sephiroth74/go_adb_client/types"
	streams "github.com/sephiroth74/go_streams"
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
	StdOut     io.Writer
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

func (a *ADBCommand) WithArgs(args ...string) *ADBCommand {
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

func (a *ADBCommand) WithStdOut(writer io.Writer) *ADBCommand {
	a.StdOut = writer
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

func (a *ADBCommand) ToCommand() *processbuilder.Command {
	cmd := processbuilder.NewCommand(a.ADBPath, a.FullArgs()...)
	if a.StdOut != nil {
		cmd.WithStdOut(a.StdOut)
	}
	return cmd
}

func NewSuccessOutputResult(message string) OutputResult {
	return OutputResult{ExitCode: 0, StdOut: *bytes.NewBufferString(message)}
}

func NewErrorOutputResult(message string) OutputResult {
	return OutputResult{ExitCode: 1, StdErr: *bytes.NewBufferString(message)}
}

type OutputResult struct {
	ExitCode   int
	ExitStatus *os.ProcessState
	StdOut     bytes.Buffer
	StdErr     bytes.Buffer
}

func (o OutputResult) IsOk() bool {
	return o.ExitCode == 0
}

func (o OutputResult) IsInterrupted() bool {
	return o.ExitCode == int(syscall.SIGINT)
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

func (o OutputResult) OutputLines(trim bool) []string {
	splitted := strings.Split(o.StdOut.String(), "\n")
	if trim {
		return streams.Map(splitted, func(line string) string {
			return strings.TrimSpace(line)
		})
	} else {
		return splitted
	}
}

func (o OutputResult) String() string {
	return fmt.Sprintf("OutputResult(isOk=`%t`, Stdout=`%s`, Stderr=`%s`, ExitCode=%d, ExitStatus=%#v)", o.IsOk(), o.Output(), o.Error(), o.ExitCode, o.ExitStatus)
}

func SimpleOutput(command *ADBCommand, verbose bool) (OutputResult, error) {
	option := processbuilder.Option{
		Timeout: command.Timeout,
	}

	if verbose {
		option.LogLevel = log.TraceLevel
	} else {
		option.LogLevel = log.InfoLevel
	}

	cmd := processbuilder.NewCommand(command.ADBPath, command.FullArgs()...)

	if command.StdOut != nil {
		cmd.WithStdOut(command.StdOut)
	}

	sout, serr, code, state, err := processbuilder.Output(
		option,
		cmd,
	)

	if sout == nil {
		sout = &bytes.Buffer{}
	}

	if serr == nil {
		serr = &bytes.Buffer{}
	}

	result := OutputResult{
		ExitCode:   code,
		ExitStatus: state,
		StdOut:     *sout,
		StdErr:     *serr,
	}

	return result, err
}

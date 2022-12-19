package transport

import (
	"bufio"
	"os"

	"github.com/sephiroth74/go_adb_client/logging"
	"github.com/sephiroth74/go_adb_client/types"

	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alecthomas/repr"
)

type Result struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
}

func (r Result) IsOk() bool {
	return r.ExitCode == 0
}

func (r Result) NewError() error {
	return fmt.Errorf("invalid exit code: %d", r.ExitCode)
}

func (r Result) Output() string {
	return strings.TrimSpace(string(r.Stdout))
}

func (r Result) OutputLines() []string {
	return strings.Split(strings.TrimSpace(string(r.Stdout)), "\n")
}

func (r Result) Error() string {
	return strings.TrimSpace(string(r.Stderr))
}

func (r Result) ToString() string {
	return fmt.Sprintf("Result(isOk=`%t`, Stdout=`%s`, Stderr=`%s`)", r.IsOk(), strings.TrimSpace(string(r.Stdout)), strings.TrimSpace(string(r.Stderr)))
}

func (r Result) String() string {
	return r.ToString()
}

func (r Result) Repr() string {
	return repr.String(r)
}

func ErrorResult(err string) Result {
	r := Result{
		ExitCode: 1,
		Stderr:   []byte(err),
	}
	return r
}

func OkResult(str string) Result {
	r := Result{
		ExitCode: 0,
		Stdout:   []byte(str),
	}
	return r
}

type TransportCommand struct {
	path    *string
	command *string
	args    []string
}

type ProcessBuilder struct {
	serial  string
	timeout time.Duration
	command *TransportCommand
	verbose bool
	stdout  *os.File
}

func NewProcessBuilder() *ProcessBuilder {
	b := new(ProcessBuilder)
	b.serial = ""
	b.timeout = 0
	b.command = &TransportCommand{
		path:    nil,
		command: nil,
		args:    []string{},
	}
	b.verbose = false
	b.stdout = nil
	return b
}

func (p *ProcessBuilder) WithSerial(t *types.Serial) *ProcessBuilder {
	p.serial = (*t).GetSerialAddress()
	return p
}

func (p *ProcessBuilder) WithSerialAddr(t string) *ProcessBuilder {
	p.serial = t
	return p
}

func (p *ProcessBuilder) WithArgs(args ...string) *ProcessBuilder {
	p.command.args = append(p.command.args, args...)
	return p
}

func (p *ProcessBuilder) Verbose(value bool) *ProcessBuilder {
	p.verbose = value
	return p
}

func (p *ProcessBuilder) WithStdout(value *os.File) *ProcessBuilder {
	p.stdout = value
	return p
}

func (p *ProcessBuilder) WithTimeout(time time.Duration) *ProcessBuilder {
	p.timeout = time
	return p
}

func (p *ProcessBuilder) WithPath(path *string) *ProcessBuilder {
	p.command.path = path
	return p
}

func (p *ProcessBuilder) WithCommand(command string) *ProcessBuilder {
	p.command.command = &command
	return p
}

func (p *ProcessBuilder) Invoke() (Result, error) {
	if p.verbose {
		logging.Log.Debug().Msgf(repr.String(p.command))
	}

	var adb = filepath.Base(*p.command.path)
	var finalArgs []string

	if p.serial != "" {
		finalArgs = append(finalArgs, "-s", p.serial)
	}

	finalArgs = append(finalArgs, *p.command.command)
	finalArgs = append(finalArgs, p.command.args...)

	var cmd *exec.Cmd = nil

	if p.timeout > 0 {
		logging.Log.Debug().Msgf("Executing (timeout=%s) `%s %s`", p.timeout.String(), adb, strings.Join(finalArgs, " "))
		var ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
		defer cancel()
		cmd = exec.CommandContext(ctx, *p.command.path, finalArgs...)
	} else {
		logging.Log.Debug().Msgf("Executing `%s %s`", adb, strings.Join(finalArgs, " "))
		cmd = exec.Command(*p.command.path, finalArgs...)
	}

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	if p.stdout != nil {
		cmd.Stdout = bufio.NewWriter(p.stdout)
	}

	err := cmd.Run()
	exitCode := cmd.ProcessState.ExitCode()

	var result = Result{
		ExitCode: exitCode,
		Stdout:   outb.Bytes(),
		Stderr:   errb.Bytes(),
	}

	if err != nil {
		return result, err
	}

	return result, nil

}

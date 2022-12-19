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

var log = logging.GetLogger("transport")

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

func Invoke(path *string, timeout time.Duration, args ...string) (Result, error) {
	return invokeInternal(path, timeout, args...)
}

// region Private Methods

func invokeInternal(path *string, timeout time.Duration, args ...string) (Result, error) {
	log.Debugf("Executing (timeout=%s) `%s %s`", timeout, filepath.Base(*path), strings.Join(args, " "))

	var cmd *exec.Cmd = nil

	if timeout > 0 {
		var ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
		cmd = exec.CommandContext(ctx, *path, args...)
	} else {
		cmd = exec.Command(*path, args...)
	}

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	err := cmd.Run()
	exitCode := cmd.ProcessState.ExitCode()

	var result = Result{
		ExitCode: exitCode,
		Stdout:   outb.Bytes(),
		Stderr:   errb.Bytes(),
	}
	return result, err
}

type TransportCommand struct {
	path    *string
	command *string
	args    []string
}

type ProcessBuilder struct {
	serial  *types.Serial
	timeout time.Duration
	command *TransportCommand
	verbose bool
	stdout  *os.File
}

func NewProcessBuilder(t types.Serial) *ProcessBuilder {
	b := new(ProcessBuilder)
	b.serial = &t
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

func (p *ProcessBuilder) Args(args ...string) {
	p.command.args = append(p.command.args, args...)
}

func (p *ProcessBuilder) Verbose(value bool) {
	p.verbose = value
}

func (p *ProcessBuilder) Stdout(value *os.File) {
	p.stdout = value
}

func (p *ProcessBuilder) Timeout(time time.Duration) {
	p.timeout = time
}

func (p *ProcessBuilder) Path(path *string) {
	p.command.path = path
}

func (p *ProcessBuilder) Command(command string) {
	p.command.command = &command
}

func (p *ProcessBuilder) Invoke() (Result, error) {
	if p.verbose {
		log.Debugf(repr.String(p.command))
	}

	var adb = filepath.Base(*p.command.path)
	var finalArgs []string

	if p.serial != nil {
		var p3 = *p.serial
		finalArgs = append(finalArgs, "-s", p3.GetSerialAddress())
	}

	finalArgs = append(finalArgs, *p.command.command)
	finalArgs = append(finalArgs, p.command.args...)

	log.Debugf("Executing (timeout=%s) `%s %s`", p.timeout.String(), adb, strings.Join(finalArgs, " "))

	var cmd *exec.Cmd = nil

	if p.timeout > 0 {
		var ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
		defer cancel()
		cmd = exec.CommandContext(ctx, *p.command.path, finalArgs...)
	} else {
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

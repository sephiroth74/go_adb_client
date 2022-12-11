package transport

import (
	"it.sephiroth/adbclient/logging"
	"it.sephiroth/adbclient/types"

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

	if err != nil {
		return result, err
	}

	return result, nil
}

type TransportCommand struct {
	path    *string
	command *string
	args    []string
}

type ProcessBuilder[T types.Serial] struct {
	serial  *T
	timeout time.Duration
	command *TransportCommand
}

func NewProcessBuilder[T types.Serial](t T) *ProcessBuilder[T] {
	b := new(ProcessBuilder[T])
	b.serial = &t
	b.timeout = 0
	b.command = &TransportCommand{
		path:    nil,
		command: nil,
		args:    []string{},
	}
	return b
}

func (p ProcessBuilder[T]) Args(args ...string) *ProcessBuilder[T] {
	p.command.args = append(p.command.args, args...)
	return &p
}

func (p ProcessBuilder[T]) Timeout(time time.Duration) *ProcessBuilder[T] {
	p.timeout = time
	return &p
}

func (p ProcessBuilder[T]) Path(path *string) *ProcessBuilder[T] {
	p.command.path = path
	return &p
}

func (p ProcessBuilder[T]) Command(command string) *ProcessBuilder[T] {
	p.command.command = &command
	return &p
}

func (p ProcessBuilder[T]) Invoke() (Result, error) {
	repr.Println(p)

	var adb = filepath.Base(*p.command.path)

	final_args := []string{}

	if p.serial != nil {
		var p3 = *p.serial
		final_args = append(final_args, "-s", p3.Serial())
	}

	final_args = append(final_args, *p.command.command)
	final_args = append(final_args, p.command.args...)

	log.Debugf("Executing (timeout=%s) `%s %s`", p.timeout.String(), adb, strings.Join(final_args, " "))

	var cmd *exec.Cmd = nil

	if p.timeout > 0 {
		var ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
		defer cancel()
		cmd = exec.CommandContext(ctx, *p.command.path, final_args...)
	} else {
		cmd = exec.Command(*p.command.path, final_args...)
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

	if err != nil {
		return result, err
	}

	return result, nil

}

package transport

import (
	"github.com/pkg/errors"
	"io"
	"os"
	"syscall"

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

func (r Result) IsInterrupted() bool {
	return r.ExitCode == int(syscall.SIGINT)
}

func (r Result) IsOk() bool {
	return r.ExitCode == 0
}

func (r Result) NewError() error {
	return errors.New(fmt.Sprintf("invalid exit code: %d\n%s", r.ExitCode, r.Error()))
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
	stdout  *io.Writer
	stderr  *io.Writer
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

func (p *ProcessBuilder) WithStdout(value *io.Writer) *ProcessBuilder {
	p.stdout = value
	return p
}

func (p *ProcessBuilder) WithStderr(value *io.Writer) *ProcessBuilder {
	p.stderr = value
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

func (p *ProcessBuilder) start(stdout *bytes.Buffer, stderr *bytes.Buffer) (*exec.Cmd, context.CancelFunc, error) {
	var adb = filepath.Base(*p.command.path)
	var finalArgs []string

	if p.serial != "" {
		finalArgs = append(finalArgs, "-s", p.serial)
	}

	finalArgs = append(finalArgs, *p.command.command)
	finalArgs = append(finalArgs, p.command.args...)

	var cmd *exec.Cmd = nil
	var ctx context.Context
	var cancel context.CancelFunc

	if p.timeout > 0 {
		if p.verbose {
			logging.Log.Debug().Msgf("Executing (timeout=%s) `%s %s`", p.timeout.String(), adb, strings.Join(finalArgs, " "))
		}
		ctx, cancel = context.WithTimeout(context.Background(), p.timeout)
		cmd = exec.CommandContext(ctx, *p.command.path, finalArgs...)
	} else {
		if p.verbose {
			logging.Log.Debug().Msgf("Executing `%s %s`", adb, strings.Join(finalArgs, " "))
		}
		ctx, cancel = context.WithCancel(context.Background())
		cmd = exec.Command(*p.command.path, finalArgs...)
	}

	if p.stdout != nil {
		cmd.Stdout = *p.stdout
	} else if stdout != nil {
		cmd.Stdout = stdout
	}

	if p.stderr != nil {
		cmd.Stderr = *p.stderr
	} else if stderr != nil {
		cmd.Stderr = stderr
	}

	if err := cmd.Start(); err != nil {
		return cmd, cancel, err
	}
	return cmd, cancel, nil
}

func (p *ProcessBuilder) Invoke() (Result, error) {
	var outBuf, errBuf bytes.Buffer
	cmd, cancel, err := p.start(&outBuf, &errBuf)
	defer cancel()

	if err != nil {
		return Result{}, err
	}

	if err := cmd.Wait(); err != nil {
		return Result{}, err
	}

	exitCode := cmd.ProcessState.ExitCode()

	var result = Result{
		ExitCode: exitCode,
		Stdout:   outBuf.Bytes(),
		Stderr:   errBuf.Bytes(),
	}
	if err != nil {
		return result, err
	}
	return result, nil
}

func (p *ProcessBuilder) InvokeWithCancel(closeChannel chan os.Signal) (Result, error) {
	var outBuf, errBuf bytes.Buffer
	cmd, cancel, err := p.start(&outBuf, &errBuf)
	defer cancel()

	if err != nil {
		return Result{}, err
	}

	go func() {
		<-closeChannel
		logging.Log.Warn().Msg("Kill Process!")
		err := cmd.Process.Kill()
		if err != nil {
			return
		}
	}()

	err = cmd.Wait()
	status := cmd.ProcessState.Sys().(syscall.WaitStatus)
	exitStatus := status
	signaled := status.Signaled()
	signal := status.Signal()
	exitCode := cmd.ProcessState.ExitCode()

	if signaled {
		logging.Log.Warn().Msgf("Signal: %s", signal)
		exitCode = int(exitStatus)
	} else {
		exitCode = int(exitStatus)
	}

	var result = Result{
		ExitCode: exitCode,
		Stdout:   outBuf.Bytes(),
		Stderr:   errBuf.Bytes(),
	}

	if err != nil {
		return result, err
	}
	return result, nil
}

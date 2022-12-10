package transport

import (
	"it.sephiroth/adbclient/logging"

	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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

func (r Result) GetOutput() string {
	return strings.TrimSpace(string(r.Stdout))
}

func (r Result) GetError() string {
	return strings.TrimSpace(string(r.Stderr))
}

func (r Result) ToString() string {
	return fmt.Sprintf("Result(isOk=`%t`, Stdout=`%s`, Stderr=`%s`)", r.IsOk(), strings.TrimSpace(string(r.Stdout)), strings.TrimSpace(string(r.Stderr)))
}

func (r Result) String() string {
	return r.ToString()
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

package commander

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type RunOpts struct {
	Commands []string
	UseSudo  bool

	StdOuFifooPath string
	StdOutBuffer   *bytes.Buffer // will be prefered over fifo-path

	StdErrFifoPath string
	StdErrBuffer   *bytes.Buffer // will be prefered over fifo-path
}

func (opts RunOpts) Validate() (err error) {
	if len(opts.Commands) == 0 {
		err = errors.New("missing commands")
	}
	return
}

// Run an cmd in background and detached it
func (cmder *Commander) Run(opts RunOpts) (err error) {
	// Validate
	err = opts.Validate()
	if err != nil {
		return
	}

	// Sudo
	if opts.UseSudo {
		opts.Commands = append([]string{"sudo"}, opts.Commands...)
	}

	cmd := opts.Commands[0]
	args := opts.Commands[1:]

	// prepare
	cmder.command = exec.Command(cmd, args...)

	// stderr
	if opts.StdErrFifoPath != "" {
		var stdErrFifo *os.File
		stdErrFifo, err = os.OpenFile(opts.StdErrFifoPath, os.O_WRONLY /*|syscall.O_NONBLOCK*/, 0600)
		if err != nil {
			err = fmt.Errorf("could not open stderr: %v", err)
			return
		}

		cmder.command.Stderr = stdErrFifo
	}
	if opts.StdErrBuffer != nil {
		cmder.command.Stderr = opts.StdErrBuffer
	}

	// stdout
	if opts.StdOuFifooPath != "" {
		var stdOutInfo *os.File
		stdOutInfo, err = os.OpenFile(opts.StdOuFifooPath, os.O_WRONLY, 0600)
		if err != nil {
			err = fmt.Errorf("could not open stdout: %v", err)
			return
		}

		cmder.command.Stdout = stdOutInfo
	}
	if opts.StdOutBuffer != nil {
		cmder.command.Stdout = opts.StdOutBuffer
	}

	cmder.command.SysProcAttr = &syscall.SysProcAttr{
		Foreground: false,
		Setsid:     true, // new session
	}

	err = cmder.command.Start()
	if err != nil {
		return
	}

	cmder.pid = cmder.command.Process.Pid

	go func() {
		cmder.command.Wait()
		cmder.waitFinished <- true
		close(cmder.waitFinished)
	}()

	return
}

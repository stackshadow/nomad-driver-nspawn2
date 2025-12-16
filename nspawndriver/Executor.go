//go:build linux

package nspawndriver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/hashicorp/nomad/plugins/drivers"
	"golang.org/x/sys/unix"
)

type Commander struct {
	command     *exec.Cmd
	pid         int
	returnValue chan int
}

func (opts ExecBackgroundWithFIFOOpts) Validate() (err error) {
	if len(opts.Commands) == 0 {
		err = errors.New("missing commands")
	}
	return
}

type ExecBackgroundWithFIFOOpts struct {
	Commands []string
	UseSudo  bool

	StdOuFifooPath string
	StdOutBuffer   *bytes.Buffer // will be prefered over fifo-path

	StdErrFifoPath string
	StdErrBuffer   *bytes.Buffer // will be prefered over fifo-path
}

// Run an cmd in background and detached it
func NewCommander(opts ExecBackgroundWithFIFOOpts) (cmder *Commander, err error) {
	// Validate
	err = opts.Validate()
	if err != nil {
		return
	}

	cmder = &Commander{
		returnValue: make(chan int, 1),
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

	if err = cmder.command.Start(); err != nil {
		return
	}

	cmder.pid = cmder.command.Process.Pid

	// This prevent Zombies after kill
	go func() {
		err = cmder.command.Wait()
		state := cmder.command.ProcessState
		cmder.returnValue <- state.ExitCode()
	}() // Reaper-Goroutine

	return
}

func (cmder *Commander) IsAlive() bool {
	// /proc/PID existiert → Prozess läuft
	_, err := os.Stat(fmt.Sprintf("/proc/%d", cmder.pid))
	return err == nil
}

func (cmder *Commander) Stop() (err error) {

	timeTicker := time.NewTicker(time.Millisecond * 200)
	defer timeTicker.Stop()
	timeoutSigterm := time.NewTimer(time.Second * 2)
	defer timeoutSigterm.Stop()

	// Wait for "normal return value"
	select {
	case <-cmder.returnValue:
		if !cmder.IsAlive() {
			return
		}
		break
	case <-timeoutSigterm.C:
		break
	}

	// SIGTERM
	timeoutSigterm.Reset(time.Second * 10)
	unix.Kill(cmder.pid, unix.SIGTERM)
waitForTerm:
	for {
		select {
		case <-timeTicker.C:
			if !cmder.IsAlive() {
				select {
				case <-cmder.returnValue:
					return
				case <-timeoutSigterm.C:
					return
				}
			}
		case <-timeoutSigterm.C:
			break waitForTerm
		}
	}

	// SIGKILL
	timeoutSigterm.Reset(time.Second * 10)
	unix.Kill(cmder.pid, unix.SIGKILL)
waitForKill:
	for {
		select {
		case <-timeTicker.C:
			if !cmder.IsAlive() {
				select {
				case <-cmder.returnValue:
					return
				case <-timeoutSigterm.C:
					return
				}
			}
		case <-timeoutSigterm.C:
			break waitForKill
		}
	}

	return errors.New("timeout on killing process")
}

func (cmder *Commander) Destroy() {
	close(cmder.returnValue)
	cmder.returnValue = nil
}

func ExecStats(ctx context.Context, interval time.Duration) (resourceUsageChannel <-chan *drivers.TaskResourceUsage, err error) {
	resourceUsageChannel = make(<-chan *drivers.TaskResourceUsage)
	// @TODO
	return
}

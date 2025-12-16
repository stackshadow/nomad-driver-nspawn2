//go:build linux

package commander

import (
	"context"
	"os/exec"
	"time"

	"github.com/hashicorp/nomad/plugins/drivers"
)

type Commander struct {
	command         *exec.Cmd
	pid             int
	waitFinished    chan bool
	timeoutDuration time.Duration
}

type NewCommanderOpts struct {
	// Durtation of waiting for commands
	// defaults to 10sec
	TimeoutDuration time.Duration
}

// Run an cmd in background and detached it
func New(opts NewCommanderOpts) (cmder *Commander) {

	if opts.TimeoutDuration == 0 {
		opts.TimeoutDuration = time.Second * 10
	}

	cmder = &Commander{
		waitFinished:    make(chan bool, 1),
		timeoutDuration: opts.TimeoutDuration,
	}
	return
}

func (cmder *Commander) Destroy() {

}

func (cmder *Commander) PID() (pid int) {
	return cmder.pid
}

func ExecStats(ctx context.Context, interval time.Duration) (resourceUsageChannel <-chan *drivers.TaskResourceUsage, err error) {
	resourceUsageChannel = make(<-chan *drivers.TaskResourceUsage)
	// @TODO
	return
}

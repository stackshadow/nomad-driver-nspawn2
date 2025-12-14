// Copyright IBM Corp. 2019, 2025
// SPDX-License-Identifier: MPL-2.0

package nspawndriver

import (
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/nomad/drivers/shared/executor"
	"github.com/hashicorp/nomad/plugins/drivers"
)

// taskState should store all relevant runtime information
// such as process ID if this is a local task or other meta
// data if this driver deals with external APIs
type taskState struct {
	// stateLock syncs access to all fields below
	stateLock sync.RWMutex

	logger hclog.Logger

	taskConfig  *drivers.TaskConfig
	procState   drivers.TaskState
	startedAt   time.Time
	completedAt time.Time
	exitResult  *drivers.ExitResult

	exec         executor.Executor
	pluginClient *plugin.Client
	pid          int

	machineName string // the name of the systemd-machine
}

func (h *taskState) TaskStatus() *drivers.TaskStatus {
	h.stateLock.RLock()
	defer h.stateLock.RUnlock()

	return &drivers.TaskStatus{
		ID:          h.taskConfig.ID,
		Name:        h.taskConfig.Name,
		State:       h.procState,
		StartedAt:   h.startedAt,
		CompletedAt: h.completedAt,
		ExitResult:  h.exitResult,
		DriverAttributes: map[string]string{
			"pid": strconv.Itoa(h.pid),
		},
	}
}

func (h *taskState) IsRunning() bool {
	h.stateLock.RLock()
	defer h.stateLock.RUnlock()
	return h.procState == drivers.TaskStateRunning
}

// func (h *taskState) run() {
// 	h.stateLock.Lock()
// 	if h.exitResult == nil {
// 		h.exitResult = &drivers.ExitResult{}
// 	}
// 	h.stateLock.Unlock()

// 	// TODO: wait for your task to complete and upate its state.
// 	ps, err := h.exec.Wait(context.Background())
// 	h.stateLock.Lock()
// 	defer h.stateLock.Unlock()

// 	if err != nil {
// 		h.exitResult.Err = err
// 		h.procState = drivers.TaskStateUnknown
// 		h.completedAt = time.Now()
// 		return
// 	}
// 	h.procState = drivers.TaskStateExited
// 	h.exitResult.ExitCode = ps.ExitCode
// 	h.exitResult.Signal = ps.Signal
// 	h.completedAt = ps.Time
// }

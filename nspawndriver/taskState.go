// Copyright IBM Corp. 2019, 2025
// SPDX-License-Identifier: MPL-2.0

package nspawndriver

import (
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/stackshadow/nspawn2/pkg/Stats/service"
)

// taskState should store all relevant runtime information
// such as process ID if this is a local task or other meta
// data if this driver deals with external APIs
type taskState struct {
	// stateLock syncs access to all fields below
	stateLock sync.RWMutex

	logger     hclog.Logger
	taskConfig *drivers.TaskConfig
	procState  drivers.TaskState

	machineName string // the name of the systemd-machine
	statService *service.Service

	// state
	startedAt   time.Time
	completedAt time.Time
	exitResult  *drivers.ExitResult
}

func (h *taskState) TaskStatus() *drivers.TaskStatus {
	h.stateLock.RLock()
	defer h.stateLock.RUnlock()

	return &drivers.TaskStatus{
		ID:               h.taskConfig.ID,
		Name:             h.taskConfig.Name,
		State:            h.procState,
		StartedAt:        h.startedAt,
		CompletedAt:      h.completedAt,
		ExitResult:       h.exitResult,
		DriverAttributes: map[string]string{},
	}
}

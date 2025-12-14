// Copyright IBM Corp. 2019, 2025
// SPDX-License-Identifier: MPL-2.0

package nspawndriver

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/consul-template/signals"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/drivers/shared/eventer"
	"github.com/hashicorp/nomad/drivers/shared/executor"
	"github.com/hashicorp/nomad/helper/uuid"
	"github.com/hashicorp/nomad/plugins/base"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/hashicorp/nomad/plugins/shared/hclspec"
	"github.com/hashicorp/nomad/plugins/shared/structs"
)

const (
	// pluginName is the name of the plugin
	// this is used for logging and (along with the version) for uniquely
	// identifying plugin binaries fingerprinted by the client
	pluginName = "nspawn2"

	// pluginVersion allows the client to identify and use newer versions of
	// an installed plugin
	pluginVersion = "v0.1.0"

	// fingerprintPeriod is the interval at which the plugin will send
	// fingerprint responses
	fingerprintPeriod = 30 * time.Second

	// taskHandleVersion is the version of task handle which this plugin sets
	// and understands how to decode
	// this is used to allow modification and migration of the task schema
	// used by the plugin
	taskHandleVersion = 1
)

var (
	// pluginInfo describes the plugin
	pluginInfo = &base.PluginInfoResponse{
		Type:              base.PluginTypeDriver,
		PluginApiVersions: []string{drivers.ApiVersion010},
		PluginVersion:     pluginVersion,
		Name:              pluginName,
	}

	// capabilities indicates what optional features this driver supports
	// this should be set according to the target run time.
	capabilities = &drivers.Capabilities{
		// TODO: set plugin's capabilities
		//
		// The plugin's capabilities signal Nomad which extra functionalities
		// are supported. For a list of available options check the docs page:
		// https://godoc.org/github.com/hashicorp/nomad/plugins/drivers#Capabilities
		SendSignals: true,
		Exec:        false,

		// we support mounts via bind
		MountConfigs: drivers.MountConfigSupportAll,
	}
)

// NSpawnDriverPlugin is an example driver plugin. When provisioned in a job,
// the taks will output a greet specified by the user.
type NSpawnDriverPlugin struct {
	// eventer is used to handle multiplexing of TaskEvents calls such that an
	// event can be broadcast to all callers
	eventer *eventer.Eventer

	// config is the plugin configuration set by the SetConfig RPC
	config *DriverConfig

	// nomadConfig is the client config from Nomad
	nomadConfig *base.ClientDriverConfig

	// tasks is the in memory datastore mapping taskIDs to driver handles
	tasks *taskStates

	// ctx is the context for the driver. It is passed to other subsystems to
	// coordinate shutdown
	ctx context.Context

	// signalShutdown is called when the driver is shutting down and cancels
	// the ctx passed to any subsystems
	signalShutdown context.CancelFunc

	// logger will log to the Nomad agent
	logger hclog.Logger
}

// NewPlugin returns a new example driver plugin
func NewPlugin(logger hclog.Logger) drivers.DriverPlugin {
	ctx, cancel := context.WithCancel(context.Background())
	logger = logger.Named(pluginName)

	return &NSpawnDriverPlugin{
		eventer:        eventer.NewEventer(ctx, logger),
		config:         &DriverConfig{},
		tasks:          newTaskStore(),
		ctx:            ctx,
		signalShutdown: cancel,
		logger:         logger,
	}
}

// PluginInfo returns information describing the plugin.
func (d *NSpawnDriverPlugin) PluginInfo() (*base.PluginInfoResponse, error) {
	return pluginInfo, nil
}

// TaskConfigSchema returns the HCL schema for the configuration of a task.
func (d *NSpawnDriverPlugin) TaskConfigSchema() (*hclspec.Spec, error) {
	return TaskConfigSpec, nil
}

// Capabilities returns the features supported by the driver.
func (d *NSpawnDriverPlugin) Capabilities() (*drivers.Capabilities, error) {
	return capabilities, nil
}

// StartTask returns a task handle and a driver network if necessary.
func (d *NSpawnDriverPlugin) StartTask(cfg *drivers.TaskConfig) (retTaskHandle *drivers.TaskHandle, retDriverNetwork *drivers.DriverNetwork, err error) {
	if _, ok := d.tasks.Get(cfg.ID); ok {
		return nil, nil, fmt.Errorf("task with ID %q already started", cfg.ID)
	}

	var driverConfig TaskConfig
	err = cfg.DecodeDriverConfig(&driverConfig)
	if err != nil {
		return
	}

	err = driverConfig.Validate()
	if err != nil {
		return
	}

	// update machine name
	if driverConfig.MachineName != "" {
		driverConfig.MachineName = driverConfig.MachineName + "-" + uuid.Generate()
	} else {
		driverConfig.MachineName = uuid.Generate()
	}

	d.logger.Info("info task", "driverConfig", hclog.Fmt("%+v", driverConfig))
	d.logger.Info("info task", "mounts", hclog.Fmt("%+v", cfg.Mounts))
	d.logger.Info("info task", "env", hclog.Fmt("%+v", cfg.Env))

	retTaskHandle = drivers.NewTaskHandle(taskHandleVersion)
	retTaskHandle.Config = cfg

	h := &taskState{
		taskConfig:  cfg,
		procState:   drivers.TaskStateRunning,
		startedAt:   time.Now().Round(time.Millisecond),
		logger:      d.logger,
		machineName: driverConfig.MachineName,
	}

	retDriverNetwork, err = d.StartContainer(StartContainerOpts{
		taskConfig:       cfg,
		driverTaskConfig: driverConfig,
		handle:           h,
	})
	if err != nil {
		return
	}

	driverState := DriverState{
		ReattachConfig: structs.ReattachConfigFromGoPlugin(h.pluginClient.ReattachConfig()),
		Pid:            h.pid,
		TaskConfig:     cfg,
		StartedAt:      h.startedAt,
	}

	err = retTaskHandle.SetDriverState(&driverState)
	if err != nil {
		return
	}

	d.tasks.Set(cfg.ID, h)

	return
}

// RecoverTask recreates the in-memory state of a task from a TaskHandle.
func (d *NSpawnDriverPlugin) RecoverTask(handle *drivers.TaskHandle) error {
	if handle == nil {
		return errors.New("error: handle cannot be nil")
	}

	if _, ok := d.tasks.Get(handle.Config.ID); ok {
		return nil
	}

	var driverState DriverState
	if err := handle.GetDriverState(&driverState); err != nil {
		return fmt.Errorf("failed to decode task state from handle: %v", err)
	}

	var driverConfig TaskConfig
	if err := driverState.TaskConfig.DecodeDriverConfig(&driverConfig); err != nil {
		return fmt.Errorf("failed to decode driver config: %v", err)
	}

	// TODO: implement driver specific logic to recover a task.
	//
	// Recovering a task involves recreating and storing a taskState as if the
	// task was just started.
	//
	// In the example below we use the executor to re-attach to the process
	// that was created when the task first started.
	plugRC, err := structs.ReattachConfigToGoPlugin(driverState.ReattachConfig)
	if err != nil {
		return fmt.Errorf("failed to build ReattachConfig from taskConfig state: %v", err)
	}

	execImpl, pluginClient, err := executor.ReattachToExecutor(plugRC, d.logger, d.nomadConfig.Topology.Compute())
	if err != nil {
		return fmt.Errorf("failed to reattach to executor: %v", err)
	}

	h := &taskState{
		exec:         execImpl,
		pid:          driverState.Pid,
		pluginClient: pluginClient,
		taskConfig:   driverState.TaskConfig,
		procState:    drivers.TaskStateRunning,
		startedAt:    driverState.StartedAt,
		exitResult:   &drivers.ExitResult{},
	}

	d.tasks.Set(driverState.TaskConfig.ID, h)

	//go h.run()
	return nil
}

// WaitTask returns a channel used to notify Nomad when a task exits.
func (d *NSpawnDriverPlugin) WaitTask(ctx context.Context, taskID string) (<-chan *drivers.ExitResult, error) {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return nil, drivers.ErrTaskNotFound
	}

	ch := make(chan *drivers.ExitResult)
	go d.handleWait(ctx, handle, ch)
	return ch, nil
}

func (d *NSpawnDriverPlugin) handleWait(ctx context.Context, handle *taskState, ch chan *drivers.ExitResult) {
	defer close(ch)
	var result *drivers.ExitResult

	// TODO: implement driver specific logic to notify Nomad the task has been
	// completed and what was the exit result.
	//
	// When a result is sent in the result channel Nomad will stop the task and
	// emit an event that an operator can use to get an insight on why the task
	// stopped.
	//
	// In the example below we block and wait until the executor finishes
	// running, at which point we send the exit code and signal in the result
	// channel.
	ps, err := handle.exec.Wait(ctx)
	if err != nil {
		result = &drivers.ExitResult{
			Err: fmt.Errorf("executor: error waiting on process: %v", err),
		}
	} else {
		result = &drivers.ExitResult{
			ExitCode: ps.ExitCode,
			Signal:   ps.Signal,
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.ctx.Done():
			return
		case ch <- result:
		}
	}
}

// StopTask stops a running task with the given signal and within the timeout window.
func (d *NSpawnDriverPlugin) StopTask(taskID string, timeout time.Duration, signal string) error {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return drivers.ErrTaskNotFound
	}

	// TODO: implement driver specific logic to stop a task.
	//
	// The StopTask function is expected to stop a running task by sending the
	// given signal to it. If the task does not stop during the given timeout,
	// the driver must forcefully kill the task.
	//
	// In the example below we let the executor handle the task shutdown
	// process for us, but you might need to customize this for your own
	// implementation.
	if err := handle.exec.Shutdown(signal, timeout); err != nil {
		if handle.pluginClient.Exited() {
			return nil
		}
		return fmt.Errorf("executor Shutdown failed: %v", err)
	}

	return nil
}

// DestroyTask cleans up and removes a task that has terminated.
func (d *NSpawnDriverPlugin) DestroyTask(taskID string, force bool) error {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return drivers.ErrTaskNotFound
	}

	if handle.IsRunning() && !force {
		return errors.New("cannot destroy running task")
	}

	// TODO: implement driver specific logic to destroy a complete task.
	//
	// Destroying a task includes removing any resources used by task and any
	// local references in the plugin. If force is set to true the task should
	// be destroyed even if it's currently running.
	//
	// In the example below we use the executor to force shutdown the task
	// (timeout equals 0).
	if !handle.pluginClient.Exited() {
		if err := handle.exec.Shutdown("", 0); err != nil {
			handle.logger.Error("destroying executor failed", "err", err)
		}

		handle.pluginClient.Kill()
	}

	d.tasks.Delete(taskID)
	return nil
}

// InspectTask returns detailed status information for the referenced taskID.
func (d *NSpawnDriverPlugin) InspectTask(taskID string) (*drivers.TaskStatus, error) {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return nil, drivers.ErrTaskNotFound
	}

	return handle.TaskStatus(), nil
}

// TaskStats returns a channel which the driver should send stats to at the given interval.
func (d *NSpawnDriverPlugin) TaskStats(ctx context.Context, taskID string, interval time.Duration) (<-chan *drivers.TaskResourceUsage, error) {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return nil, drivers.ErrTaskNotFound
	}

	// TODO: implement driver specific logic to send task stats.
	//
	// This function returns a channel that Nomad will use to listen for task
	// stats (e.g., CPU and memory usage) in a given interval. It should send
	// stats until the context is canceled or the task stops running.
	//
	// In the example below we use the Stats function provided by the executor,
	// but you can build a set of functions similar to the fingerprint process.
	return handle.exec.Stats(ctx, interval)
}

// TaskEvents returns a channel that the plugin can use to emit task related events.
func (d *NSpawnDriverPlugin) TaskEvents(ctx context.Context) (<-chan *drivers.TaskEvent, error) {
	return d.eventer.TaskEvents(ctx)
}

// SignalTask forwards a signal to a task.
// This is an optional capability.
func (d *NSpawnDriverPlugin) SignalTask(taskID string, signal string) error {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return drivers.ErrTaskNotFound
	}

	// TODO: implement driver specific signal handling logic.
	//
	// The given signal must be forwarded to the target taskID. If this plugin
	// doesn't support receiving signals (capability SendSignals is set to
	// false) you can just return nil.
	sig := os.Interrupt
	if s, ok := signals.SignalLookup[signal]; ok {
		sig = s
	} else {
		d.logger.Warn("unknown signal to send to task, using SIGINT instead", "signal", signal, "task_id", handle.taskConfig.ID)

	}
	return handle.exec.Signal(sig)
}

// ExecTask returns the result of executing the given command inside a task.
// This is an optional capability.
func (d *NSpawnDriverPlugin) ExecTask(taskID string, cmd []string, timeout time.Duration) (*drivers.ExecTaskResult, error) {
	// TODO: implement driver specific logic to execute commands in a task.
	return nil, errors.New("this driver does not support exec yet")
}

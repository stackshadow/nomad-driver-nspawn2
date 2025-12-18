// Copyright IBM Corp. 2019, 2025
// SPDX-License-Identifier: MPL-2.0

package nspawndriver

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/drivers/shared/eventer"
	"github.com/hashicorp/nomad/helper/uuid"
	"github.com/hashicorp/nomad/plugins/base"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/hashicorp/nomad/plugins/shared/hclspec"
	"github.com/hashicorp/nomad/plugins/shared/structs"
	machineManagerDomain "github.com/stackshadow/nspawn2/pkg/MachineManager/domain"
	machineManager "github.com/stackshadow/nspawn2/pkg/MachineManager/service"
	nspawnStats "github.com/stackshadow/nspawn2/pkg/Stats"
	statDomain "github.com/stackshadow/nspawn2/pkg/Stats/domain"
	"github.com/stackshadow/nspawn2/pkg/Stats/service"
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

	// the manager who takes care about systemd-machines
	manager *machineManager.Service

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

	// d.logger.Info("info task", "driverConfig", hclog.Fmt("%+v", driverConfig))
	d.logger.Info("info task", "mounts", hclog.Fmt("%+v", cfg.Mounts))
	// d.logger.Info("info task", "env", hclog.Fmt("%+v", cfg.Env))

	retTaskHandle = drivers.NewTaskHandle(taskHandleVersion)
	retTaskHandle.Config = cfg

	h := &taskState{
		taskConfig:  cfg,
		procState:   drivers.TaskStateRunning,
		startedAt:   time.Now().Round(time.Millisecond),
		logger:      d.logger,
		machineName: driverConfig.MachineName,
	}

	ipv4, ipv6, err := d.manager.MachineStart(machineManagerDomain.StartOpts{
		MachineName: driverConfig.MachineName,
		Hostname:    driverConfig.Hostname,

		ImageFileName: driverConfig.Image,
		IsSystemd:     driverConfig.Boot,
		Ephemeral:     driverConfig.Ephemeral,

		Environment: driverConfig.Environment,

		Bind:         driverConfig.Bind,
		BindReadOnly: driverConfig.BindReadOnly,

		Networking: machineManagerDomain.StartNeworkingOpts{
			Veth:       driverConfig.NetworkVeth,
			VethName:   driverConfig.NetworkVethExtra,
			BridgeName: driverConfig.NetworkBridge,
		},

		StdOutPath: cfg.StdoutPath,
		StdErrPath: cfg.StderrPath,
	})
	if err != nil {
		return
	}

	if ipv6 != "" || ipv4 != "" {
		retDriverNetwork = &drivers.DriverNetwork{}
		if ipv4 != "" {
			d.logger.Info("start task found ipv4", "ipv4", ipv4)
			retDriverNetwork.IP = ipv4
		}
	}

	driverState := DriverState{
		ReattachConfig: &structs.ReattachConfig{},

		TaskConfig:  cfg,
		StartedAt:   h.startedAt,
		MachineName: driverConfig.MachineName,
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

	// // TODO: implement driver specific logic to recover a task.
	// //
	// // Recovering a task involves recreating and storing a taskState as if the
	// // task was just started.
	// //
	// // In the example below we use the executor to re-attach to the process
	// // that was created when the task first started.
	plugRC, err := structs.ReattachConfigToGoPlugin(driverState.ReattachConfig)
	if err != nil {
		return fmt.Errorf("failed to build ReattachConfig from taskConfig state: %v", err)
	}
	_ = plugRC

	h := &taskState{
		logger: d.logger,

		taskConfig: driverState.TaskConfig,
		procState:  drivers.TaskStateRunning,

		machineName: driverState.MachineName,

		startedAt:  driverState.StartedAt,
		exitResult: &drivers.ExitResult{},
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

	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.ctx.Done():
			return
		case <-ticker.C:

			state, err := d.manager.State(handle.machineName)
			if err != nil {
				ch <- &drivers.ExitResult{
					ExitCode: -1,
					Err:      err,
				}
			}
			if state == machineManagerDomain.MachineStateNotExist {
				ch <- &drivers.ExitResult{
					ExitCode: 0,
					Err:      nil,
				}
			}
			if state == machineManagerDomain.MachineStateStopped {
				ch <- &drivers.ExitResult{
					ExitCode: 0,
					Err:      nil,
				}
			}
		}
	}
}

// StopTask stops a running task with the given signal and within the timeout window.
func (d *NSpawnDriverPlugin) StopTask(taskID string, timeout time.Duration, signal string) error {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return drivers.ErrTaskNotFound
	}

	if handle.statService != nil {
		d.logger.Info("stoptask - stop metrics", "machine_name", handle.machineName)
		handle.statService.Destroy()
		handle.statService = nil
	}

	d.logger.Info("stoptask - request", "machine_name", handle.machineName)
	d.manager.Stop(handle.machineName)

	d.logger.Info("stoptask - wait for stopping", "machine_name", handle.machineName)
	err := d.manager.WaitForStopping(handle.machineName)
	if err != nil {
		d.logger.Error("stop task", err)
		err = fmt.Errorf("stop task: %w", err)
		return err
	}

	d.logger.Info("stoptask - stopped", "machine_name", handle.machineName)

	return nil
}

// DestroyTask cleans up and removes a task that has terminated.
func (d *NSpawnDriverPlugin) DestroyTask(taskID string, force bool) error {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return drivers.ErrTaskNotFound
	}
	_ = handle

	d.tasks.Delete(taskID)
	return nil
}

// InspectTask returns detailed status information for the referenced taskID.
func (d *NSpawnDriverPlugin) InspectTask(taskID string) (*drivers.TaskStatus, error) {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return nil, drivers.ErrTaskNotFound
	}

	d.logger.Debug("Inspect task...")

	return handle.TaskStatus(), nil
}

// TaskStats returns a channel which the driver should send stats to at the given interval.
func (d *NSpawnDriverPlugin) TaskStats(ctx context.Context, taskID string, interval time.Duration) (<-chan *drivers.TaskResourceUsage, error) {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return nil, drivers.ErrTaskNotFound
	}
	_ = handle

	resourceUsage := make(chan *drivers.TaskResourceUsage)
	if handle.statService == nil {
		d.logger.Debug("stat-service - start with interval", "interval", interval.String())
		handle.statService = nspawnStats.NewService(
			service.SetMachineName(handle.machineName),
			service.SetRefreshTime(interval),
		)
		go handle.statService.Watch(func(stat statDomain.Stat) {
			resourceUsage <- &drivers.TaskResourceUsage{
				ResourceUsage: &drivers.ResourceUsage{
					MemoryStats: &drivers.MemoryStats{
						// RSS: stat.Memory.,
						Usage: stat.RamUsaged,
					},
					CpuStats: &drivers.CpuStats{
						TotalTicks: stat.CPUTicks,
						Percent:    stat.CPUUsagePercent,
					},
				},
				Timestamp: time.Now().UnixNano(),
			}
		})
	} else {
		d.logger.Debug("stat-service - already running", "interval", interval.String())

	}

	return resourceUsage, nil
}

// TaskEvents returns a channel that the plugin can use to emit task related events.
func (d *NSpawnDriverPlugin) TaskEvents(ctx context.Context) (<-chan *drivers.TaskEvent, error) {
	return d.eventer.TaskEvents(ctx)
}

// SignalTask forwards a signal to a task.
// This is an optional capability.
func (d *NSpawnDriverPlugin) SignalTask(taskID string, signal string) error {
	_, ok := d.tasks.Get(taskID)
	if !ok {
		return drivers.ErrTaskNotFound
	}

	// TODO: implement driver specific signal handling logic.
	//
	// The given signal must be forwarded to the target taskID. If this plugin
	// doesn't support receiving signals (capability SendSignals is set to
	// false) you can just return nil.
	// sig := os.Interrupt
	// if s, ok := signals.SignalLookup[signal]; ok {
	// 	sig = s
	// } else {
	// 	d.logger.Warn("unknown signal to send to task, using SIGINT instead", "signal", signal, "task_id", handle.taskConfig.ID)

	// }
	return nil
}

// ExecTask returns the result of executing the given command inside a task.
// This is an optional capability.
func (d *NSpawnDriverPlugin) ExecTask(taskID string, cmd []string, timeout time.Duration) (*drivers.ExecTaskResult, error) {
	// TODO: implement driver specific logic to execute commands in a task.
	return nil, errors.New("this driver does not support exec yet")
}

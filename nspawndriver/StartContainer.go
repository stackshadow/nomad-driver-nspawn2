package nspawndriver

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/drivers/shared/executor"
	"github.com/hashicorp/nomad/plugins/drivers"
)

type StartContainerOpts struct {
	taskConfig       *drivers.TaskConfig
	driverTaskConfig TaskConfig
	handle           *taskState
}

func (d *NSpawnDriverPlugin) StartContainer(opts StartContainerOpts) (driverNetwork *drivers.DriverNetwork, err error) {

	executorConfig := &executor.ExecutorConfig{
		LogFile:  filepath.Join(opts.taskConfig.TaskDir().Dir, "executor.out"),
		LogLevel: "debug",
	}

	exec, pluginClient, err := executor.CreateExecutor(d.logger, d.nomadConfig, executorConfig)
	if err != nil {
		return
	}

	systemdParameter, err := opts.driverTaskConfig.ToCLIParameter()
	if err != nil {
		return
	}

	cmd := d.config.NSPawnPath
	args := systemdParameter

	if d.config.Sudo {
		cmd = "bash"
		args = []string{"-c"}
		args = append(args, "sudo "+d.config.NSPawnPath+" "+strings.Join(systemdParameter, " "))
	}
	d.logger.Info("exec systemd-nspawn", "args", hclog.Fmt("%s %+v", cmd, args))

	execCmd := &executor.ExecCommand{
		Cmd:        cmd,
		Args:       args,
		StdoutPath: opts.taskConfig.StdoutPath,
		StderrPath: opts.taskConfig.StderrPath,
	}

	ps, err := exec.Launch(execCmd)
	defer func() {
		if err != nil {
			pluginClient.Kill()
		}
	}()
	if err != nil {
		return
	}

	// wait for ready
	_, err = WaitForMachineRunning(opts.driverTaskConfig.MachineName, time.Second*30)
	defer func() {
		if err != nil {
			exec.Shutdown("", time.Second*30)
		}
	}()
	if err != nil {
		return
	}

	// wait for veth
	if opts.driverTaskConfig.NetworkVethExtra != "" {
		d.logger.Info("wait for interface", "interface", opts.driverTaskConfig.NetworkVethExtra)
		err = NetworkWaitForInterface(opts.driverTaskConfig.NetworkVethExtra, time.Second*15, time.Second)
		if err != nil {
			return
		}

		d.logger.Info("try to bring up interface", "interface", opts.driverTaskConfig.NetworkVethExtra)
		err = d.NetworkInterfaceUp(opts.driverTaskConfig.NetworkVethExtra)
		if err != nil {
			return
		}
	}

	// wait for IP-Adress

	if opts.driverTaskConfig.NetworkVeth || opts.driverTaskConfig.NetworkVethExtra != "" ||
		opts.driverTaskConfig.NetworkBridge != "" {

		driverNetwork = &drivers.DriverNetwork{}

		driverNetwork.IP, err = WaitForMachineIPv4(opts.driverTaskConfig.MachineName, time.Second*15)
		if err != nil {
			return
		}
	}

	opts.handle.exec = exec
	opts.handle.pluginClient = pluginClient
	opts.handle.pid = ps.Pid

	return
}

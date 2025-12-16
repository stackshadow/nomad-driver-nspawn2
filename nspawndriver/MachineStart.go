package nspawndriver

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/drivers"
)

type MachineStartOpts struct {
	taskConfig       *drivers.TaskConfig
	driverTaskConfig TaskConfig
	handle           *taskState
}

func (d *NSpawnDriverPlugin) MachineStart(opts MachineStartOpts) (driverNetwork *drivers.DriverNetwork, err error) {

	systemdParameter, err := opts.driverTaskConfig.ToCLIParameter()
	if err != nil {
		return
	}

	cmds := []string{d.config.NSPawnPath}
	cmds = append(cmds, systemdParameter...)

	d.logger.Info("exec systemd-nspawn", "cmds", hclog.Fmt("%+v", cmds))

	var cmder *Commander
	cmder, err = NewCommander(ExecBackgroundWithFIFOOpts{
		Commands:       cmds,
		StdErrFifoPath: opts.taskConfig.StderrPath,
		StdOuFifooPath: opts.taskConfig.StdoutPath,
		UseSudo:        d.config.Sudo,
	})
	defer cmder.Destroy()

	if err != nil {
		err = fmt.Errorf("error on exec: %v", err)
		return
	}
	// cmder.Stop()

	// wait for ready
	_, err = MachineWaitForRunning(opts.driverTaskConfig.MachineName, time.Second*30)
	if err != nil {
		err = fmt.Errorf("error on wait for running: %v", err)
		return
	}

	// wait for veth
	if opts.driverTaskConfig.NetworkVethExtra != "" {
		d.logger.Info("wait for interface", "interface", opts.driverTaskConfig.NetworkVethExtra)
		err = NetworkWaitForInterface(opts.driverTaskConfig.NetworkVethExtra, time.Second*15, time.Second)
		if err != nil {
			err = fmt.Errorf("error on wait for interface: %v", err)
			return
		}

		d.logger.Info("try to bring up interface", "interface", opts.driverTaskConfig.NetworkVethExtra)
		err = d.NetworkInterfaceUp(opts.driverTaskConfig.NetworkVethExtra)
		if err != nil {
			err = fmt.Errorf("error on wait for interface up: %v", err)
			return
		}
	}

	// wait for IP-Adress
	if opts.driverTaskConfig.NetworkVeth || opts.driverTaskConfig.NetworkVethExtra != "" ||
		opts.driverTaskConfig.NetworkBridge != "" {

		driverNetwork = &drivers.DriverNetwork{}

		driverNetwork.IP, err = MachineWaitForIPv4(opts.driverTaskConfig.MachineName, time.Second*15)
		if err != nil {
			err = fmt.Errorf("error on wait for ipv4: %v", err)
			return
		}
	}

	return
}

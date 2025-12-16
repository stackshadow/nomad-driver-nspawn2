package nspawndriver

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/nomad/plugins/drivers"
)

type MachineStopOpts struct {
	handle *taskState
}

func (d *NSpawnDriverPlugin) MachineStop(opts MachineStopOpts) (driverNetwork *drivers.DriverNetwork, err error) {

	var stdout, stderr bytes.Buffer
	cmder, err := NewCommander(ExecBackgroundWithFIFOOpts{
		Commands:     []string{d.config.MachineCtl, "poweroff", opts.handle.machineName},
		UseSudo:      d.config.Sudo,
		StdOutBuffer: &stdout,
		StdErrBuffer: &stderr,
	})
	defer cmder.Destroy()
	if err != nil {
		err = fmt.Errorf("error call pweroff machine '%s': %w", opts.handle.machineName, err)
		return
	}
	cmder.Stop()

	return
}

package clibased

import (
	"fmt"

	commander "github.com/stackshadow/nspawn2/pkg/Commander"
	"github.com/stackshadow/nspawn2/pkg/MachineManager/domain"
)

func convertOptsToParameter(opts domain.StartOpts) (args []string, err error) {
	if opts.ImageFileName == "" {
		return nil, fmt.Errorf("image-path is needed")
	}

	args = []string{"-i", opts.ImageFileName}

	if opts.MachineName != "" {
		args = append(args, "--machine", opts.MachineName)
	}
	if opts.Hostname != "" {
		args = append(args, "--hostname", opts.Hostname)
	}

	if opts.IsSystemd {
		args = append(args, "--boot")
	}
	if opts.Ephemeral {
		args = append(args, "--ephemeral")
	}

	if opts.CopyResolvConf {
		args = append(args, "--resolv-conf", "copy")
	}

	for k, v := range opts.Environment {
		args = append(args, "-E", k+"="+v)
	}
	for k, v := range opts.Bind {
		args = append(args, "--bind", k+":"+v)
	}
	for k, v := range opts.BindReadOnly {
		args = append(args, "--bind-ro", k+":"+v)
	}

	// networking
	if opts.Networking.Veth && opts.Networking.VethName == "" {
		args = append(args, "--network-veth")
	}
	if opts.Networking.VethName != "" {
		args = append(args, "--network-veth-extra", opts.Networking.VethName)
	}
	if opts.Networking.BridgeName != "" {
		args = append(args, "--network-bridge", opts.Networking.BridgeName)
	}

	return
}

// Start implements ports.MachineCommands.
func (a *adapterData) Start(opts domain.StartOpts) (err error) {

	var args []string
	args, err = convertOptsToParameter(opts)
	if err != nil {
		return
	}

	cmds := []string{a.binNSPawnPath}
	cmds = append(cmds, args...)

	cmder := commander.New(commander.NewCommanderOpts{})
	err = cmder.Run(commander.RunOpts{
		Commands:       cmds,
		UseSudo:        a.useSudo,
		StdOuFifooPath: opts.StdOutPath,
		StdErrFifoPath: opts.StdErrPath,
	})
	cmder.Destroy()

	return
}

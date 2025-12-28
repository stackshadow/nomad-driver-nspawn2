package service

import (
	"fmt"

	"github.com/stackshadow/nspawn2/pkg/MachineManager/domain"
)

func (svc *Service) MachineStart(opts domain.StartOpts) (ipv4 string, ipv6 string, err error) {

	// start command
	err = svc.repo.Start(opts)
	if err != nil {
		err = fmt.Errorf("start machine %s: %w", opts.MachineName, err)
		return
	}

	// Wait for running
	err = svc.WaitForRunning(opts.MachineName)
	if err != nil {
		err = fmt.Errorf("start machine %s: %w", opts.MachineName, err)
		return
	}

	// wait for network ( if needed )
	if opts.Networking.VethName != "" {
		err = svc.WaitForInterface(opts.Networking.VethName)
		if err != nil {
			err = fmt.Errorf("start machine %s: %w", opts.MachineName, err)
			return
		}

		err = svc.network.Up(opts.Networking.VethName)
		if err != nil {
			err = fmt.Errorf("start machine %s: %w", opts.MachineName, err)
			return
		}
	}

	// wait for IPv4
	if opts.Networking.Veth || opts.Networking.VethName != "" || opts.Networking.BridgeName != "" {
		ipv4, err = svc.WaitForIPv4(opts.MachineName)
		if err != nil {
			err = fmt.Errorf("start machine %s: %w", opts.MachineName, err)
			return
		}
	}

	return
}

package service

import (
	"time"

	machinePorts "github.com/stackshadow/nspawn2/pkg/MachineManager/ports"
	networkPorts "github.com/stackshadow/nspawn2/pkg/Networking/ports"
)

type Service struct {
	repo    machinePorts.MachineCommands
	network networkPorts.NetworkCommands

	timeout time.Duration
}

type NewServiceOpts struct {
	Repo    machinePorts.MachineCommands
	Network networkPorts.NetworkCommands

	Timeout time.Duration
}

func New(opts NewServiceOpts) *Service {
	return &Service{
		repo:    opts.Repo,
		network: opts.Network,
		timeout: opts.Timeout,
	}
}

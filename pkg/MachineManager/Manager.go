package machinemanager

import (
	"github.com/stackshadow/nspawn2/pkg/MachineManager/adapter/clibased"
	"github.com/stackshadow/nspawn2/pkg/MachineManager/ports"
	"github.com/stackshadow/nspawn2/pkg/MachineManager/service"
)

type MachineCommandsAdapter ports.MachineCommands

type NewOpts struct {
	BinNSPawnPath     string
	BinMachinectlPath string
	UseSudo           bool
}

func NewMachineCommandsCliBased(opts NewOpts) MachineCommandsAdapter {
	return clibased.New(opts)
}

func NewService(opts service.NewServiceOpts) *service.Service {
	return service.New(opts)
}

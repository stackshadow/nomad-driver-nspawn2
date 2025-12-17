package ports

import "github.com/stackshadow/nspawn2/pkg/MachineManager/domain"

type MachineCommands interface {
	List() (machines domain.MachineList, err error)
	Start(opts domain.StartOpts) (err error)
	State(machineName string) (state domain.MachineState, err error)
	Stop(machineName string) (err error)
}

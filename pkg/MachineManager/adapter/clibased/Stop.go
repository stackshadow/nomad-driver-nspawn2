package clibased

import (
	"fmt"

	commander "github.com/stackshadow/nspawn2/pkg/Commander"
)

// Stop implements ports.MachineCommands.
func (a *adapterData) Stop(machineName string) (err error) {

	cmder := commander.New(commander.NewCommanderOpts{})
	defer cmder.Destroy()
	defer cmder.Stop()

	err = cmder.Run(commander.RunOpts{
		Commands: []string{a.binMachinectlPath, "poweroff", machineName},
		UseSudo:  a.useSudo,
	})
	if err != nil {
		err = fmt.Errorf("stop machine '%s': %w", machineName, err)
		return
	}
	cmder.Wait()

	return
}

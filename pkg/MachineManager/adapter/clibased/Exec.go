package clibased

import (
	"fmt"

	commander "github.com/stackshadow/nspawn2/pkg/Commander"
)

// Exec implements ports.MachineCommands.
func (a *adapterData) Exec(machineName string, command ...string) (err error) {

	cmdArray := []string{a.binMachinectlPath, "shell", machineName}
	cmdArray = append(cmdArray, command...)

	cmder := commander.New(commander.NewCommanderOpts{})
	defer cmder.Destroy()
	defer cmder.Stop()

	err = cmder.Run(commander.RunOpts{
		Commands: cmdArray,
		UseSudo:  a.useSudo,
	})
	if err != nil {
		err = fmt.Errorf("run command in machine '%s': %w", machineName, err)
		return
	}
	cmder.Wait()

	return
}

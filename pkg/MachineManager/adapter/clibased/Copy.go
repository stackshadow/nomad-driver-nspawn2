package clibased

import (
	"fmt"

	commander "github.com/stackshadow/nspawn2/pkg/Commander"
)

// Exec implements ports.MachineCommands.
func (a *adapterData) Download(machineName string, pathSource, pathTarget string) (err error) {

	cmder := commander.New(commander.NewCommanderOpts{})
	defer cmder.Destroy()
	defer cmder.Stop()

	err = cmder.Run(commander.RunOpts{
		Commands: []string{a.binMachinectlPath, "copy-from", machineName, pathSource, pathTarget},
		UseSudo:  a.useSudo,
	})
	if err != nil {
		err = fmt.Errorf("run command in machine '%s': %w", machineName, err)
		return
	}
	cmder.Wait()

	return

}

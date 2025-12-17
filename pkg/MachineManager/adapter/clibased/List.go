package clibased

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	commander "github.com/stackshadow/nspawn2/pkg/Commander"
	"github.com/stackshadow/nspawn2/pkg/MachineManager/domain"
)

// List implements ports.MachineCommands.
func (a *adapterData) List() (machines domain.MachineList, err error) {
	machines = make(domain.MachineList)

	var stdout, stderr bytes.Buffer

	cmder := commander.New(commander.NewCommanderOpts{})
	defer cmder.Destroy()
	defer cmder.Stop()

	err = cmder.Run(commander.RunOpts{
		Commands:     []string{"machinectl", "-o", "json", "--no-pager", "list"},
		StdOutBuffer: &stdout,
		StdErrBuffer: &stderr,
	})
	if err != nil {
		err = fmt.Errorf("run list of machines: %w", err)
		return
	}

	err = cmder.Wait()
	if err != nil {
		err = fmt.Errorf("wait for list of machines: %w", err)
		return
	}

	var machinesRaw []MachineInfoRaw
	if err = json.Unmarshal(stdout.Bytes(), &machinesRaw); err != nil {
		err = fmt.Errorf("JSON parse failed: %w", err)
		return
	}

	// convert
	for _, machine := range machinesRaw {

		newMachineInfo := domain.MachineInfo{
			OS:      machine.OS,
			Class:   machine.Class,
			Service: machine.Service,
			Version: machine.Version,
		}

		for machineAddress := range strings.SplitSeq(machine.Addresses, "\n") {
			ipAddress := net.ParseIP(machineAddress)
			if ipAddress == nil {
				continue
			}

			if ipAddress.To4() != nil {
				newMachineInfo.Addresses.IPv4 = machineAddress
			} else {
				newMachineInfo.Addresses.IPv6 = machineAddress
			}

		}

		machines[machine.Name] = newMachineInfo
	}

	return
}

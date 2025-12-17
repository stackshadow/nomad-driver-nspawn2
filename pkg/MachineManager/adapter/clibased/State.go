package clibased

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	commander "github.com/stackshadow/nspawn2/pkg/Commander"
	"github.com/stackshadow/nspawn2/pkg/MachineManager/domain"
)

type MachineStateJson struct {
	ID    string `json:"Id"`
	Name  string `json:"Name"`
	State string `json:"State"`
}

// State implements ports.MachineCommands.
func (a *adapterData) State(machineName string) (state domain.MachineState, err error) {

	var stdout, stderr bytes.Buffer

	cmder := commander.New(commander.NewCommanderOpts{})
	defer cmder.Destroy()
	defer cmder.Stop()

	err = cmder.Run(commander.RunOpts{
		Commands:     []string{a.binMachinectlPath, "--all", "show", machineName},
		StdOutBuffer: &stdout,
		StdErrBuffer: &stderr,
	})
	if err != nil {
		err = fmt.Errorf("get state of machine: %w", err)
		return
	}

	err = cmder.Wait()
	if err != nil {
		err = fmt.Errorf("get state of machine: %w", err)
		return
	}

	err = cmder.Stop()
	if err != nil {
		err = fmt.Errorf("get state of machine: %w", err)
		return
	}

	// Parst key=value Format zu map
	propsMap := make(map[string]any)
	lines := strings.SplitSeq(stdout.String(), "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "=") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Unescape systemd-Quotes (\" → ")
		value = strings.ReplaceAll(value, "\\\"", "\"")
		value = strings.ReplaceAll(value, "\\\\", "\\")

		propsMap[key] = value
	}

	if len(propsMap) == 0 {
		return domain.MachineStateNotExist, nil
	}

	var jsonBytes []byte
	jsonBytes, err = json.Marshal(propsMap)
	if err != nil {
		err = fmt.Errorf("get state of machine: %w", err)
		return
	}

	props := MachineStateJson{}
	err = json.Unmarshal(jsonBytes, &props)
	if err != nil {
		err = fmt.Errorf("get state of machine: %w", err)
		return
	}

	state = domain.MachineState(props.State)
	return
}

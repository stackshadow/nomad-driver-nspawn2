package nspawndriver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

type MachineInfo struct {
	OS        string   `json:"os"`
	Class     string   `json:"class"`
	Service   string   `json:"service"`
	Version   string   `json:"version"`
	Addresses []string `json:"Addresses"`
}

type MachineInfoRaw struct {
	Name      string `json:"machine"`
	OS        string `json:"os"`
	Class     string `json:"class"`
	Service   string `json:"service"`
	Version   string `json:"version"`
	Addresses string `json:"Addresses"`
}

type MachineList map[string]MachineInfo

func Machines() (machines MachineList, err error) {
	machines = make(MachineList)

	var stdout, stderr bytes.Buffer

	cmder, err := NewCommander(ExecBackgroundWithFIFOOpts{
		Commands:     []string{"machinectl", "-o", "json", "--no-pager", "list"},
		StdOutBuffer: &stdout,
		StdErrBuffer: &stderr,
	})
	defer cmder.Destroy()

	if err != nil {
		err = fmt.Errorf("error call list of machines: %w", err)
		return
	}
	cmder.Stop()

	var machinesRaw []MachineInfoRaw
	if err = json.Unmarshal(stdout.Bytes(), &machinesRaw); err != nil {
		err = fmt.Errorf("JSON parse failed: %w", err)
		return
	}

	// convert
	for _, machine := range machinesRaw {
		machines[machine.Name] = MachineInfo{
			OS:        machine.OS,
			Class:     machine.Class,
			Service:   machine.Service,
			Version:   machine.Version,
			Addresses: strings.Split(machine.Addresses, "\n"),
		}
	}

	return
}

type MachineState struct {
	ID    string `json:"Id"`
	Name  string `json:"Name"`
	State string `json:"State"`
}

func MachineStateFromName(machineName string) (props MachineState, err error) {
	// machinectl show gibt key=value Zeilen, kein JSON direkt

	var stdout, stderr bytes.Buffer
	cmder, err := NewCommander(ExecBackgroundWithFIFOOpts{
		Commands:     []string{"machinectl", "--all", "show", machineName},
		StdOutBuffer: &stdout,
		StdErrBuffer: &stderr,
	})
	defer cmder.Destroy()
	if err != nil {
		err = fmt.Errorf("error call list of machines: %w", err)
		return
	}
	cmder.Stop()

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

	var jsonBytes []byte
	jsonBytes, err = json.Marshal(propsMap)
	if err != nil {
		return
	}
	err = json.Unmarshal(jsonBytes, &props)

	return
}

func MachineWaitForRunning(machineName string, timeout time.Duration) (state MachineState, err error) {

	ctx, ctxCancel := context.WithTimeout(context.Background(), timeout)
	defer ctxCancel()

	ticker := time.NewTicker(time.Second * 1)

waitLoop:
	for {
		select {
		case <-ticker.C:

			var machineList MachineList
			machineList, err = Machines()
			_, machineExist := machineList[machineName]
			if !machineExist {
				continue
			}

			state, err = MachineStateFromName(machineName)
			if err != nil {
				break waitLoop
			}

			if state.State == "running" {
				break waitLoop
			}

		case <-ctx.Done():
			err = errors.New("timeout on waiting for running-state")
			return
		}
	}

	return
}

func MachineWaitForStopping(machineName string, timeout time.Duration) (err error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), timeout)
	defer ctxCancel()

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	// check that the machine is running
	machineList, err := Machines()
	_, machineExist := machineList[machineName]
	if !machineExist {
		return
	}

	stopCheckLoop := make(chan bool)
	defer func() {
		stopCheckLoop <- true
		close(stopCheckLoop)
	}()
	machineStopped := make(chan bool)
	defer close(machineStopped)

	go func() {
		for {
			select {
			case <-ticker.C:
				machineList, _ := Machines()
				_, machineExist := machineList[machineName]
				if !machineExist {
					machineStopped <- true
				}

			case <-stopCheckLoop:
				return
			}
		}
	}()

	for {
		select {
		case <-machineStopped:
			return

		case <-ctx.Done():
			err = errors.New("timeout on waiting for running-state")
			return
		}
	}
}

func MachineWaitForIPv4(machineName string, timeout time.Duration) (ip string, err error) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), timeout)
	defer ctxCancel()

	ticker := time.NewTicker(time.Second * 1)

waitLoop:
	for {
		select {
		case <-ticker.C:
			var machines MachineList
			machines, err = Machines()
			if err != nil {
				break waitLoop
			}

			currentMachine, exist := machines[machineName]
			if exist {
				for _, address := range currentMachine.Addresses {
					ipAddress := net.ParseIP(address)
					if ipAddress == nil {
						continue
					}

					if ipAddress.To4() != nil {
						ip = address
					}
				}
			}

			if ip != "" {
				break waitLoop
			}

		case <-ctx.Done():
			err = errors.New("timeout on waiting for ip")
			return
		}
	}

	return
}

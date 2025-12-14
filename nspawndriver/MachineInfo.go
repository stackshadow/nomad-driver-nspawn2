package nspawndriver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os/exec"
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

	cmd := exec.Command("machinectl", "-o", "json", "--no-pager", "list")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err = cmd.Run(); err != nil {
		return nil, fmt.Errorf("machinectl failed: %w (stderr: %s)", err, stderr.String())
	}

	var machinesRaw []MachineInfoRaw
	if err = json.Unmarshal(stdout.Bytes(), &machinesRaw); err != nil {
		return nil, fmt.Errorf("JSON parse failed: %w", err)
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

	return machines, nil
}

type MachineState struct {
	ID    string `json:"Id"`
	Name  string `json:"Name"`
	State string `json:"State"`
}

func MachineStateFromName(machineName string) (props MachineState, err error) {
	// machinectl show gibt key=value Zeilen, kein JSON direkt
	cmd := exec.Command("machinectl", "show", machineName)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err = cmd.Run(); err != nil {
		// err = fmt.Errorf("machinectl show %s failed: %w (stderr: %s)", machineName, err, stderr.String())
		err = nil
		return
	}

	// Parst key=value Format zu map
	propsMap := make(map[string]interface{})
	lines := strings.Split(stdout.String(), "\n")
	for _, line := range lines {
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

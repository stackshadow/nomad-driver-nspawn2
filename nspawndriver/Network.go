package nspawndriver

import (
	"fmt"
	"net"
	"os/exec"
	"time"
)

type DriverPort struct {
	Name     string
	HostIP   int // NOMAD_HOST_IP_XXX:127.0.0.1
	HostPort int // NOMAD_HOST_PORT_XXX:26775
	Port     int // NOMAD_PORT_XXX:3000
}

func NetworkWaitForInterface(ifaceName string, timeout time.Duration, pollInterval time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		ifaces, err := net.Interfaces()
		if err != nil {
			return fmt.Errorf("could not list interfaces: %w", err)
		}

		for _, iface := range ifaces {
			if iface.Name == ifaceName {
				return nil
			}
		}

		time.Sleep(pollInterval)
	}

	return fmt.Errorf("timeout: could not find interface '%s'", ifaceName)
}

func (d *NSpawnDriverPlugin) NetworkInterfaceUp(ifaceName string) error {

	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return err
	}

	commandFull := []string{d.config.IPPath, "link", "set", "dev", iface.Name, "up"}
	if d.config.Sudo {
		commandFull = []string{"sudo", d.config.IPPath, "link", "set", "dev", iface.Name, "up"}
	}

	command := commandFull[0]
	commandArgs := commandFull[1:]

	// ip link set dev <iface> up ausführen
	cmd := exec.Command(command, commandArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ip-Befehl fehlgeschlagen für %s: %w\nOutput: %s",
			iface.Name, err, output)
	}

	return nil
}

func (d *NSpawnDriverPlugin) NetworkAddRoute(ifaceName string, ip string) error {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return err
	}

	commandFull := []string{d.config.IPPath, "link", "set", "dev", iface.Name, "up"}
	if d.config.Sudo {
		commandFull = []string{"sudo", d.config.IPPath, "link", "set", "dev", iface.Name, "up"}
	}

	command := commandFull[0]
	commandArgs := commandFull[1:]

	// ip link set dev <iface> up ausführen
	cmd := exec.Command(command, commandArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ip-Befehl fehlgeschlagen für %s: %w\nOutput: %s",
			iface.Name, err, output)
	}

	return nil
}

// sudo ip link set ve-test-driver up
// sudo ip route add 169.254.206.189 dev ve-test-driver

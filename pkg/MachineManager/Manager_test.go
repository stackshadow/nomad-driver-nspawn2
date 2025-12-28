package machinemanager_test

import (
	"testing"
	"time"

	"github.com/shoenig/test/must"
	machinemanager "github.com/stackshadow/nspawn2/pkg/MachineManager"
	"github.com/stackshadow/nspawn2/pkg/MachineManager/domain"
	"github.com/stackshadow/nspawn2/pkg/MachineManager/service"
	networking "github.com/stackshadow/nspawn2/pkg/Networking"
	testutils "github.com/stackshadow/nspawn2/pkg/TestUtils"
)

func TestStart(t *testing.T) {
	basePath := testutils.GetProjectPath("./")

	manager := machinemanager.NewService(service.NewServiceOpts{
		Repo: machinemanager.NewMachineCommandsCliBased(machinemanager.NewOpts{
			UseSudo:           true,
			BinNSPawnPath:     "/run/current-system/systemd/bin/systemd-nspawn",
			BinMachinectlPath: "/run/current-system/systemd/bin/machinectl",
		}),
		Network: networking.NewCLIBasedPort(networking.NewOpts{
			UseSudo:   true,
			BinIPPath: "/nix/store/49av73h3l9rabx0jrac5hcsf1h3x5y6s-iproute2-6.17.0/bin/ip",
		}),
		Timeout: time.Second * 15,
	})

	_, _, err := manager.MachineStart(domain.StartOpts{
		MachineName:   "no-network",
		ImageFileName: basePath + "/debian.raw",
		IsSystemd:     true,
		Ephemeral:     true,
	})
	must.NoError(t, err)
	err = manager.Stop("no-network")
	must.NoError(t, err)

	_, _, err = manager.MachineStart(domain.StartOpts{
		MachineName:   "should-not-be-there",
		ImageFileName: basePath + "/debian.raw",
		IsSystemd:     true,
		Ephemeral:     true,
		Networking: domain.StartNeworkingOpts{
			VethName: "ve-test-driver-with-to-long-network-name",
		},
	})
	must.Error(t, err)

	ipv4, _, err := manager.MachineStart(domain.StartOpts{
		MachineName:   "vethernet",
		ImageFileName: basePath + "/debian.raw",
		IsSystemd:     true,
		Ephemeral:     true,
		Networking: domain.StartNeworkingOpts{
			VethName: "ve-test",
			// BridgeName: "development0",
		},
	})
	must.NoError(t, err)
	_ = ipv4

	err = manager.Stop("vethernet")
	must.NoError(t, err)

	err = manager.WaitForStopping("vethernet")
	must.NoError(t, err)

}

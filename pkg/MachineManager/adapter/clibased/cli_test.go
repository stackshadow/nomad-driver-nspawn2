package clibased_test

import (
	"testing"
	"time"

	"github.com/shoenig/test/must"
	"github.com/stackshadow/nspawn2/pkg/MachineManager/adapter/clibased"
	"github.com/stackshadow/nspawn2/pkg/MachineManager/domain"
)

func TestStartStop(t *testing.T) {
	var err error

	adapter := clibased.New(struct {
		BinNSPawnPath     string
		BinMachinectlPath string
		UseSudo           bool
	}{
		BinNSPawnPath:     "/run/current-system/systemd/bin/systemd-nspawn",
		BinMachinectlPath: "/run/current-system/systemd/bin/machinectl",
		UseSudo:           true,
	})

	startOpts := domain.StartOpts{
		MachineName:   "integrationtest-1",
		ImageFileName: "/mnt/synced/develop/nomad/nspawn-reduced/debian.raw",
		IsSystemd:     true,
		Ephemeral:     true,
	}
	err = adapter.Start(startOpts)
	must.NoError(t, err)

	for {
		state, _ := adapter.State(startOpts.MachineName)
		if state == domain.MachineStateRunning {
			break
		}
		time.Sleep(time.Second)
	}

	machines, err := adapter.List()
	must.NoError(t, err)
	must.Greater(t, 0, len(machines))

	err = adapter.Stop(startOpts.MachineName)
	must.NoError(t, err)

}

func TestStart(t *testing.T) {

	var err error

	adapter := clibased.New(struct {
		BinNSPawnPath     string
		BinMachinectlPath string
		UseSudo           bool
	}{
		BinNSPawnPath:     "/run/current-system/systemd/bin/systemd-nspawn",
		BinMachinectlPath: "/run/current-system/systemd/bin/machinectl",
		UseSudo:           true,
	})

	startOpts := domain.StartOpts{
		MachineName:   "integrationtest-2",
		ImageFileName: "/mnt/synced/develop/nomad/nspawn-reduced/debian.raw",
		IsSystemd:     true,
		Ephemeral:     true,
	}
	err = adapter.Start(startOpts)
	must.NoError(t, err)

	for {
		state, _ := adapter.State(startOpts.MachineName)
		if state == domain.MachineStateRunning {
			break
		}
		time.Sleep(time.Second)
	}

	adapter2 := clibased.New(struct {
		BinNSPawnPath     string
		BinMachinectlPath string
		UseSudo           bool
	}{
		BinNSPawnPath:     "/run/current-system/systemd/bin/systemd-nspawn",
		BinMachinectlPath: "/run/current-system/systemd/bin/machinectl",
		UseSudo:           true,
	})
	err = adapter2.Stop(startOpts.MachineName)
	must.NoError(t, err)

}

func TestStateNotExistingMachine(t *testing.T) {
	var err error

	adapter := clibased.New(struct {
		BinNSPawnPath     string
		BinMachinectlPath string
		UseSudo           bool
	}{
		BinNSPawnPath:     "/run/current-system/systemd/bin/systemd-nspawn",
		BinMachinectlPath: "/run/current-system/systemd/bin/machinectl",
		UseSudo:           true,
	})

	state, err := adapter.State("notexist")
	must.NoError(t, err)
	must.Eq(t, domain.MachineStateNotExist, state)
}

// func TestStop(t *testing.T) {
// 	var err error

// 	adapter := clibased.New(struct {
// 		BinNSPawnPath     string
// 		BinMachinectlPath string
// 		UseSudo           bool
// 	}{
// 		BinNSPawnPath:     "/run/current-system/systemd/bin/systemd-nspawn",
// 		BinMachinectlPath: "/run/current-system/systemd/bin/machinectl",
// 		UseSudo:           true,
// 	})

// 	err = adapter.Stop("development-nspawn-5b59dce6-7ad5-db9a-585a-b1d4ab04c6d0")
// 	assert.NoError(t, err)
// }

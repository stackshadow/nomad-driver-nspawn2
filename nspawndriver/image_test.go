package nspawndriver_test

import (
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/stackshadow/nspawn2/nspawndriver"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/nomad/ci"
	"github.com/hashicorp/nomad/helper/pluginutils/hclspecutils"
	"github.com/hashicorp/nomad/helper/pluginutils/hclutils"
	"github.com/hashicorp/nomad/helper/testlog"
	"github.com/hashicorp/nomad/helper/uuid"
	"github.com/hashicorp/nomad/nomad/structs"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/shoenig/test/must"
)

var testResources = &drivers.Resources{
	NomadResources: &structs.AllocatedTaskResources{
		Memory: structs.AllocatedMemoryResources{
			MemoryMB: 128,
		},
		Cpu: structs.AllocatedCpuResources{
			CpuShares: 100,
		},
	},
	LinuxResources: &drivers.LinuxResources{
		MemoryLimitBytes: 134217728,
		CPUShares:        100,
	},
}

func TestDriver_Start_Fingerprint(t *testing.T) {

	logger := testlog.HCLogger(t)
	if testing.Verbose() {
		logger.SetLevel(hclog.Trace)
	} else {
		logger.SetLevel(hclog.Info)
	}

	driver := nspawndriver.NewPlugin(logger)

	fingerprintChannel, err := driver.Fingerprint(t.Context())
	must.NoError(t, err)

	fingerprint := <-fingerprintChannel
	must.NoError(t, fingerprint.Err)

	must.MapContainsKey(t, fingerprint.Attributes, "driver.nspawn2.systemd_version")
}

func TestDriver_InvalidConfig(t *testing.T) {
	ci.Parallel(t)

	task := &drivers.TaskConfig{
		ID:      uuid.Generate(),
		Name:    "echo",
		AllocID: uuid.Generate(),
	}
	taskCfg := nspawndriver.TaskConfig{
		Image:            "/mnt/synced/develop/nomad/nspawn-reduced/debian.raw",
		Boot:             true,
		Ephemeral:        true,
		NetworkPrivate:   false,
		NetworkVeth:      false,
		NetworkVethExtra: "ve-test-driver-with-to-long-network-name",
	}
	must.NoError(t, task.EncodeConcreteDriverConfig(&taskCfg))

	d := driverHarness(t)
	cleanup := d.MkAllocDir(task, false)
	defer cleanup()

	_, _, err := d.StartTask(task)
	must.Error(t, err)
}

func TestDriver_schema(t *testing.T) {
	logger := testlog.HCLogger(t)
	if testing.Verbose() {
		logger.SetLevel(hclog.Trace)
	} else {
		logger.SetLevel(hclog.Info)
	}

	driver := nspawndriver.NewPlugin(logger)

	spec, err := driver.ConfigSchema()
	must.NoError(t, err)

	validHCL := `
	config {
		sudo = true
		nspawn_path = "" # default
	}
  `
	var tc nspawndriver.DriverConfig
	parser := hclutils.NewConfigParser(spec)
	parser.ParseHCL(t, validHCL, &tc)

	hclutils.HclConfigToInterface(t, validHCL)
	_, diags := hclspecutils.Convert(spec)
	require.Empty(t, diags)

}

func TestDriver_Start_nonetwork(t *testing.T) {
	ci.Parallel(t)

	task := &drivers.TaskConfig{
		ID:      uuid.Generate(),
		Name:    "echo",
		AllocID: uuid.Generate(),
	}
	taskCfg := nspawndriver.TaskConfig{
		Image:     "/mnt/synced/develop/nomad/nspawn-reduced/debian.raw",
		Boot:      true,
		Ephemeral: true,
	}
	must.NoError(t, task.EncodeConcreteDriverConfig(&taskCfg))

	d := driverHarness(t)
	cleanup := d.MkAllocDir(task, false)
	defer cleanup()

	_, _, err := d.StartTask(task)
	must.NoError(t, err)
	defer func() {
		err = d.StopTask(task.ID, time.Second*30, "")
		must.NoError(t, err)

		result, err := d.WaitTask(t.Context(), task.ID)
		<-result
		must.NoError(t, err)

		err = d.DestroyTask(task.ID, true)
		must.NoError(t, err)
	}()

}

func TestDriver_Start_network(t *testing.T) {
	ci.Parallel(t)

	task := &drivers.TaskConfig{
		ID:      "nsdriver-integrationtest-2",
		Name:    "echo",
		AllocID: uuid.Generate(),
	}
	taskCfg := nspawndriver.TaskConfig{
		Image:            "/mnt/synced/develop/nomad/nspawn-reduced/debian.raw",
		Boot:             true,
		Ephemeral:        true,
		NetworkPrivate:   false,
		NetworkVeth:      false,
		NetworkVethExtra: "ve-test-driver",
	}
	must.NoError(t, task.EncodeConcreteDriverConfig(&taskCfg))

	d := driverHarness(t)
	cleanup := d.MkAllocDir(task, false)
	defer cleanup()

	_, _, err := d.StartTask(task)
	must.NoError(t, err)
	defer func() {
		err = d.StopTask(task.ID, time.Second*30, "")
		must.NoError(t, err)

		result, err := d.WaitTask(t.Context(), task.ID)
		<-result
		must.NoError(t, err)

		err = d.DestroyTask(task.ID, true)
		must.NoError(t, err)
	}()

}
func TestDriver_Stop(t *testing.T) {
	ci.Parallel(t)

	task := &drivers.TaskConfig{
		ID:      uuid.Generate(),
		Name:    "echo",
		AllocID: uuid.Generate(),
	}
	taskCfg := nspawndriver.TaskConfig{
		Image:            "/mnt/synced/develop/nomad/nspawn-reduced/debian.raw",
		Boot:             true,
		Ephemeral:        true,
		NetworkPrivate:   false,
		NetworkVeth:      false,
		NetworkVethExtra: "ve-test-driver",
	}
	must.NoError(t, task.EncodeConcreteDriverConfig(&taskCfg))

	d := driverHarness(t)
	cleanup := d.MkAllocDir(task, false)
	defer cleanup()

}

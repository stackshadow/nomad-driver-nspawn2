package nspawndriver_test

import (
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/client/lib/numalib"
	"github.com/hashicorp/nomad/helper/testlog"
	"github.com/hashicorp/nomad/plugins/base"
	dtestutil "github.com/hashicorp/nomad/plugins/drivers/testutils"
	"github.com/shoenig/test/must"
	"github.com/stackshadow/nspawn2/nspawndriver"
)

// podmanDriverHarness wires up everything needed to launch a task with a podman driver.
// A driver plugin interface and cleanup function is returned
func driverHarness(t *testing.T) *dtestutil.DriverHarness {
	logger := testlog.HCLogger(t)
	if testing.Verbose() {
		logger.SetLevel(hclog.Trace)
	} else {
		logger.SetLevel(hclog.Info)
	}

	baseConfig := base.Config{
		AgentConfig: &base.AgentConfig{
			Driver: &base.ClientDriverConfig{
				Topology: numalib.Scan(numalib.PlatformScanners(true)),
			},
		},
	}
	pluginConfig := nspawndriver.DriverConfig{
		Sudo:       true,
		NSPawnPath: "/run/current-system/systemd/bin/systemd-nspawn",
		IPPath:     "/nix/store/49av73h3l9rabx0jrac5hcsf1h3x5y6s-iproute2-6.17.0/bin/ip",
	}

	if err := base.MsgPackEncode(&baseConfig.PluginConfig, &pluginConfig); err != nil {
		t.Error("Unable to encode plugin config", err)
	}

	driver := nspawndriver.NewPlugin(logger)
	must.NoError(t, driver.SetConfig(&baseConfig))
	//driver.buildFingerprint()

	harness := dtestutil.NewDriverHarness(t, driver)

	return harness
}

func getDriver(t *testing.T, harness *dtestutil.DriverHarness) *nspawndriver.NSpawnDriverPlugin {
	driver, ok := harness.Impl().(*nspawndriver.NSpawnDriverPlugin)
	must.True(t, ok)
	return driver
}

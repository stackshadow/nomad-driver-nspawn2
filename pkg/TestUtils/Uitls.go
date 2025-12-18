package testutils

import (
	"io/fs"
	"path/filepath"
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
func DriverHarness(t *testing.T) *dtestutil.DriverHarness {
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
		MachineCtl: "/run/current-system/systemd/bin/machinectl",
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

func GetProjectPath(basePath string) (projectPath string) {

	basePath, _ = filepath.Abs(basePath)

	for range 5 {
		filepath.Walk(basePath, func(basePath string, info fs.FileInfo, err error) error {

			if projectPath != "" {
				return nil
			}

			// dir
			if info.IsDir() {
				return nil
			}

			// file
			if filepath.Base(basePath) == "go.mod" {
				projectPath = basePath
				return nil
			}
			return nil
		})

		if projectPath == "" {
			basePath, _ = filepath.Abs(basePath + "/..")
		}
	}

	if projectPath != "" {
		projectPath = filepath.Dir(projectPath)
	}

	return
}

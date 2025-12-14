package nspawndriver

import (
	"github.com/hashicorp/nomad/plugins/base"
	"github.com/hashicorp/nomad/plugins/shared/hclspec"
)

var (
	// configSpec is the specification of the plugin's configuration
	// this is used to validate the configuration specified for the plugin
	// on the client.
	// this is not global, but can be specified on a per-client basis.
	// configSpec = hclspec.NewObject(map[string]*hclspec.Spec{
	// 	"sudo": hclspec.NewDefault(
	// 		hclspec.NewAttr("sudo", "bool", false),
	// 		hclspec.NewLiteral("false"),
	// 	),
	// 	"nspawnpath": hclspec.NewDefault(
	// 		hclspec.NewAttr("nspawnpath", "string", false),
	// 		hclspec.NewLiteral("/usr/bin/systemd-nspawn"),
	// 	),
	// })

	configSpec = hclspec.NewObject(map[string]*hclspec.Spec{
		"sudo":        hclspec.NewAttr("sudo", "bool", false),
		"nspawn_path": hclspec.NewAttr("nspawn_path", "string", false),
		"ip_path":     hclspec.NewAttr("ip_path", "string", false),

		// "nspawnpath": hclspec.NewDefault(
		// 	hclspec.NewAttr("nspawnpath", "string", false),
		// 	hclspec.NewLiteral("/usr/bin/systemd-nspawn"),
		// ),
	})
)

// DriverConfig contains configuration information for the plugin
type DriverConfig struct {
	Sudo       bool   `codec:"sudo"`
	NSPawnPath string `codec:"nspawn_path"` // path to systemd-nspawn
	IPPath     string `codec:"ip_path"`     // path to ip
}

// ConfigSchema returns the plugin configuration schema.
func (d *NSpawnDriverPlugin) ConfigSchema() (*hclspec.Spec, error) {
	return configSpec, nil
}

// SetConfig is called by the client to pass the configuration for the plugin.
func (d *NSpawnDriverPlugin) SetConfig(cfg *base.Config) error {

	var config DriverConfig
	if len(cfg.PluginConfig) != 0 {
		if err := base.MsgPackDecode(cfg.PluginConfig, &config); err != nil {
			return err
		}
	}

	if config.NSPawnPath == "" {
		if config.Sudo {
			config.Sudo = false
			d.logger.Warn("no absolut path to systemd-nspawn, can not use sudo")
		}

		config.NSPawnPath = "systemd-nspawn"
		d.logger.Debug("no nspawn-path set, use default", "nspawn_path", config.NSPawnPath)
	}

	if config.IPPath == "" {
		if config.Sudo {
			config.Sudo = false
			d.logger.Warn("no absolut path to systemd-nspawn, can not use sudo")
		}

		config.IPPath = "ip"
		d.logger.Debug("no ip set, use default", "ip_path", config.NSPawnPath)
	}

	// @TODO: Check if binary paths exist

	// Save the configuration to the plugin
	d.config = &config

	// Save the Nomad agent configuration
	if cfg.AgentConfig != nil {
		d.nomadConfig = cfg.AgentConfig.Driver
	}

	// TODO: initialize any extra requirements if necessary.
	//
	// Here you can use the config values to initialize any resources that are
	// shared by all tasks that use this driver, such as a daemon process.

	return nil
}

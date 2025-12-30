package nspawndriver

import (
	"fmt"

	"github.com/hashicorp/nomad/helper/pluginutils/hclutils"
	"github.com/hashicorp/nomad/plugins/shared/hclspec"
)

var (

	// taskConfigSpec is the specification of the plugin's configuration for
	// a task
	// this is used to validated the configuration specified for the plugin
	// when a job is submitted.
	TaskConfigSpec = hclspec.NewObject(map[string]*hclspec.Spec{
		"hostname": hclspec.NewAttr("hostname", "string", false),

		"boot": hclspec.NewDefault(
			hclspec.NewAttr("boot", "bool", false),
			hclspec.NewLiteral("true"),
		),
		"ephemeral": hclspec.NewAttr("ephemeral", "bool", false),
		"read_only": hclspec.NewAttr("read_only", "bool", false),

		"image": hclspec.NewAttr("image", "string", true),

		"environment": hclspec.NewAttr("environment", "list(map(string))", false),

		"bind":           hclspec.NewAttr("bind", "list(map(string))", false),
		"bind_read_only": hclspec.NewAttr("bind_read_only", "list(map(string))", false),

		"commands": hclspec.NewAttr("commands", "list(string)", false),

		"network_private":    hclspec.NewAttr("network_private", "bool", false),
		"network_veth":       hclspec.NewAttr("network_veth", "bool", false),
		"network_veth_extra": hclspec.NewAttr("network_veth_extra", "string", false),
		"network_bridge":     hclspec.NewAttr("network_bridge", "string", false),
	})
)

// TaskConfig contains configuration information for a task that runs with
// this plugin
type TaskConfig struct {
	Hostname string `codec:"hostname"`

	Boot      bool   `codec:"boot"`
	Ephemeral bool   `codec:"ephemeral"`
	ReadOnly  bool   `codec:"read_only"`
	Image     string `codec:"image"`

	Environment hclutils.MapStrStr `codec:"environment"`

	Bind         hclutils.MapStrStr `codec:"bind"`
	BindReadOnly hclutils.MapStrStr `codec:"bind_read_only"`

	Commands []string `codec:"commands"`

	// Networking
	NetworkPrivate bool `codec:"network_private"`
	NetworkVeth    bool `codec:"network_veth"`

	// Sets the virtual ethernet interface name, this will try to bring the interface up
	// and sets a route to it
	NetworkVethExtra string `codec:"network_veth_extra"`

	NetworkBridge string `codec:"network_bridge"`
}

func (c *TaskConfig) Validate() error {

	if len(c.NetworkVethExtra) > 15 {
		return fmt.Errorf("network interface names can not be larger than 15, yours is %v", len(c.NetworkVethExtra))
	}

	if len(c.NetworkBridge) > 0 && len(c.NetworkVethExtra) > 0 {
		return fmt.Errorf("you can not use bridges with veth-extra")
	}

	return nil
}

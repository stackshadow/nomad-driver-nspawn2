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
		"name":     hclspec.NewAttr("name", "string", false),
		"hostname": hclspec.NewAttr("hostname", "string", false),

		"boot": hclspec.NewDefault(
			hclspec.NewAttr("boot", "bool", false),
			hclspec.NewLiteral("true"),
		),
		"ephemeral": hclspec.NewAttr("ephemeral", "bool", false),

		"image": hclspec.NewAttr("image", "string", true),

		"resolv_conf": hclspec.NewAttr("resolv_conf", "string", false),

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
	MachineName string `codec:"name"`
	Hostname    string `codec:"hostname"`

	Boot      bool   `codec:"boot"`
	Ephemeral bool   `codec:"ephemeral"`
	Image     string `codec:"image"`

	ResolvConf string `codec:"resolv_conf"`

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

func (c *TaskConfig) ToCLIParameter() ([]string, error) {
	if c.Image == "" {
		return nil, fmt.Errorf("no image configured")
	}

	args := []string{"-i", c.Image}

	if c.MachineName != "" {
		args = append(args, "--machine", c.MachineName)
	}
	if c.Hostname != "" {
		args = append(args, "--hostname", c.Hostname)
	}
	if c.Boot {
		args = append(args, "--boot")
	}
	if c.Ephemeral {
		args = append(args, "--ephemeral")
	}

	if c.ResolvConf != "" {
		args = append(args, "--resolv-conf", c.ResolvConf)
	}

	for k, v := range c.Bind {
		args = append(args, "--bind", k+":"+v)
	}
	for k, v := range c.BindReadOnly {
		args = append(args, "--bind-ro", k+":"+v)
	}
	for k, v := range c.Environment {
		args = append(args, "-E", k+"="+v)
	}

	if c.NetworkPrivate {
		args = append(args, "--private-network")
	}
	if c.NetworkVeth {
		args = append(args, "--network-veth")
	}
	if c.NetworkVethExtra != "" {
		args = append(args, "--network-veth-extra", c.NetworkVethExtra)
	}
	if len(c.NetworkBridge) > 0 {
		args = append(args, "--network-bridge", c.NetworkBridge)
	}

	return args, nil
}

package networking

import (
	"github.com/stackshadow/nspawn2/pkg/Networking/adapter/clibased"
	"github.com/stackshadow/nspawn2/pkg/Networking/ports"
)

type NewOpts struct {
	BinIPPath string
	UseSudo   bool
}

func NewCLIBasedPort(opts NewOpts) ports.NetworkCommands {
	return clibased.New(opts)
}

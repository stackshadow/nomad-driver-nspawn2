package clibased

import (
	"github.com/stackshadow/nspawn2/pkg/MachineManager/ports"
)

type adapterData struct {
	binNSPawnPath     string
	binMachinectlPath string
	useSudo           bool
}

func New(opts struct {
	BinNSPawnPath     string
	BinMachinectlPath string
	UseSudo           bool
}) (adapter ports.MachineCommands) {
	return &adapterData{
		useSudo:           opts.UseSudo,
		binNSPawnPath:     opts.BinNSPawnPath,
		binMachinectlPath: opts.BinMachinectlPath,
	}
}

type MachineInfoRaw struct {
	Name      string `json:"machine"`
	OS        string `json:"os"`
	Class     string `json:"class"`
	Service   string `json:"service"`
	Version   string `json:"version"`
	Addresses string `json:"Addresses"`
}

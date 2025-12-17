package clibased

import (
	"fmt"
	"net"

	commander "github.com/stackshadow/nspawn2/pkg/Commander"
	"github.com/stackshadow/nspawn2/pkg/Networking/ports"
)

type adapterData struct {
	binIPPath string
	useSudo   bool
}

func New(opts struct {
	BinIPPath string
	UseSudo   bool
}) (adapter ports.NetworkCommands) {
	return &adapterData{
		binIPPath: opts.BinIPPath,
		useSudo:   opts.UseSudo,
	}
}

func (a *adapterData) Exist(name string) (exist bool, err error) {

	var iface *net.Interface
	iface, err = net.InterfaceByName(name)
	if err != nil {
		return
	}

	exist = iface != nil
	return
}

func (a *adapterData) Up(name string) (err error) {

	cmder := commander.New(commander.NewCommanderOpts{})
	defer cmder.Destroy()
	defer cmder.Stop()

	err = cmder.Run(commander.RunOpts{
		UseSudo:  a.useSudo,
		Commands: []string{a.binIPPath, "link", "set", "dev", name, "up"},
	})
	if err != nil {
		err = fmt.Errorf("bring interface '%s' up: %w", name, err)
		return
	}
	cmder.Wait()

	return
}

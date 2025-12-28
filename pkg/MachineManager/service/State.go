package service

import (
	"github.com/stackshadow/nspawn2/pkg/MachineManager/domain"
)

func (svc *Service) State(machineName string) (state domain.MachineState, err error) {
	return svc.repo.State(machineName)
}

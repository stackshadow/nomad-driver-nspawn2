package service

import (
	"fmt"
)

func (svc *Service) Stop(machineName string) (err error) {

	err = svc.repo.Stop(machineName)
	if err != nil {
		err = fmt.Errorf("start machine %s: %w", machineName, err)
		return
	}

	return
}

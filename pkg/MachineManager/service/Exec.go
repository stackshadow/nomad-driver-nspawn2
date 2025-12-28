package service

import "fmt"

func (svc *Service) Exec(machineName string, commands ...string) (err error) {
	err = svc.repo.Exec(machineName, commands...)
	if err != nil {
		err = fmt.Errorf("exec in machine %s: %w", machineName, err)
		return
	}

	return
}

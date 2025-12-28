package service

import (
	"context"
	"fmt"
	"time"

	"github.com/stackshadow/nspawn2/pkg/MachineManager/domain"
)

func (svc *Service) WaitForIPv4(machineName string) (ip string, err error) {

	ctx, ctxCancel := context.WithTimeout(context.Background(), svc.timeout)
	defer ctxCancel()

	ticker := time.NewTicker(time.Second * 1)

waitLoop:
	for {
		select {
		case <-ticker.C:

			var machines domain.MachineList
			machines, err = svc.repo.List()

			currentMachine, exist := machines[machineName]
			if exist {
				if currentMachine.Addresses.IPv4 != "" {
					ip = currentMachine.Addresses.IPv4
					break waitLoop

				}
			}

		case <-ctx.Done():
			err = fmt.Errorf("timeout on waiting for ipv4 of machine %s", machineName)
			return
		}
	}

	return
}

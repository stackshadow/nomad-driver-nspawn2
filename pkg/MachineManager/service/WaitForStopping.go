package service

import (
	"context"
	"fmt"
	"time"

	"github.com/stackshadow/nspawn2/pkg/MachineManager/domain"
)

func (svc *Service) WaitForStopping(machineName string) (err error) {

	ctx, ctxCancel := context.WithTimeout(context.Background(), svc.timeout)
	defer ctxCancel()

	ticker := time.NewTicker(time.Second * 1)

waitLoop:
	for {
		select {
		case <-ticker.C:

			var state domain.MachineState
			state, err = svc.repo.State(machineName)
			if err != nil {
				return
			}

			if state == domain.MachineStateNotExist {
				break waitLoop
			}

			if state == domain.MachineStateRunning {
				break waitLoop
			}

		case <-ctx.Done():
			err = fmt.Errorf("timeout on waiting for machine %s to stop", machineName)
			return
		}
	}

	return
}

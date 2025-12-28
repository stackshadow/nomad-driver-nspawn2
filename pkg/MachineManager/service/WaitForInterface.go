package service

import (
	"context"
	"fmt"
	"time"
)

func (svc *Service) WaitForInterface(interfaceName string) (err error) {

	ctx, ctxCancel := context.WithTimeout(context.Background(), svc.timeout)
	defer ctxCancel()

	ticker := time.NewTicker(time.Second * 1)

waitLoop:
	for {
		select {
		case <-ticker.C:

			var exist bool
			exist, err = svc.network.Exist(interfaceName)
			if err != nil {
				return
			}

			if exist {
				break waitLoop
			}

		case <-ctx.Done():
			err = fmt.Errorf("timeout on waiting for interface %s", interfaceName)
			return
		}
	}

	return
}

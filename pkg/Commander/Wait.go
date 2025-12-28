package commander

import (
	"errors"
	"time"
)

func (cmder *Commander) Wait() (err error) {

	// Already stopped
	if !cmder.IsAlive() {
		return
	}

	timoutTimer := time.NewTimer(cmder.timeoutDuration)

	select {
	case <-timoutTimer.C:
		err = errors.New("timeout on waiting for finish tasks")
		return
	case <-cmder.waitFinished:
		return
	}

}

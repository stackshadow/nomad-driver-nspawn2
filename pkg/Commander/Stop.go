package commander

import (
	"errors"
	"time"

	"golang.org/x/sys/unix"
)

func (cmder *Commander) Stop() (err error) {

	timeTicker := time.NewTicker(time.Millisecond * 200)
	defer timeTicker.Stop()
	timeoutSigterm := time.NewTimer(cmder.timeoutDuration)
	defer timeoutSigterm.Stop()

	// Already stopped
	if !cmder.IsAlive() {
		return
	}

	// SIGTERM
	timeoutSigterm.Reset(cmder.timeoutDuration)
	unix.Kill(cmder.pid, unix.SIGTERM)
waitForTerm:
	for {
		select {
		case <-timeTicker.C:
			if !cmder.IsAlive() {
				return
			}
		case <-cmder.waitFinished:
			continue
		case <-timeoutSigterm.C:
			break waitForTerm
		}
	}

	// SIGKILL
	timeoutSigterm.Reset(cmder.timeoutDuration)
	unix.Kill(cmder.pid, unix.SIGKILL)
waitForKill:
	for {
		select {
		case <-timeTicker.C:
			if !cmder.IsAlive() {
				return
			}
		case <-cmder.waitFinished:
			continue
		case <-timeoutSigterm.C:
			break waitForKill
		}
	}

	return errors.New("timeout on killing process")
}

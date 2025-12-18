package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/stackshadow/nspawn2/pkg/Stats/domain"
)

type Service struct {
	machineName string
	stop        chan bool
	refreshTime time.Duration
}

func New(options ...Option) *Service {
	newService := &Service{
		stop:        make(chan bool),
		refreshTime: time.Second,
	}

	// call all options
	for _, opt := range options {
		opt(newService)
	}

	return newService
}

func (svc *Service) Watch(callback func(stat domain.Stat)) (err error) {

	machineName := strings.ReplaceAll(svc.machineName, "-", "\\x2d")
	path := fmt.Sprintf("/machine.slice/machine-%s.scope", machineName)

	cg, err := cgroup2.Load(path)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(svc.refreshTime)

	var prevCPU uint64

readLoop:
	for {
		select {

		case <-ticker.C:
			st, err := cg.Stat()
			if err != nil {
				continue
			}

			// // CPU % (delta über 1s)
			cpuPercent := cpuUsagePercent(st.CPU.UsageUsec, prevCPU)

			// // RAM %
			// ramPercent := ramUsagePercent(st.Memory)

			if callback != nil {
				callback(domain.Stat{
					CPUTicks:        float64(st.CPU.UsageUsec - prevCPU),
					CPUUsagePercent: cpuPercent,

					RamUsaged: st.Memory.Usage,
				})
			}

			prevCPU = st.CPU.UsageUsec

		case <-svc.stop:
			svc.stop <- true
			break readLoop
		}

	}

	return
}

// func ramUsagePercent(mem *stats.MemoryStat) float64 {
// 	if mem.MaxUsage == 18446744073709551615 { // max uint64
// 		return 0
// 	}
// 	return (float64(mem.Usage) / float64(mem.UsageLimit)) * 100
// }

func cpuUsagePercent(current, prev uint64) float64 {
	delta := current - prev
	return float64(delta) / 1000000.0 // usec -> Sekunden * 100
}

func (svc *Service) Destroy() {
	svc.stop <- true
	<-svc.stop

	close(svc.stop)
}

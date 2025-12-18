package service

import "time"

type Option func(adapter *Service)

func SetRefreshTime(refreshTime time.Duration) Option {
	return func(adapter *Service) {
		adapter.refreshTime = refreshTime
	}
}

func SetMachineName(machineName string) Option {
	return func(adapter *Service) {
		adapter.machineName = machineName
	}
}

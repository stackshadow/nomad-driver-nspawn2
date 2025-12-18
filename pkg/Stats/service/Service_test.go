package service_test

import (
	"testing"

	"github.com/stackshadow/nspawn2/pkg/Stats/domain"
	"github.com/stackshadow/nspawn2/pkg/Stats/service"
)

func TestService(t *testing.T) {

	receiver := make(chan domain.Stat)

	svc := service.New(
		service.SetMachineName("no-network"),
	)

	go svc.Watch(func(stat domain.Stat) {
		receiver <- stat
	})
	<-receiver
	<-receiver
	<-receiver

	svc.Destroy()
}

package stats

import "github.com/stackshadow/nspawn2/pkg/Stats/service"

func NewService(options ...service.Option) *service.Service {
	return service.New(options...)
}

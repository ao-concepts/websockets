package mock

import (
	"github.com/ao-concepts/logging"
	"github.com/ao-concepts/websockets"
	"github.com/stretchr/testify/mock"
)

// ServiceContainer mock
type ServiceContainer struct {
	mock.Mock
	log logging.Logger
	cfg *websockets.ServerConfig
}

func (cnt *ServiceContainer) GetLogger() websockets.Logger {
	return cnt.log
}

func (cnt *ServiceContainer) GetWebsocketsConfig() *websockets.ServerConfig {
	return cnt.cfg
}

func NewServiceContainer() websockets.ServiceContainer {
	return &ServiceContainer{
		log: logging.New(logging.Debug, nil),
		cfg: nil,
	}
}

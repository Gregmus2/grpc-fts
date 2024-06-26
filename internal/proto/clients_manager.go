package proto

import (
	"github.com/pkg/errors"
	"github.com/res-am/grpc-fts/internal/config"
)

type clientsManager struct {
	clients map[string]Client
}

func NewClientsManager(services config.Services, manager DescriptorsManager) (ClientsManager, error) {
	cm := &clientsManager{
		clients: make(map[string]Client, len(services)),
	}
	for name, service := range services {
		conn, err := newConnection(service.Address, service.TLS)
		if err != nil {
			return nil, errors.Wrapf(err, "error creating connection for service %v", service)
		}

		cm.clients[name] = newClient(conn, manager)
	}

	return cm, nil
}

func (c *clientsManager) GetClient(serviceName string) Client {
	return c.clients[serviceName]
}

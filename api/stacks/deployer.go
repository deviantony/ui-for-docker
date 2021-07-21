package stacks

import (
	"sync"

	portainer "github.com/portainer/portainer/api"
)

type StackDeployer interface {
	DeploySwarmStack(stack *portainer.Stack, endpoint *portainer.Endpoint, registries []portainer.Registry, prune bool) error
	DeployComposeStack(stack *portainer.Stack, endpoint *portainer.Endpoint, registries []portainer.Registry) error
}

type stackDeployer struct {
	lock                *sync.Mutex
	swarmStackManager   portainer.SwarmStackManager
	composeStackManager portainer.ComposeStackManager
}

func NewStackDeployer(swarmStackManager portainer.SwarmStackManager, composeStackManager portainer.ComposeStackManager) *stackDeployer {
	return &stackDeployer{
		lock:                &sync.Mutex{},
		swarmStackManager:   swarmStackManager,
		composeStackManager: composeStackManager,
	}
}

func (d *stackDeployer) DeploySwarmStack(stack *portainer.Stack, endpoint *portainer.Endpoint, registries []portainer.Registry, prune bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.swarmStackManager.Login(registries, endpoint)
	defer d.swarmStackManager.Logout(endpoint)

	return d.swarmStackManager.Deploy(stack, prune, endpoint)
}

func (d *stackDeployer) DeployComposeStack(stack *portainer.Stack, endpoint *portainer.Endpoint, registries []portainer.Registry) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.swarmStackManager.Login(registries, endpoint)
	defer d.swarmStackManager.Logout(endpoint)

	return d.composeStackManager.Up(stack, endpoint)
}

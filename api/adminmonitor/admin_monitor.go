package adminmonitor

import (
	"context"
	"log"
	"time"

	portainer "github.com/portainer/portainer/api"
)

var logFatalf = log.Fatalf

type Monitor struct {
	timeout          time.Duration
	datastore        portainer.DataStore
	shutdownCtx      context.Context
	cancellationFunc context.CancelFunc
}

// New creates a monitor that when started will wait for the timeout duration and then shutdown the application unless it has been initialized.
func New(timeout time.Duration, datastore portainer.DataStore, shutdownCtx context.Context) *Monitor {
	return &Monitor{
		timeout:     timeout,
		datastore:   datastore,
		shutdownCtx: shutdownCtx,
	}
}

// Starts starts the monitor. Active monitor could be stopped or shuttted down by cancelling the shutdown context.
func (m *Monitor) Start() {
	cancellationCtx, cancellationFunc := context.WithCancel(context.Background())
	m.cancellationFunc = cancellationFunc

	go func() {
		log.Println("[DEBUG] [internal,init] [message: start initialization monitor ]")
		select {
		case <-time.After(m.timeout):
			initialized, err := m.WasInitialized()
			if err != nil {
				logFatalf("failed getting admin user: %s", err)
			}
			if !initialized {
				logFatalf("[FATAL] [internal,init] No administrator account was created in %f mins. Shutting down the Portainer instance for security reasons", m.timeout.Minutes())
			}
		case <-cancellationCtx.Done():
			log.Println("[DEBUG] [internal,init] [message: canceling initialization monitor]")
		case <-m.shutdownCtx.Done():
			log.Println("[DEBUG] [internal,init] [message: shutting down initialization monitor]")
		}
	}()
}

// Stop stops monitor. Safe to call even if monitor wasn't started.
func (m *Monitor) Stop() {
	if m.cancellationFunc == nil {
		return
	}
	m.cancellationFunc()
	m.cancellationFunc = nil
}

// WasInitialized is a system initialization check
func (m *Monitor) WasInitialized() (bool, error) {
	users, err := m.datastore.User().UsersByRole(portainer.AdministratorRole)
	if err != nil {
		return false, err
	}
	return len(users) > 0, nil
}

package exec

import (
	"fmt"
	"os"
	"path"
	"strings"

	wrapper "github.com/portainer/docker-compose-wrapper"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/proxy"
	"github.com/portainer/portainer/api/http/proxy/factory"
	"github.com/portainer/portainer/api/internal/stackutils"
)

// ComposeStackManager is a wrapper for docker-compose binary
type ComposeStackManager struct {
	wrapper      *wrapper.ComposeWrapper
	dataPath     string
	proxyManager *proxy.Manager
}

// NewComposeStackManager returns a docker-compose wrapper if corresponding binary present, otherwise nil
func NewComposeStackManager(binaryPath string, dataPath string, proxyManager *proxy.Manager) (*ComposeStackManager, error) {
	wrap, err := wrapper.NewComposeWrapper(binaryPath)
	if err != nil {
		return nil, err
	}

	return &ComposeStackManager{
		wrapper:      wrap,
		proxyManager: proxyManager,
		dataPath:     dataPath,
	}, nil
}

// ComposeSyntaxMaxVersion returns the maximum supported version of the docker compose syntax
func (w *ComposeStackManager) ComposeSyntaxMaxVersion() string {
	return portainer.ComposeSyntaxMaxVersion
}

// Up builds, (re)creates and starts containers in the background. Wraps `docker-compose up -d` command
func (w *ComposeStackManager) Up(stack *portainer.Stack, endpoint *portainer.Endpoint) error {
	url, proxy, err := w.fetchEndpointProxy(endpoint)
	if err != nil {
		return err
	}

	if proxy != nil {
		defer proxy.Close()
	}

	envFilePath, err := createEnvFile(stack)
	if err != nil {
		return err
	}

	filePaths := stackutils.GetStackFilePaths(stack)
	_, err = w.wrapper.Up(filePaths, url, stack.Name, envFilePath, w.dataPath)
	return err
}

// Down stops and removes containers, networks, images, and volumes. Wraps `docker-compose down --remove-orphans` command
func (w *ComposeStackManager) Down(stack *portainer.Stack, endpoint *portainer.Endpoint) error {
	url, proxy, err := w.fetchEndpointProxy(endpoint)
	if err != nil {
		return err
	}
	if proxy != nil {
		defer proxy.Close()
	}

	filePaths := stackutils.GetStackFilePaths(stack)

	_, err = w.wrapper.Down(filePaths, url, stack.Name)
	return err
}

// NormalizeStackName returns the passed stack name, for interface implementation only
func (w *ComposeStackManager) NormalizeStackName(name string) string {
	return name
}

func (w *ComposeStackManager) fetchEndpointProxy(endpoint *portainer.Endpoint) (string, *factory.ProxyServer, error) {
	if strings.HasPrefix(endpoint.URL, "unix://") || strings.HasPrefix(endpoint.URL, "npipe://") {
		return "", nil, nil
	}

	proxy, err := w.proxyManager.CreateComposeProxyServer(endpoint)
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("http://127.0.0.1:%d", proxy.Port), proxy, nil
}

func createEnvFile(stack *portainer.Stack) (string, error) {
	if stack.Env == nil || len(stack.Env) == 0 {
		return "", nil
	}

	envFilePath := path.Join(stack.ProjectPath, "stack.env")

	envfile, err := os.OpenFile(envFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return "", err
	}

	for _, v := range stack.Env {
		envfile.WriteString(fmt.Sprintf("%s=%s\n", v.Name, v.Value))
	}
	envfile.Close()

	return envFilePath, nil
}

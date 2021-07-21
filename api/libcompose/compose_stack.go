package libcompose

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/portainer/libcompose/config"
	"github.com/portainer/libcompose/docker"
	"github.com/portainer/libcompose/docker/client"
	"github.com/portainer/libcompose/docker/ctx"
	"github.com/portainer/libcompose/lookup"
	"github.com/portainer/libcompose/project"
	"github.com/portainer/libcompose/project/options"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/internal/stackutils"
)

const (
	dockerClientVersion     = "1.24"
	composeSyntaxMaxVersion = "2"
)

// ComposeStackManager represents a service for managing compose stacks.
type ComposeStackManager struct {
	dataPath             string
	reverseTunnelService portainer.ReverseTunnelService
}

// NewComposeStackManager initializes a new ComposeStackManager service.
func NewComposeStackManager(dataPath string, reverseTunnelService portainer.ReverseTunnelService) *ComposeStackManager {
	return &ComposeStackManager{
		dataPath:             dataPath,
		reverseTunnelService: reverseTunnelService,
	}
}

func (manager *ComposeStackManager) createClient(endpoint *portainer.Endpoint) (client.Factory, error) {

	endpointURL := endpoint.URL
	if endpoint.Type == portainer.EdgeAgentOnDockerEnvironment {
		tunnel := manager.reverseTunnelService.GetTunnelDetails(endpoint.ID)
		endpointURL = fmt.Sprintf("tcp://127.0.0.1:%d", tunnel.Port)
	}

	clientOpts := client.Options{
		Host:       endpointURL,
		APIVersion: dockerClientVersion,
	}

	if endpoint.TLSConfig.TLS {
		clientOpts.TLS = endpoint.TLSConfig.TLS
		clientOpts.TLSVerify = !endpoint.TLSConfig.TLSSkipVerify
		clientOpts.TLSCAFile = endpoint.TLSConfig.TLSCACertPath
		clientOpts.TLSCertFile = endpoint.TLSConfig.TLSCertPath
		clientOpts.TLSKeyFile = endpoint.TLSConfig.TLSKeyPath
	}

	return client.NewDefaultFactory(clientOpts)
}

// ComposeSyntaxMaxVersion returns the maximum supported version of the docker compose syntax
func (manager *ComposeStackManager) ComposeSyntaxMaxVersion() string {
	return composeSyntaxMaxVersion
}

// NormalizeStackName returns a new stack name with unsupported characters replaced
func (manager *ComposeStackManager) NormalizeStackName(name string) string {
	// this is coming from libcompose
	// https://github.com/portainer/libcompose/blob/master/project/context.go#L117-L120
	r := regexp.MustCompile("[^a-z0-9]+")
	return r.ReplaceAllString(strings.ToLower(name), "")
}

// Up will deploy a compose stack (equivalent of docker-compose up)
func (manager *ComposeStackManager) Up(stack *portainer.Stack, endpoint *portainer.Endpoint) error {

	clientFactory, err := manager.createClient(endpoint)
	if err != nil {
		return err
	}

	env := make(map[string]string)
	for _, envvar := range stack.Env {
		env[envvar.Name] = envvar.Value
	}
	filePaths := stackutils.GetStackFilePaths(stack)

	proj, err := docker.NewProject(&ctx.Context{
		ConfigDir: manager.dataPath,
		Context: project.Context{
			ComposeFiles: filePaths,
			EnvironmentLookup: &lookup.ComposableEnvLookup{
				Lookups: []config.EnvironmentLookup{
					&lookup.EnvfileLookup{
						Path: filepath.Join(stack.ProjectPath, ".env"),
					},
					&lookup.MapLookup{
						Vars: env,
					},
				},
			},
			ProjectName: stack.Name,
		},
		ClientFactory: clientFactory,
	}, nil)
	if err != nil {
		return err
	}

	return proj.Up(context.Background(), options.Up{})
}

// Down will shutdown a compose stack (equivalent of docker-compose down)
func (manager *ComposeStackManager) Down(stack *portainer.Stack, endpoint *portainer.Endpoint) error {
	clientFactory, err := manager.createClient(endpoint)
	if err != nil {
		return err
	}

	var composeFiles []string
	for _, file := range append([]string{stack.EntryPoint}, stack.AdditionalFiles...) {
		composeFiles = append(composeFiles, path.Join(stack.ProjectPath, file))
	}
	proj, err := docker.NewProject(&ctx.Context{
		Context: project.Context{
			ComposeFiles: composeFiles,
			ProjectName:  stack.Name,
		},
		ClientFactory: clientFactory,
	}, nil)
	if err != nil {
		return err
	}

	return proj.Down(context.Background(), options.Down{RemoveVolume: false, RemoveOrphans: true})
}

func stackFilePaths(stack *portainer.Stack) []string {
	var filePaths []string
	for _, file := range append([]string{stack.EntryPoint}, stack.AdditionalFiles...) {
		filePaths = append(filePaths, path.Join(stack.ProjectPath, file))
	}
	return filePaths
}

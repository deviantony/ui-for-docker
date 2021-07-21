package exec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/internal/stackutils"
)

// SwarmStackManager represents a service for managing stacks.
type SwarmStackManager struct {
	binaryPath           string
	dataPath             string
	signatureService     portainer.DigitalSignatureService
	fileService          portainer.FileService
	reverseTunnelService portainer.ReverseTunnelService
}

// NewSwarmStackManager initializes a new SwarmStackManager service.
// It also updates the configuration of the Docker CLI binary.
func NewSwarmStackManager(binaryPath, dataPath string, signatureService portainer.DigitalSignatureService, fileService portainer.FileService, reverseTunnelService portainer.ReverseTunnelService) (*SwarmStackManager, error) {
	manager := &SwarmStackManager{
		binaryPath:           binaryPath,
		dataPath:             dataPath,
		signatureService:     signatureService,
		fileService:          fileService,
		reverseTunnelService: reverseTunnelService,
	}

	err := manager.updateDockerCLIConfiguration(dataPath)
	if err != nil {
		return nil, err
	}

	return manager, nil
}

// Login executes the docker login command against a list of registries (including DockerHub).
func (manager *SwarmStackManager) Login(registries []portainer.Registry, endpoint *portainer.Endpoint) {
	command, args := manager.prepareDockerCommandAndArgs(manager.binaryPath, manager.dataPath, endpoint)
	for _, registry := range registries {
		if registry.Authentication {
			registryArgs := append(args, "login", "--username", registry.Username, "--password", registry.Password, registry.URL)
			runCommandAndCaptureStdErr(command, registryArgs, nil, "")
		}
	}
}

// Logout executes the docker logout command.
func (manager *SwarmStackManager) Logout(endpoint *portainer.Endpoint) error {
	command, args := manager.prepareDockerCommandAndArgs(manager.binaryPath, manager.dataPath, endpoint)
	args = append(args, "logout")
	return runCommandAndCaptureStdErr(command, args, nil, "")
}

// Deploy executes the docker stack deploy command.
func (manager *SwarmStackManager) Deploy(stack *portainer.Stack, prune bool, endpoint *portainer.Endpoint) error {
	filePaths := stackutils.GetStackFilePaths(stack)
	command, args := manager.prepareDockerCommandAndArgs(manager.binaryPath, manager.dataPath, endpoint)

	if prune {
		args = append(args, "stack", "deploy", "--prune", "--with-registry-auth")
	} else {
		args = append(args, "stack", "deploy", "--with-registry-auth")
	}

	args = configureFilePaths(args, filePaths)
	args = append(args, stack.Name)

	env := make([]string, 0)
	for _, envvar := range stack.Env {
		env = append(env, envvar.Name+"="+envvar.Value)
	}
	return runCommandAndCaptureStdErr(command, args, env, stack.ProjectPath)
}

// Remove executes the docker stack rm command.
func (manager *SwarmStackManager) Remove(stack *portainer.Stack, endpoint *portainer.Endpoint) error {
	command, args := manager.prepareDockerCommandAndArgs(manager.binaryPath, manager.dataPath, endpoint)
	args = append(args, "stack", "rm", stack.Name)
	return runCommandAndCaptureStdErr(command, args, nil, "")
}

func runCommandAndCaptureStdErr(command string, args []string, env []string, workingDir string) error {
	var stderr bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.Stderr = &stderr
	cmd.Dir = workingDir

	if env != nil {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, env...)
	}

	err := cmd.Run()
	if err != nil {
		return errors.New(stderr.String())
	}

	return nil
}

func (manager *SwarmStackManager) prepareDockerCommandAndArgs(binaryPath, dataPath string, endpoint *portainer.Endpoint) (string, []string) {
	// Assume Linux as a default
	command := path.Join(binaryPath, "docker")

	if runtime.GOOS == "windows" {
		command = path.Join(binaryPath, "docker.exe")
	}

	args := make([]string, 0)
	args = append(args, "--config", dataPath)

	endpointURL := endpoint.URL
	if endpoint.Type == portainer.EdgeAgentOnDockerEnvironment {
		tunnel := manager.reverseTunnelService.GetTunnelDetails(endpoint.ID)
		endpointURL = fmt.Sprintf("tcp://127.0.0.1:%d", tunnel.Port)
	}

	args = append(args, "-H", endpointURL)

	if endpoint.TLSConfig.TLS {
		args = append(args, "--tls")

		if !endpoint.TLSConfig.TLSSkipVerify {
			args = append(args, "--tlsverify", "--tlscacert", endpoint.TLSConfig.TLSCACertPath)
		} else {
			args = append(args, "--tlscacert", "''")
		}

		if endpoint.TLSConfig.TLSCertPath != "" && endpoint.TLSConfig.TLSKeyPath != "" {
			args = append(args, "--tlscert", endpoint.TLSConfig.TLSCertPath, "--tlskey", endpoint.TLSConfig.TLSKeyPath)
		}
	}

	return command, args
}

func (manager *SwarmStackManager) updateDockerCLIConfiguration(dataPath string) error {
	configFilePath := path.Join(dataPath, "config.json")
	config, err := manager.retrieveConfigurationFromDisk(configFilePath)
	if err != nil {
		return err
	}

	signature, err := manager.signatureService.CreateSignature(portainer.PortainerAgentSignatureMessage)
	if err != nil {
		return err
	}

	if config["HttpHeaders"] == nil {
		config["HttpHeaders"] = make(map[string]interface{})
	}
	headersObject := config["HttpHeaders"].(map[string]interface{})
	headersObject["X-PortainerAgent-ManagerOperation"] = "1"
	headersObject["X-PortainerAgent-Signature"] = signature
	headersObject["X-PortainerAgent-PublicKey"] = manager.signatureService.EncodedPublicKey()

	err = manager.fileService.WriteJSONToFile(configFilePath, config)
	if err != nil {
		return err
	}

	return nil
}

func (manager *SwarmStackManager) retrieveConfigurationFromDisk(path string) (map[string]interface{}, error) {
	var config map[string]interface{}

	raw, err := manager.fileService.GetFileContent(path)
	if err != nil {
		return make(map[string]interface{}), nil
	}

	err = json.Unmarshal(raw, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func configureFilePaths(args []string, filePaths []string) []string {
	for _, path := range filePaths {
		args = append(args, "--compose-file", path)
	}
	return args
}

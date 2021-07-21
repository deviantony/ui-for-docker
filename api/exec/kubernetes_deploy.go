package exec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/portainer/portainer/api/http/proxy/factory/kubernetes"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/kubernetes/cli"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/crypto"
)

// KubernetesDeployer represents a service to deploy resources inside a Kubernetes environment.
type KubernetesDeployer struct {
	binaryPath                  string
	dataStore                   portainer.DataStore
	reverseTunnelService        portainer.ReverseTunnelService
	signatureService            portainer.DigitalSignatureService
	kubernetesClientFactory     *cli.ClientFactory
	kubernetesTokenCacheManager *kubernetes.TokenCacheManager
}

// NewKubernetesDeployer initializes a new KubernetesDeployer service.
func NewKubernetesDeployer(kubernetesTokenCacheManager *kubernetes.TokenCacheManager, kubernetesClientFactory *cli.ClientFactory, datastore portainer.DataStore, reverseTunnelService portainer.ReverseTunnelService, signatureService portainer.DigitalSignatureService, binaryPath string) *KubernetesDeployer {
	return &KubernetesDeployer{
		binaryPath:                  binaryPath,
		dataStore:                   datastore,
		reverseTunnelService:        reverseTunnelService,
		signatureService:            signatureService,
		kubernetesClientFactory:     kubernetesClientFactory,
		kubernetesTokenCacheManager: kubernetesTokenCacheManager,
	}
}

func (deployer *KubernetesDeployer) getToken(request *http.Request, endpoint *portainer.Endpoint, setLocalAdminToken bool) (string, error) {
	tokenData, err := security.RetrieveTokenData(request)
	if err != nil {
		return "", err
	}

	kubecli, err := deployer.kubernetesClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return "", err
	}

	tokenCache := deployer.kubernetesTokenCacheManager.GetOrCreateTokenCache(int(endpoint.ID))

	tokenManager, err := kubernetes.NewTokenManager(kubecli, deployer.dataStore, tokenCache, setLocalAdminToken)
	if err != nil {
		return "", err
	}

	if tokenData.Role == portainer.AdministratorRole {
		return tokenManager.GetAdminServiceAccountToken(), nil
	}

	token, err := tokenManager.GetUserServiceAccountToken(int(tokenData.ID))
	if err != nil {
		return "", err
	}

	if token == "" {
		return "", fmt.Errorf("can not get a valid user service account token")
	}
	return token, nil
}

// Deploy will deploy a Kubernetes manifest inside a specific namespace in a Kubernetes endpoint.
// Otherwise it will use kubectl to deploy the manifest.
func (deployer *KubernetesDeployer) Deploy(request *http.Request, endpoint *portainer.Endpoint, stackConfig string, namespace string) (string, error) {
	if endpoint.Type == portainer.KubernetesLocalEnvironment {
		token, err := deployer.getToken(request, endpoint, true);
		if err != nil {
			return "", err
		}

		command := path.Join(deployer.binaryPath, "kubectl")
		if runtime.GOOS == "windows" {
			command = path.Join(deployer.binaryPath, "kubectl.exe")
		}

		args := make([]string, 0)
		args = append(args, "--server", endpoint.URL)
		args = append(args, "--insecure-skip-tls-verify")
		args = append(args, "--token", token)
		args = append(args, "--namespace", namespace)
		args = append(args, "apply", "-f", "-")

		var stderr bytes.Buffer
		cmd := exec.Command(command, args...)
		cmd.Stderr = &stderr
		cmd.Stdin = strings.NewReader(stackConfig)

		output, err := cmd.Output()
		if err != nil {
			return "", errors.New(stderr.String())
		}

		return string(output), nil
	}

	// agent

	endpointURL := endpoint.URL
	if endpoint.Type == portainer.EdgeAgentOnKubernetesEnvironment {
		tunnel := deployer.reverseTunnelService.GetTunnelDetails(endpoint.ID)
		if tunnel.Status == portainer.EdgeAgentIdle {

			err := deployer.reverseTunnelService.SetTunnelStatusToRequired(endpoint.ID)
			if err != nil {
				return "", err
			}

			settings, err := deployer.dataStore.Settings().Settings()
			if err != nil {
				return "", err
			}

			waitForAgentToConnect := time.Duration(settings.EdgeAgentCheckinInterval) * time.Second
			time.Sleep(waitForAgentToConnect * 2)
		}

		endpointURL = fmt.Sprintf("http://127.0.0.1:%d", tunnel.Port)
	}

	transport := &http.Transport{}

	if endpoint.TLSConfig.TLS {
		tlsConfig, err := crypto.CreateTLSConfigurationFromDisk(endpoint.TLSConfig.TLSCACertPath, endpoint.TLSConfig.TLSCertPath, endpoint.TLSConfig.TLSKeyPath, endpoint.TLSConfig.TLSSkipVerify)
		if err != nil {
			return "", err
		}
		transport.TLSClientConfig = tlsConfig
	}

	httpCli := &http.Client{
		Transport: transport,
	}

	if !strings.HasPrefix(endpointURL, "http") {
		endpointURL = fmt.Sprintf("https://%s", endpointURL)
	}

	url, err := url.Parse(fmt.Sprintf("%s/v2/kubernetes/stack", endpointURL))
	if err != nil {
		return "", err
	}

	reqPayload, err := json.Marshal(
		struct {
			StackConfig string
			Namespace   string
		}{
			StackConfig: stackConfig,
			Namespace:   namespace,
		})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewReader(reqPayload))
	if err != nil {
		return "", err
	}

	signature, err := deployer.signatureService.CreateSignature(portainer.PortainerAgentSignatureMessage)
	if err != nil {
		return "", err
	}

	token, err := deployer.getToken(request, endpoint, false);
	if err != nil {
		return "", err
	}

	req.Header.Set(portainer.PortainerAgentPublicKeyHeader, deployer.signatureService.EncodedPublicKey())
	req.Header.Set(portainer.PortainerAgentSignatureHeader, signature)
	req.Header.Set(portainer.PortainerAgentKubernetesSATokenHeader, token)

	resp, err := httpCli.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponseData struct {
			Message string
			Details string
		}
		err = json.NewDecoder(resp.Body).Decode(&errorResponseData)
		if err != nil {
			output, parseStringErr := ioutil.ReadAll(resp.Body)
			if parseStringErr != nil {
				return "", parseStringErr
			}

			return "", fmt.Errorf("Failed parsing, body: %s, error: %w", output, err)

		}

		return "", fmt.Errorf("Deployment to agent failed: %s", errorResponseData.Details)
	}

	var responseData struct{ Output string }
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		parsedOutput, parseStringErr := ioutil.ReadAll(resp.Body)
		if parseStringErr != nil {
			return "", parseStringErr
		}

		return "", fmt.Errorf("Failed decoding, body: %s, err: %w", parsedOutput, err)
	}

	return responseData.Output, nil

}

// ConvertCompose leverages the kompose binary to deploy a compose compliant manifest.
func (deployer *KubernetesDeployer) ConvertCompose(data string) ([]byte, error) {
	command := path.Join(deployer.binaryPath, "kompose")
	if runtime.GOOS == "windows" {
		command = path.Join(deployer.binaryPath, "kompose.exe")
	}

	args := make([]string, 0)
	args = append(args, "convert", "-f", "-", "--stdout")

	var stderr bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.Stderr = &stderr
	cmd.Stdin = strings.NewReader(data)

	output, err := cmd.Output()
	if err != nil {
		return nil, errors.New(stderr.String())
	}

	return output, nil
}

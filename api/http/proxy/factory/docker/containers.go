package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/docker/docker/client"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/proxy/factory/responseutils"
	"github.com/portainer/portainer/api/http/security"
)

const (
	containerObjectIdentifier = "Id"
)

func getInheritedResourceControlFromContainerLabels(dockerClient *client.Client, containerID string, resourceControls []portainer.ResourceControl) (*portainer.ResourceControl, error) {
	container, err := dockerClient.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return nil, err
	}

	serviceName := container.Config.Labels[resourceLabelForDockerServiceID]
	if serviceName != "" {
		serviceResourceControl := portainer.GetResourceControlByResourceIDAndType(serviceName, portainer.ServiceResourceControl, resourceControls)
		if serviceResourceControl != nil {
			return serviceResourceControl, nil
		}
	}

	swarmStackName := container.Config.Labels[resourceLabelForDockerSwarmStackName]
	if swarmStackName != "" {
		return portainer.GetResourceControlByResourceIDAndType(swarmStackName, portainer.StackResourceControl, resourceControls), nil
	}

	composeStackName := container.Config.Labels[resourceLabelForDockerComposeStackName]
	if composeStackName != "" {
		return portainer.GetResourceControlByResourceIDAndType(composeStackName, portainer.StackResourceControl, resourceControls), nil
	}

	return nil, nil
}

// containerListOperation extracts the response as a JSON array, loop through the containers array
// decorate and/or filter the containers based on resource controls before rewriting the response.
func (transport *Transport) containerListOperation(response *http.Response, executor *operationExecutor) error {
	// ContainerList response is a JSON array
	// https://docs.docker.com/engine/api/v1.28/#operation/ContainerList
	responseArray, err := responseutils.GetResponseAsJSONArray(response)
	if err != nil {
		return err
	}

	resourceOperationParameters := &resourceOperationParameters{
		resourceIdentifierAttribute: containerObjectIdentifier,
		resourceType:                portainer.ContainerResourceControl,
		labelsObjectSelector:        selectorContainerLabelsFromContainerListOperation,
	}

	responseArray, err = transport.applyAccessControlOnResourceList(resourceOperationParameters, responseArray, executor)
	if err != nil {
		return err
	}

	if executor.labelBlackList != nil {
		responseArray, err = filterContainersWithBlackListedLabels(responseArray, executor.labelBlackList)
		if err != nil {
			return err
		}
	}

	return responseutils.RewriteResponse(response, responseArray, http.StatusOK)
}

// containerInspectOperation extracts the response as a JSON object, verify that the user
// has access to the container based on resource control and either rewrite an access denied response or a decorated container.
func (transport *Transport) containerInspectOperation(response *http.Response, executor *operationExecutor) error {
	//ContainerInspect response is a JSON object
	// https://docs.docker.com/engine/api/v1.28/#operation/ContainerInspect
	responseObject, err := responseutils.GetResponseAsJSONOBject(response)
	if err != nil {
		return err
	}

	resourceOperationParameters := &resourceOperationParameters{
		resourceIdentifierAttribute: containerObjectIdentifier,
		resourceType:                portainer.ContainerResourceControl,
		labelsObjectSelector:        selectorContainerLabelsFromContainerInspectOperation,
	}

	return transport.applyAccessControlOnResource(resourceOperationParameters, responseObject, response, executor)
}

// selectorContainerLabelsFromContainerInspectOperation retrieve the labels object associated to the container object.
// This selector is specific to the containerInspect Docker operation.
// Labels are available under the "Config.Labels" property.
// API schema reference: https://docs.docker.com/engine/api/v1.28/#operation/ContainerInspect
func selectorContainerLabelsFromContainerInspectOperation(responseObject map[string]interface{}) map[string]interface{} {
	containerConfigObject := responseutils.GetJSONObject(responseObject, "Config")
	if containerConfigObject != nil {
		containerLabelsObject := responseutils.GetJSONObject(containerConfigObject, "Labels")
		return containerLabelsObject
	}
	return nil
}

// selectorContainerLabelsFromContainerListOperation retrieve the labels object associated to the container object.
// This selector is specific to the containerList Docker operation.
// Labels are available under the "Labels" property.
// API schema reference: https://docs.docker.com/engine/api/v1.28/#operation/ContainerList
func selectorContainerLabelsFromContainerListOperation(responseObject map[string]interface{}) map[string]interface{} {
	containerLabelsObject := responseutils.GetJSONObject(responseObject, "Labels")
	return containerLabelsObject
}

// filterContainersWithLabels loops through a list of containers, and filters containers that do not contains
// any labels in the labels black list.
func filterContainersWithBlackListedLabels(containerData []interface{}, labelBlackList []portainer.Pair) ([]interface{}, error) {
	filteredContainerData := make([]interface{}, 0)

	for _, container := range containerData {
		containerObject := container.(map[string]interface{})

		containerLabels := selectorContainerLabelsFromContainerListOperation(containerObject)
		if containerLabels != nil {
			if !containerHasBlackListedLabel(containerLabels, labelBlackList) {
				filteredContainerData = append(filteredContainerData, containerObject)
			}
		} else {
			filteredContainerData = append(filteredContainerData, containerObject)
		}
	}

	return filteredContainerData, nil
}

func containerHasBlackListedLabel(containerLabels map[string]interface{}, labelBlackList []portainer.Pair) bool {
	for key, value := range containerLabels {
		labelName := key
		labelValue := value.(string)

		for _, blackListedLabel := range labelBlackList {
			if blackListedLabel.Name == labelName && blackListedLabel.Value == labelValue {
				return true
			}
		}
	}

	return false
}

func (transport *Transport) decorateContainerCreationOperation(request *http.Request, resourceIdentifierAttribute string, resourceType portainer.ResourceControlType) (*http.Response, error) {
	type PartialContainer struct {
		HostConfig struct {
			Privileged bool          `json:"Privileged"`
			PidMode    string        `json:"PidMode"`
			Devices    []interface{} `json:"Devices"`
		} `json:"HostConfig"`
	}

	forbiddenResponse := &http.Response{
		StatusCode: http.StatusForbidden,
	}

	tokenData, err := security.RetrieveTokenData(request)
	if err != nil {
		return nil, err
	}

	user, err := transport.userService.User(tokenData.ID)
	if err != nil {
		return nil, err
	}

	rbacExtension, err := transport.extensionService.Extension(portainer.RBACExtension)
	if err != nil && err != portainer.ErrObjectNotFound {
		return nil, err
	}

	endpointResourceAccess := false
	_, ok := user.EndpointAuthorizations[portainer.EndpointID(transport.endpoint.ID)][portainer.EndpointResourcesAccess]
	if ok {
		endpointResourceAccess = true
	}

	isAdmin := (rbacExtension != nil && endpointResourceAccess) || tokenData.Role == portainer.AdministratorRole

	if !isAdmin {
		settings, err := transport.settingsService.Settings()
		if err != nil {
			return nil, err
		}

		if !settings.AllowPrivilegedModeForRegularUsers ||
			!settings.AllowHostNamespaceForRegularUsers ||
			settings.DisableDeviceMappingForRegularUsers {

			body, err := ioutil.ReadAll(request.Body)
			if err != nil {
				return nil, err
			}

			partialContainer := &PartialContainer{}
			err = json.Unmarshal(body, partialContainer)
			if err != nil {
				return nil, err
			}

			if partialContainer.HostConfig.Privileged {
				return forbiddenResponse, errors.New("forbidden to use privileged mode")
			}

			if partialContainer.HostConfig.PidMode == "host" {
				return forbiddenResponse, errors.New("forbidden to use pid host namespace")
			}

			if len(partialContainer.HostConfig.Devices) > 0 {
				return nil, errors.New("forbidden to use device mapping")
			}

			request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		}
	}

	response, err := transport.executeDockerRequest(request)
	if err != nil {
		return response, err
	}

	if response.StatusCode == http.StatusCreated {
		err = transport.decorateGenericResourceCreationResponse(response, resourceIdentifierAttribute, resourceType, tokenData.ID)
	}

	return response, err
}

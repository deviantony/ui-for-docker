package websocket

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portainer "github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
	"github.com/portainer/portainer/api/http/security"
)

// websocketShellPodExec handles GET requests on /websocket/pod?token=<token>&endpointId=<endpointID>
// The request will be upgraded to the websocket protocol.
// Authentication and access is controlled via the mandatory token query parameter.
// The request will proxy input from the client to the pod via long-lived websocket connection.
// The following query parameters are mandatory:
// * token: JWT token used for authentication against this endpoint
// * endpointId: endpoint ID of the endpoint where the resource is located
func (handler *Handler) websocketShellPodExec(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericQueryParameter(r, "endpointId", false)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid query parameter: endpointId", err}
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portainer.EndpointID(endpointID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find the endpoint associated to the stack inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find the endpoint associated to the stack inside the database", err}
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusForbidden, "Permission denied to access endpoint", err}
	}

	cli, err := handler.KubernetesClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to create Kubernetes client", err}
	}

	serviceAccount, err := cli.GetServiceAccount(tokenData)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find serviceaccount associated with user", err}
	}

	shellPod, err := cli.CreateUserShellPod(r.Context(), serviceAccount.Name)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to create user shell", err}
	}

	// Modifying request params mid-flight before forewarding to K8s API server (websocket)
	q := r.URL.Query()

	q.Add("namespace", shellPod.Namespace)
	q.Add("podName", shellPod.PodName)
	q.Add("containerName", shellPod.ContainerName)
	q.Add("command", shellPod.ShellExecCommand)

	r.URL.RawQuery = q.Encode()

	// Modify url path mid-flight before forewarding to k8s API server (websocket)
	r.URL.Path = "/websocket/pod"

	/*
		Note: The following websocket proxying logic is duplicated from `api/http/handler/websocket/pod.go`
	*/
	params := &webSocketRequestParams{
		endpoint: endpoint,
	}

	r.Header.Del("Origin")

	if endpoint.Type == portainer.AgentOnKubernetesEnvironment {
		err := handler.proxyAgentWebsocketRequest(w, r, params)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to proxy websocket request to agent", err}
		}
		return nil
	} else if endpoint.Type == portainer.EdgeAgentOnKubernetesEnvironment {
		err := handler.proxyEdgeAgentWebsocketRequest(w, r, params)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to proxy websocket request to Edge agent", err}
		}
		return nil
	}

	handlerErr := handler.hijackPodExecStartOperation(
		w,
		r,
		cli,
		endpoint,
		shellPod.Namespace,
		shellPod.PodName,
		shellPod.ContainerName,
		shellPod.ShellExecCommand,
	)
	if handlerErr != nil {
		return handlerErr
	}

	return nil
}

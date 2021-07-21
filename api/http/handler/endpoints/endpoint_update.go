package endpoints

import (
	"net/http"
	"reflect"
	"strconv"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/bolt/errors"
	"github.com/portainer/portainer/api/http/client"
	"github.com/portainer/portainer/api/internal/edge"
	"github.com/portainer/portainer/api/internal/tag"
)

type endpointUpdatePayload struct {
	// Name that will be used to identify this endpoint
	Name *string `example:"my-endpoint"`
	// URL or IP address of a Docker host
	URL *string `example:"docker.mydomain.tld:2375"`
	// URL or IP address where exposed containers will be reachable.\
	// Defaults to URL if not specified
	PublicURL *string `example:"docker.mydomain.tld:2375"`
	// Group identifier
	GroupID *int `example:"1"`
	// Require TLS to connect against this endpoint
	TLS *bool `example:"true"`
	// Skip server verification when using TLS
	TLSSkipVerify *bool `example:"false"`
	// Skip client verification when using TLS
	TLSSkipClientVerify *bool `example:"false"`
	// The status of the endpoint (1 - up, 2 - down)
	Status *int `example:"1"`
	// Azure application ID
	AzureApplicationID *string `example:"eag7cdo9-o09l-9i83-9dO9-f0b23oe78db4"`
	// Azure tenant ID
	AzureTenantID *string `example:"34ddc78d-4fel-2358-8cc1-df84c8o839f5"`
	// Azure authentication key
	AzureAuthenticationKey *string `example:"cOrXoK/1D35w8YQ8nH1/8ZGwzz45JIYD5jxHKXEQknk="`
	// List of tag identifiers to which this endpoint is associated
	TagIDs             []portainer.TagID `example:"1,2"`
	UserAccessPolicies portainer.UserAccessPolicies
	TeamAccessPolicies portainer.TeamAccessPolicies
	// The check in interval for edge agent (in seconds)
	EdgeCheckinInterval *int `example:"5"`
	// Associated Kubernetes data
	Kubernetes *portainer.KubernetesData
}

func (payload *endpointUpdatePayload) Validate(r *http.Request) error {
	return nil
}

// @id EndpointUpdate
// @summary Update an endpoint
// @description Update an endpoint.
// @description **Access policy**: administrator
// @security jwt
// @tags endpoints
// @accept json
// @produce json
// @param id path int true "Endpoint identifier"
// @param body body endpointUpdatePayload true "Endpoint details"
// @success 200 {object} portainer.Endpoint "Success"
// @failure 400 "Invalid request"
// @failure 404 "Endpoint not found"
// @failure 500 "Server error"
// @router /endpoints/{id} [put]
func (handler *Handler) endpointUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid endpoint identifier route variable", err}
	}

	var payload endpointUpdatePayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portainer.EndpointID(endpointID))
	if err == errors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an endpoint with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an endpoint with the specified identifier inside the database", err}
	}

	if payload.Name != nil {
		endpoint.Name = *payload.Name
	}

	if payload.URL != nil {
		endpoint.URL = *payload.URL
	}

	if payload.PublicURL != nil {
		endpoint.PublicURL = *payload.PublicURL
	}

	if payload.EdgeCheckinInterval != nil {
		endpoint.EdgeCheckinInterval = *payload.EdgeCheckinInterval
	}

	groupIDChanged := false
	if payload.GroupID != nil {
		groupID := portainer.EndpointGroupID(*payload.GroupID)
		groupIDChanged = groupID != endpoint.GroupID
		endpoint.GroupID = groupID
	}

	tagsChanged := false
	if payload.TagIDs != nil {
		payloadTagSet := tag.Set(payload.TagIDs)
		endpointTagSet := tag.Set((endpoint.TagIDs))
		union := tag.Union(payloadTagSet, endpointTagSet)
		intersection := tag.Intersection(payloadTagSet, endpointTagSet)
		tagsChanged = len(union) > len(intersection)

		if tagsChanged {
			removeTags := tag.Difference(endpointTagSet, payloadTagSet)

			for tagID := range removeTags {
				tag, err := handler.DataStore.Tag().Tag(tagID)
				if err != nil {
					return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find a tag inside the database", err}
				}

				delete(tag.Endpoints, endpoint.ID)
				err = handler.DataStore.Tag().UpdateTag(tag.ID, tag)
				if err != nil {
					return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist tag changes inside the database", err}
				}
			}

			endpoint.TagIDs = payload.TagIDs
			for _, tagID := range payload.TagIDs {
				tag, err := handler.DataStore.Tag().Tag(tagID)
				if err != nil {
					return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find a tag inside the database", err}
				}

				tag.Endpoints[endpoint.ID] = true

				err = handler.DataStore.Tag().UpdateTag(tag.ID, tag)
				if err != nil {
					return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist tag changes inside the database", err}
				}
			}
		}
	}

	updateAuthorizations := false

	if payload.Kubernetes != nil {
		if payload.Kubernetes.Configuration.RestrictDefaultNamespace !=
			endpoint.Kubernetes.Configuration.RestrictDefaultNamespace {
			updateAuthorizations = true
		}

		endpoint.Kubernetes = *payload.Kubernetes
	}

	if payload.UserAccessPolicies != nil && !reflect.DeepEqual(payload.UserAccessPolicies, endpoint.UserAccessPolicies) {
		updateAuthorizations = true
		endpoint.UserAccessPolicies = payload.UserAccessPolicies
	}

	if payload.TeamAccessPolicies != nil && !reflect.DeepEqual(payload.TeamAccessPolicies, endpoint.TeamAccessPolicies) {
		updateAuthorizations = true
		endpoint.TeamAccessPolicies = payload.TeamAccessPolicies
	}

	if payload.Status != nil {
		switch *payload.Status {
		case 1:
			endpoint.Status = portainer.EndpointStatusUp
			break
		case 2:
			endpoint.Status = portainer.EndpointStatusDown
			break
		default:
			break
		}
	}

	if endpoint.Type == portainer.AzureEnvironment {
		credentials := endpoint.AzureCredentials
		if payload.AzureApplicationID != nil {
			credentials.ApplicationID = *payload.AzureApplicationID
		}
		if payload.AzureTenantID != nil {
			credentials.TenantID = *payload.AzureTenantID
		}
		if payload.AzureAuthenticationKey != nil {
			credentials.AuthenticationKey = *payload.AzureAuthenticationKey
		}

		httpClient := client.NewHTTPClient()
		_, authErr := httpClient.ExecuteAzureAuthenticationRequest(&credentials)
		if authErr != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to authenticate against Azure", authErr}
		}
		endpoint.AzureCredentials = credentials
	}

	if payload.TLS != nil {
		folder := strconv.Itoa(endpointID)

		if *payload.TLS {
			endpoint.TLSConfig.TLS = true
			if payload.TLSSkipVerify != nil {
				endpoint.TLSConfig.TLSSkipVerify = *payload.TLSSkipVerify

				if !*payload.TLSSkipVerify {
					caCertPath, _ := handler.FileService.GetPathForTLSFile(folder, portainer.TLSFileCA)
					endpoint.TLSConfig.TLSCACertPath = caCertPath
				} else {
					endpoint.TLSConfig.TLSCACertPath = ""
					handler.FileService.DeleteTLSFile(folder, portainer.TLSFileCA)
				}
			}

			if payload.TLSSkipClientVerify != nil {
				if !*payload.TLSSkipClientVerify {
					certPath, _ := handler.FileService.GetPathForTLSFile(folder, portainer.TLSFileCert)
					endpoint.TLSConfig.TLSCertPath = certPath
					keyPath, _ := handler.FileService.GetPathForTLSFile(folder, portainer.TLSFileKey)
					endpoint.TLSConfig.TLSKeyPath = keyPath
				} else {
					endpoint.TLSConfig.TLSCertPath = ""
					handler.FileService.DeleteTLSFile(folder, portainer.TLSFileCert)
					endpoint.TLSConfig.TLSKeyPath = ""
					handler.FileService.DeleteTLSFile(folder, portainer.TLSFileKey)
				}
			}

		} else {
			endpoint.TLSConfig.TLS = false
			endpoint.TLSConfig.TLSSkipVerify = false
			endpoint.TLSConfig.TLSCACertPath = ""
			endpoint.TLSConfig.TLSCertPath = ""
			endpoint.TLSConfig.TLSKeyPath = ""
			err = handler.FileService.DeleteTLSFiles(folder)
			if err != nil {
				return &httperror.HandlerError{http.StatusInternalServerError, "Unable to remove TLS files from disk", err}
			}
		}

		if endpoint.Type == portainer.AgentOnKubernetesEnvironment || endpoint.Type == portainer.EdgeAgentOnKubernetesEnvironment {
			endpoint.TLSConfig.TLS = true
			endpoint.TLSConfig.TLSSkipVerify = true
		}
	}

	if payload.URL != nil || payload.TLS != nil || endpoint.Type == portainer.AzureEnvironment {
		_, err = handler.ProxyManager.CreateAndRegisterEndpointProxy(endpoint)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to register HTTP proxy for the endpoint", err}
		}
	}

	if updateAuthorizations {
		if endpoint.Type == portainer.KubernetesLocalEnvironment || endpoint.Type == portainer.AgentOnKubernetesEnvironment || endpoint.Type == portainer.EdgeAgentOnKubernetesEnvironment {
			err = handler.AuthorizationService.CleanNAPWithOverridePolicies(endpoint, nil)
			if err != nil {
				return &httperror.HandlerError{http.StatusInternalServerError, "Unable to update user authorizations", err}
			}
		}
	}

	err = handler.DataStore.Endpoint().UpdateEndpoint(endpoint.ID, endpoint)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist endpoint changes inside the database", err}
	}

	if (endpoint.Type == portainer.EdgeAgentOnDockerEnvironment || endpoint.Type == portainer.EdgeAgentOnKubernetesEnvironment) && (groupIDChanged || tagsChanged) {
		relation, err := handler.DataStore.EndpointRelation().EndpointRelation(endpoint.ID)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find endpoint relation inside the database", err}
		}

		endpointGroup, err := handler.DataStore.EndpointGroup().EndpointGroup(endpoint.GroupID)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find endpoint group inside the database", err}
		}

		edgeGroups, err := handler.DataStore.EdgeGroup().EdgeGroups()
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve edge groups from the database", err}
		}

		edgeStacks, err := handler.DataStore.EdgeStack().EdgeStacks()
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve edge stacks from the database", err}
		}

		edgeStackSet := map[portainer.EdgeStackID]bool{}

		endpointEdgeStacks := edge.EndpointRelatedEdgeStacks(endpoint, endpointGroup, edgeGroups, edgeStacks)
		for _, edgeStackID := range endpointEdgeStacks {
			edgeStackSet[edgeStackID] = true
		}

		relation.EdgeStacks = edgeStackSet

		err = handler.DataStore.EndpointRelation().UpdateEndpointRelation(endpoint.ID, relation)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist endpoint relation changes inside the database", err}
		}
	}

	return response.JSON(w, endpoint)
}

package portainer

import (
	"io"
	"time"

	gittypes "github.com/portainer/portainer/api/git/types"
)

type (
	// AccessPolicy represent a policy that can be associated to a user or team
	AccessPolicy struct {
		// Role identifier. Reference the role that will be associated to this access policy
		RoleID RoleID `json:"RoleId" example:"1"`
	}

	// AgentPlatform represents a platform type for an Agent
	AgentPlatform int

	// AuthenticationMethod represents the authentication method used to authenticate a user
	AuthenticationMethod int

	// Authorization represents an authorization associated to an operation
	Authorization string

	// Authorizations represents a set of authorizations associated to a role
	Authorizations map[Authorization]bool

	// AzureCredentials represents the credentials used to connect to an Azure
	// environment.
	AzureCredentials struct {
		// Azure application ID
		ApplicationID string `json:"ApplicationID" example:"eag7cdo9-o09l-9i83-9dO9-f0b23oe78db4"`
		// Azure tenant ID
		TenantID string `json:"TenantID" example:"34ddc78d-4fel-2358-8cc1-df84c8o839f5"`
		// Azure authentication key
		AuthenticationKey string `json:"AuthenticationKey" example:"cOrXoK/1D35w8YQ8nH1/8ZGwzz45JIYD5jxHKXEQknk="`
	}

	// CLIFlags represents the available flags on the CLI
	CLIFlags struct {
		Addr                      *string
		AddrHTTPS                 *string
		TunnelAddr                *string
		TunnelPort                *string
		AdminPassword             *string
		AdminPasswordFile         *string
		Assets                    *string
		Data                      *string
		EnableEdgeComputeFeatures *bool
		EndpointURL               *string
		Labels                    *[]Pair
		Logo                      *string
		NoAnalytics               *bool
		Templates                 *string
		TLS                       *bool
		TLSSkipVerify             *bool
		TLSCacert                 *string
		TLSCert                   *string
		TLSKey                    *string
		HTTPDisabled              *bool
		SSL                       *bool
		SSLCert                   *string
		SSLKey                    *string
		SnapshotInterval          *string
	}

	// CustomTemplate represents a custom template
	CustomTemplate struct {
		// CustomTemplate Identifier
		ID CustomTemplateID `json:"Id" example:"1"`
		// Title of the template
		Title string `json:"Title" example:"Nginx"`
		// Description of the template
		Description string `json:"Description" example:"High performance web server"`
		// Path on disk to the repository hosting the Stack file
		ProjectPath string `json:"ProjectPath" example:"/data/custom_template/3"`
		// Path to the Stack file
		EntryPoint string `json:"EntryPoint" example:"docker-compose.yml"`
		// User identifier who created this template
		CreatedByUserID UserID `json:"CreatedByUserId" example:"3"`
		// A note that will be displayed in the UI. Supports HTML content
		Note string `json:"Note" example:"This is my <b>custom</b> template"`
		// Platform associated to the template.
		// Valid values are: 1 - 'linux', 2 - 'windows'
		Platform CustomTemplatePlatform `json:"Platform" example:"1" enums:"1,2"`
		// URL of the template's logo
		Logo string `json:"Logo" example:"https://cloudinovasi.id/assets/img/logos/nginx.png"`
		// Type of created stack (1 - swarm, 2 - compose)
		Type            StackType        `json:"Type" example:"1"`
		ResourceControl *ResourceControl `json:"ResourceControl"`
	}

	// CustomTemplateID represents a custom template identifier
	CustomTemplateID int

	// CustomTemplatePlatform represents a custom template platform
	CustomTemplatePlatform int

	// DockerHub represents all the required information to connect and use the
	// Docker Hub
	DockerHub struct {
		// Is authentication against DockerHub enabled
		Authentication bool `json:"Authentication" example:"true"`
		// Username used to authenticate against the DockerHub
		Username string `json:"Username" example:"user"`
		// Password used to authenticate against the DockerHub
		Password string `json:"Password,omitempty" example:"passwd"`
	}

	// DockerSnapshot represents a snapshot of a specific Docker endpoint at a specific time
	DockerSnapshot struct {
		Time                    int64             `json:"Time"`
		DockerVersion           string            `json:"DockerVersion"`
		Swarm                   bool              `json:"Swarm"`
		TotalCPU                int               `json:"TotalCPU"`
		TotalMemory             int64             `json:"TotalMemory"`
		RunningContainerCount   int               `json:"RunningContainerCount"`
		StoppedContainerCount   int               `json:"StoppedContainerCount"`
		HealthyContainerCount   int               `json:"HealthyContainerCount"`
		UnhealthyContainerCount int               `json:"UnhealthyContainerCount"`
		VolumeCount             int               `json:"VolumeCount"`
		ImageCount              int               `json:"ImageCount"`
		ServiceCount            int               `json:"ServiceCount"`
		StackCount              int               `json:"StackCount"`
		NodeCount               int               `json:"NodeCount"`
		SnapshotRaw             DockerSnapshotRaw `json:"DockerSnapshotRaw"`
	}

	// DockerSnapshotRaw represents all the information related to a snapshot as returned by the Docker API
	DockerSnapshotRaw struct {
		Containers interface{} `json:"Containers"`
		Volumes    interface{} `json:"Volumes"`
		Networks   interface{} `json:"Networks"`
		Images     interface{} `json:"Images"`
		Info       interface{} `json:"Info"`
		Version    interface{} `json:"Version"`
	}

	// EdgeGroup represents an Edge group
	EdgeGroup struct {
		// EdgeGroup Identifier
		ID           EdgeGroupID  `json:"Id" example:"1"`
		Name         string       `json:"Name"`
		Dynamic      bool         `json:"Dynamic"`
		TagIDs       []TagID      `json:"TagIds"`
		Endpoints    []EndpointID `json:"Endpoints"`
		PartialMatch bool         `json:"PartialMatch"`
	}

	// EdgeGroupID represents an Edge group identifier
	EdgeGroupID int

	// EdgeJob represents a job that can run on Edge environments.
	EdgeJob struct {
		// EdgeJob Identifier
		ID             EdgeJobID                          `json:"Id" example:"1"`
		Created        int64                              `json:"Created"`
		CronExpression string                             `json:"CronExpression"`
		Endpoints      map[EndpointID]EdgeJobEndpointMeta `json:"Endpoints"`
		Name           string                             `json:"Name"`
		ScriptPath     string                             `json:"ScriptPath"`
		Recurring      bool                               `json:"Recurring"`
		Version        int                                `json:"Version"`
	}

	// EdgeJobEndpointMeta represents a meta data object for an Edge job and Endpoint relation
	EdgeJobEndpointMeta struct {
		LogsStatus  EdgeJobLogsStatus
		CollectLogs bool
	}

	// EdgeJobID represents an Edge job identifier
	EdgeJobID int

	// EdgeJobLogsStatus represent status of logs collection job
	EdgeJobLogsStatus int

	// EdgeSchedule represents a scheduled job that can run on Edge environments.
	// Deprecated in favor of EdgeJob
	EdgeSchedule struct {
		// EdgeSchedule Identifier
		ID             ScheduleID   `json:"Id" example:"1"`
		CronExpression string       `json:"CronExpression"`
		Script         string       `json:"Script"`
		Version        int          `json:"Version"`
		Endpoints      []EndpointID `json:"Endpoints"`
	}

	//EdgeStack represents an edge stack
	EdgeStack struct {
		// EdgeStack Identifier
		ID           EdgeStackID                    `json:"Id" example:"1"`
		Name         string                         `json:"Name"`
		Status       map[EndpointID]EdgeStackStatus `json:"Status"`
		CreationDate int64                          `json:"CreationDate"`
		EdgeGroups   []EdgeGroupID                  `json:"EdgeGroups"`
		ProjectPath  string                         `json:"ProjectPath"`
		EntryPoint   string                         `json:"EntryPoint"`
		Version      int                            `json:"Version"`
		Prune        bool                           `json:"Prune"`
	}

	//EdgeStackID represents an edge stack id
	EdgeStackID int

	//EdgeStackStatus represents an edge stack status
	EdgeStackStatus struct {
		Type       EdgeStackStatusType `json:"Type"`
		Error      string              `json:"Error"`
		EndpointID EndpointID          `json:"EndpointID"`
	}

	//EdgeStackStatusType represents an edge stack status type
	EdgeStackStatusType int

	// Endpoint represents a Docker endpoint with all the info required
	// to connect to it
	Endpoint struct {
		// Endpoint Identifier
		ID EndpointID `json:"Id" example:"1"`
		// Endpoint name
		Name string `json:"Name" example:"my-endpoint"`
		// Endpoint environment type. 1 for a Docker environment, 2 for an agent on Docker environment or 3 for an Azure environment.
		Type EndpointType `json:"Type" example:"1"`
		// URL or IP address of the Docker host associated to this endpoint
		URL string `json:"URL" example:"docker.mydomain.tld:2375"`
		// Endpoint group identifier
		GroupID EndpointGroupID `json:"GroupId" example:"1"`
		// URL or IP address where exposed containers will be reachable
		PublicURL        string              `json:"PublicURL" example:"docker.mydomain.tld:2375"`
		TLSConfig        TLSConfiguration    `json:"TLSConfig"`
		Extensions       []EndpointExtension `json:"Extensions" example:""`
		AzureCredentials AzureCredentials    `json:"AzureCredentials,omitempty" example:""`
		// List of tag identifiers to which this endpoint is associated
		TagIDs []TagID `json:"TagIds"`
		// The status of the endpoint (1 - up, 2 - down)
		Status EndpointStatus `json:"Status" example:"1"`
		// List of snapshots
		Snapshots []DockerSnapshot `json:"Snapshots" example:""`
		// List of user identifiers authorized to connect to this endpoint
		UserAccessPolicies UserAccessPolicies `json:"UserAccessPolicies"`
		// List of team identifiers authorized to connect to this endpoint
		TeamAccessPolicies TeamAccessPolicies `json:"TeamAccessPolicies" example:""`
		// The identifier of the edge agent associated with this endpoint
		EdgeID string `json:"EdgeID,omitempty" example:""`
		// The key which is used to map the agent to Portainer
		EdgeKey string `json:"EdgeKey" example:""`
		// The check in interval for edge agent (in seconds)
		EdgeCheckinInterval int `json:"EdgeCheckinInterval" example:"5"`
		// Associated Kubernetes data
		Kubernetes KubernetesData `json:"Kubernetes" example:""`
		// Maximum version of docker-compose
		ComposeSyntaxMaxVersion string `json:"ComposeSyntaxMaxVersion" example:"3.8"`
		// Endpoint specific security settings
		SecuritySettings EndpointSecuritySettings
		// LastCheckInDate mark last check-in date on checkin
		LastCheckInDate int64

		// Deprecated fields
		// Deprecated in DBVersion == 4
		TLS           bool   `json:"TLS,omitempty"`
		TLSCACertPath string `json:"TLSCACert,omitempty"`
		TLSCertPath   string `json:"TLSCert,omitempty"`
		TLSKeyPath    string `json:"TLSKey,omitempty"`

		// Deprecated in DBVersion == 18
		AuthorizedUsers []UserID `json:"AuthorizedUsers"`
		AuthorizedTeams []TeamID `json:"AuthorizedTeams"`

		// Deprecated in DBVersion == 22
		Tags []string `json:"Tags"`
	}

	// EndpointAuthorizations represents the authorizations associated to a set of endpoints
	EndpointAuthorizations map[EndpointID]Authorizations

	// EndpointExtension represents a deprecated form of Portainer extension
	// TODO: legacy extension management
	EndpointExtension struct {
		Type EndpointExtensionType `json:"Type"`
		URL  string                `json:"URL"`
	}

	// EndpointExtensionType represents the type of an endpoint extension. Only
	// one extension of each type can be associated to an endpoint
	EndpointExtensionType int

	// EndpointGroup represents a group of endpoints
	EndpointGroup struct {
		// Endpoint group Identifier
		ID EndpointGroupID `json:"Id" example:"1"`
		// Endpoint group name
		Name string `json:"Name" example:"my-endpoint-group"`
		// Description associated to the endpoint group
		Description        string             `json:"Description" example:"Endpoint group description"`
		UserAccessPolicies UserAccessPolicies `json:"UserAccessPolicies" example:""`
		TeamAccessPolicies TeamAccessPolicies `json:"TeamAccessPolicies" example:""`
		// List of tags associated to this endpoint group
		TagIDs []TagID `json:"TagIds"`

		// Deprecated fields
		Labels []Pair `json:"Labels"`

		// Deprecated in DBVersion == 18
		AuthorizedUsers []UserID `json:"AuthorizedUsers"`
		AuthorizedTeams []TeamID `json:"AuthorizedTeams"`

		// Deprecated in DBVersion == 22
		Tags []string `json:"Tags"`
	}

	// EndpointGroupID represents an endpoint group identifier
	EndpointGroupID int

	// EndpointID represents an endpoint identifier
	EndpointID int

	// EndpointStatus represents the status of an endpoint
	EndpointStatus int

	// EndpointSyncJob represents a scheduled job that synchronize endpoints based on an external file
	// Deprecated
	EndpointSyncJob struct{}

	// EndpointSecuritySettings represents settings for an endpoint
	EndpointSecuritySettings struct {
		// Whether non-administrator should be able to use bind mounts when creating containers
		AllowBindMountsForRegularUsers bool `json:"allowBindMountsForRegularUsers" example:"false"`
		// Whether non-administrator should be able to use privileged mode when creating containers
		AllowPrivilegedModeForRegularUsers bool `json:"allowPrivilegedModeForRegularUsers" example:"false"`
		// Whether non-administrator should be able to browse volumes
		AllowVolumeBrowserForRegularUsers bool `json:"allowVolumeBrowserForRegularUsers" example:"true"`
		// Whether non-administrator should be able to use the host pid
		AllowHostNamespaceForRegularUsers bool `json:"allowHostNamespaceForRegularUsers" example:"true"`
		// Whether non-administrator should be able to use device mapping
		AllowDeviceMappingForRegularUsers bool `json:"allowDeviceMappingForRegularUsers" example:"true"`
		// Whether non-administrator should be able to manage stacks
		AllowStackManagementForRegularUsers bool `json:"allowStackManagementForRegularUsers" example:"true"`
		// Whether non-administrator should be able to use container capabilities
		AllowContainerCapabilitiesForRegularUsers bool `json:"allowContainerCapabilitiesForRegularUsers" example:"true"`
		// Whether non-administrator should be able to use sysctl settings
		AllowSysctlSettingForRegularUsers bool `json:"allowSysctlSettingForRegularUsers" example:"true"`
		// Whether host management features are enabled
		EnableHostManagementFeatures bool `json:"enableHostManagementFeatures" example:"true"`
	}

	// EndpointType represents the type of an endpoint
	EndpointType int

	// EndpointRelation represents a endpoint relation object
	EndpointRelation struct {
		EndpointID EndpointID
		EdgeStacks map[EdgeStackID]bool
	}

	// Extension represents a deprecated Portainer extension
	Extension struct {
		// Extension Identifier
		ID               ExtensionID        `json:"Id" example:"1"`
		Enabled          bool               `json:"Enabled"`
		Name             string             `json:"Name,omitempty"`
		ShortDescription string             `json:"ShortDescription,omitempty"`
		Description      string             `json:"Description,omitempty"`
		DescriptionURL   string             `json:"DescriptionURL,omitempty"`
		Price            string             `json:"Price,omitempty"`
		PriceDescription string             `json:"PriceDescription,omitempty"`
		Deal             bool               `json:"Deal,omitempty"`
		Available        bool               `json:"Available,omitempty"`
		License          LicenseInformation `json:"License,omitempty"`
		Version          string             `json:"Version"`
		UpdateAvailable  bool               `json:"UpdateAvailable"`
		ShopURL          string             `json:"ShopURL,omitempty"`
		Images           []string           `json:"Images,omitempty"`
		Logo             string             `json:"Logo,omitempty"`
	}

	// ExtensionID represents a extension identifier
	ExtensionID int

	// GitlabRegistryData represents data required for gitlab registry to work
	GitlabRegistryData struct {
		ProjectID   int    `json:"ProjectId"`
		InstanceURL string `json:"InstanceURL"`
		ProjectPath string `json:"ProjectPath"`
	}

	// QuayRegistryData represents data required for Quay registry to work
	QuayRegistryData struct {
		UseOrganisation  bool   `json:"UseOrganisation"`
		OrganisationName string `json:"OrganisationName"`
	}

	// JobType represents a job type
	JobType int

	K8sNamespaceAccessPolicy struct {
		UserAccessPolicies UserAccessPolicies `json:"UserAccessPolicies"`
		TeamAccessPolicies TeamAccessPolicies `json:"TeamAccessPolicies"`
	}

	// KubernetesData contains all the Kubernetes related endpoint information
	KubernetesData struct {
		Snapshots     []KubernetesSnapshot    `json:"Snapshots"`
		Configuration KubernetesConfiguration `json:"Configuration"`
	}

	// KubernetesSnapshot represents a snapshot of a specific Kubernetes endpoint at a specific time
	KubernetesSnapshot struct {
		Time              int64  `json:"Time"`
		KubernetesVersion string `json:"KubernetesVersion"`
		NodeCount         int    `json:"NodeCount"`
		TotalCPU          int64  `json:"TotalCPU"`
		TotalMemory       int64  `json:"TotalMemory"`
	}

	// KubernetesConfiguration represents the configuration of a Kubernetes endpoint
	KubernetesConfiguration struct {
		UseLoadBalancer  bool                           `json:"UseLoadBalancer"`
		UseServerMetrics bool                           `json:"UseServerMetrics"`
		StorageClasses   []KubernetesStorageClassConfig `json:"StorageClasses"`
		IngressClasses   []KubernetesIngressClassConfig `json:"IngressClasses"`
	}

	// KubernetesStorageClassConfig represents a Kubernetes Storage Class configuration
	KubernetesStorageClassConfig struct {
		Name                 string   `json:"Name"`
		AccessModes          []string `json:"AccessModes"`
		Provisioner          string   `json:"Provisioner"`
		AllowVolumeExpansion bool     `json:"AllowVolumeExpansion"`
	}

	// KubernetesIngressClassConfig represents a Kubernetes Ingress Class configuration
	KubernetesIngressClassConfig struct {
		Name string `json:"Name"`
		Type string `json:"Type"`
	}

	// LDAPGroupSearchSettings represents settings used to search for groups in a LDAP server
	LDAPGroupSearchSettings struct {
		// The distinguished name of the element from which the LDAP server will search for groups
		GroupBaseDN string `json:"GroupBaseDN" example:"dc=ldap,dc=domain,dc=tld"`
		// The LDAP search filter used to select group elements, optional
		GroupFilter string `json:"GroupFilter" example:"(objectClass=account"`
		// LDAP attribute which denotes the group membership
		GroupAttribute string `json:"GroupAttribute" example:"member"`
	}

	// LDAPSearchSettings represents settings used to search for users in a LDAP server
	LDAPSearchSettings struct {
		// The distinguished name of the element from which the LDAP server will search for users
		BaseDN string `json:"BaseDN" example:"dc=ldap,dc=domain,dc=tld"`
		// Optional LDAP search filter used to select user elements
		Filter string `json:"Filter" example:"(objectClass=account)"`
		// LDAP attribute which denotes the username
		UserNameAttribute string `json:"UserNameAttribute" example:"uid"`
	}

	// LDAPSettings represents the settings used to connect to a LDAP server
	LDAPSettings struct {
		// Enable this option if the server is configured for Anonymous access. When enabled, ReaderDN and Password will not be used
		AnonymousMode bool `json:"AnonymousMode" example:"true"`
		// Account that will be used to search for users
		ReaderDN string `json:"ReaderDN" example:"cn=readonly-account,dc=ldap,dc=domain,dc=tld"`
		// Password of the account that will be used to search users
		Password string `json:"Password,omitempty" example:"readonly-password"`
		// URL or IP address of the LDAP server
		URL       string           `json:"URL" example:"myldap.domain.tld:389"`
		TLSConfig TLSConfiguration `json:"TLSConfig"`
		// Whether LDAP connection should use StartTLS
		StartTLS            bool                      `json:"StartTLS" example:"true"`
		SearchSettings      []LDAPSearchSettings      `json:"SearchSettings"`
		GroupSearchSettings []LDAPGroupSearchSettings `json:"GroupSearchSettings"`
		// Automatically provision users and assign them to matching LDAP group names
		AutoCreateUsers bool `json:"AutoCreateUsers" example:"true"`
	}

	// LicenseInformation represents information about an extension license
	LicenseInformation struct {
		LicenseKey string `json:"LicenseKey,omitempty"`
		Company    string `json:"Company,omitempty"`
		Expiration string `json:"Expiration,omitempty"`
		Valid      bool   `json:"Valid,omitempty"`
	}

	// MembershipRole represents the role of a user within a team
	MembershipRole int

	// OAuthSettings represents the settings used to authorize with an authorization server
	OAuthSettings struct {
		ClientID             string `json:"ClientID"`
		ClientSecret         string `json:"ClientSecret,omitempty"`
		AccessTokenURI       string `json:"AccessTokenURI"`
		AuthorizationURI     string `json:"AuthorizationURI"`
		ResourceURI          string `json:"ResourceURI"`
		RedirectURI          string `json:"RedirectURI"`
		UserIdentifier       string `json:"UserIdentifier"`
		Scopes               string `json:"Scopes"`
		OAuthAutoCreateUsers bool   `json:"OAuthAutoCreateUsers"`
		DefaultTeamID        TeamID `json:"DefaultTeamID"`
		SSO                  bool   `json:"SSO"`
		LogoutURI            string `json:"LogoutURI"`
	}

	// Pair defines a key/value string pair
	Pair struct {
		Name  string `json:"name" example:"name"`
		Value string `json:"value" example:"value"`
	}

	// Registry represents a Docker registry with all the info required
	// to connect to it
	Registry struct {
		// Registry Identifier
		ID RegistryID `json:"Id" example:"1"`
		// Registry Type (1 - Quay, 2 - Azure, 3 - Custom, 4 - Gitlab, 5 - ProGet, 6 - DockerHub)
		Type RegistryType `json:"Type" enums:"1,2,3,4,5,6"`
		// Registry Name
		Name string `json:"Name" example:"my-registry"`
		// URL or IP address of the Docker registry
		URL string `json:"URL" example:"registry.mydomain.tld:2375"`
		// Base URL, introduced for ProGet registry
		BaseURL string `json:"BaseURL" example:"registry.mydomain.tld:2375"`
		// Is authentication against this registry enabled
		Authentication bool `json:"Authentication" example:"true"`
		// Username used to authenticate against this registry
		Username string `json:"Username" example:"registry user"`
		// Password used to authenticate against this registry
		Password                string                           `json:"Password,omitempty" example:"registry_password"`
		ManagementConfiguration *RegistryManagementConfiguration `json:"ManagementConfiguration"`
		Gitlab                  GitlabRegistryData               `json:"Gitlab"`
		Quay                    QuayRegistryData                 `json:"Quay"`
		RegistryAccesses        RegistryAccesses                 `json:"RegistryAccesses"`

		// Deprecated fields
		// Deprecated in DBVersion == 31
		UserAccessPolicies UserAccessPolicies `json:"UserAccessPolicies"`
		// Deprecated in DBVersion == 31
		TeamAccessPolicies TeamAccessPolicies `json:"TeamAccessPolicies"`

		// Deprecated in DBVersion == 18
		AuthorizedUsers []UserID `json:"AuthorizedUsers"`
		// Deprecated in DBVersion == 18
		AuthorizedTeams []TeamID `json:"AuthorizedTeams"`
	}

	RegistryAccesses map[EndpointID]RegistryAccessPolicies

	RegistryAccessPolicies struct {
		UserAccessPolicies UserAccessPolicies `json:"UserAccessPolicies"`
		TeamAccessPolicies TeamAccessPolicies `json:"TeamAccessPolicies"`
		Namespaces         []string           `json:"Namespaces"`
	}

	// RegistryID represents a registry identifier
	RegistryID int

	// RegistryManagementConfiguration represents a configuration that can be used to query
	// the registry API via the registry management extension.
	RegistryManagementConfiguration struct {
		Type           RegistryType     `json:"Type"`
		Authentication bool             `json:"Authentication"`
		Username       string           `json:"Username"`
		Password       string           `json:"Password"`
		TLSConfig      TLSConfiguration `json:"TLSConfig"`
	}

	// RegistryType represents a type of registry
	RegistryType int

	// ResourceAccessLevel represents the level of control associated to a resource
	ResourceAccessLevel int

	// ResourceControl represent a reference to a Docker resource with specific access controls
	ResourceControl struct {
		// ResourceControl Identifier
		ID ResourceControlID `json:"Id" example:"1"`
		// Docker resource identifier on which access control will be applied.\
		// In the case of a resource control applied to a stack, use the stack name as identifier
		ResourceID string `json:"ResourceId" example:"617c5f22bb9b023d6daab7cba43a57576f83492867bc767d1c59416b065e5f08"`
		// List of Docker resources that will inherit this access control
		SubResourceIDs []string `json:"SubResourceIds" example:"617c5f22bb9b023d6daab7cba43a57576f83492867bc767d1c59416b065e5f08"`
		// Type of Docker resource. Valid values are: 1- container, 2 -service
		// 3 - volume, 4 - secret, 5 - stack, 6 - config or 7 - custom template
		Type         ResourceControlType  `json:"Type" example:"1"`
		UserAccesses []UserResourceAccess `json:"UserAccesses" example:""`
		TeamAccesses []TeamResourceAccess `json:"TeamAccesses" example:""`
		// Permit access to the associated resource to any user
		Public bool `json:"Public" example:"true"`
		// Permit access to resource only to admins
		AdministratorsOnly bool `json:"AdministratorsOnly" example:"true"`
		System             bool `json:"System" example:""`

		// Deprecated fields
		// Deprecated in DBVersion == 2
		OwnerID     UserID              `json:"OwnerId,omitempty"`
		AccessLevel ResourceAccessLevel `json:"AccessLevel,omitempty"`
	}

	// ResourceControlID represents a resource control identifier
	ResourceControlID int

	// ResourceControlType represents the type of resource associated to the resource control (volume, container, service...)
	ResourceControlType int

	// Role represents a set of authorizations that can be associated to a user or
	// to a team.
	Role struct {
		// Role Identifier
		ID RoleID `json:"Id" example:"1"`
		// Role name
		Name string `json:"Name" example:"HelpDesk"`
		// Role description
		Description string `json:"Description" example:"Read-only access of all resources in an endpoint"`
		// Authorizations associated to a role
		Authorizations Authorizations `json:"Authorizations"`
		Priority       int            `json:"Priority"`
	}

	// RoleID represents a role identifier
	RoleID int

	// Schedule represents a scheduled job.
	// It only contains a pointer to one of the JobRunner implementations
	// based on the JobType.
	// NOTE: The Recurring option is only used by ScriptExecutionJob at the moment
	// Deprecated in favor of EdgeJob
	Schedule struct {
		// Schedule Identifier
		ID             ScheduleID `json:"Id" example:"1"`
		Name           string
		CronExpression string
		Recurring      bool
		Created        int64
		JobType        JobType
		EdgeSchedule   *EdgeSchedule
	}

	// ScheduleID represents a schedule identifier.
	// Deprecated in favor of EdgeJob
	ScheduleID int

	// ScriptExecutionJob represents a scheduled job that can execute a script via a privileged container
	ScriptExecutionJob struct {
		Endpoints     []EndpointID
		Image         string
		ScriptPath    string
		RetryCount    int
		RetryInterval int
	}

	// Settings represents the application settings
	Settings struct {
		// URL to a logo that will be displayed on the login page as well as on top of the sidebar. Will use default Portainer logo when value is empty string
		LogoURL string `json:"LogoURL" example:"https://mycompany.mydomain.tld/logo.png"`
		// A list of label name & value that will be used to hide containers when querying containers
		BlackListedLabels []Pair `json:"BlackListedLabels"`
		// Active authentication method for the Portainer instance. Valid values are: 1 for internal, 2 for LDAP, or 3 for oauth
		AuthenticationMethod AuthenticationMethod `json:"AuthenticationMethod" example:"1"`
		LDAPSettings         LDAPSettings         `json:"LDAPSettings" example:""`
		OAuthSettings        OAuthSettings        `json:"OAuthSettings" example:""`
		// The interval in which endpoint snapshots are created
		SnapshotInterval string `json:"SnapshotInterval" example:"5m"`
		// URL to the templates that will be displayed in the UI when navigating to App Templates
		TemplatesURL string `json:"TemplatesURL" example:"https://raw.githubusercontent.com/portainer/templates/master/templates.json"`
		// The default check in interval for edge agent (in seconds)
		EdgeAgentCheckinInterval int `json:"EdgeAgentCheckinInterval" example:"5"`
		// Whether edge compute features are enabled
		EnableEdgeComputeFeatures bool `json:"EnableEdgeComputeFeatures" example:""`
		// The duration of a user session
		UserSessionTimeout string `json:"UserSessionTimeout" example:"5m"`
		// Whether telemetry is enabled
		EnableTelemetry bool `json:"EnableTelemetry" example:"false"`

		// Deprecated fields
		DisplayDonationHeader       bool
		DisplayExternalContributors bool

		// Deprecated fields v26
		EnableHostManagementFeatures              bool `json:"EnableHostManagementFeatures"`
		AllowVolumeBrowserForRegularUsers         bool `json:"AllowVolumeBrowserForRegularUsers"`
		AllowBindMountsForRegularUsers            bool `json:"AllowBindMountsForRegularUsers"`
		AllowPrivilegedModeForRegularUsers        bool `json:"AllowPrivilegedModeForRegularUsers"`
		AllowHostNamespaceForRegularUsers         bool `json:"AllowHostNamespaceForRegularUsers"`
		AllowStackManagementForRegularUsers       bool `json:"AllowStackManagementForRegularUsers"`
		AllowDeviceMappingForRegularUsers         bool `json:"AllowDeviceMappingForRegularUsers"`
		AllowContainerCapabilitiesForRegularUsers bool `json:"AllowContainerCapabilitiesForRegularUsers"`
	}

	// SnapshotJob represents a scheduled job that can create endpoint snapshots
	SnapshotJob struct{}

	// SoftwareEdition represents an edition of Portainer
	SoftwareEdition int

	// SSLSettings represents a pair of SSL certificate and key
	SSLSettings struct {
		CertPath    string `json:"certPath"`
		KeyPath     string `json:"keyPath"`
		SelfSigned  bool   `json:"selfSigned"`
		HTTPEnabled bool   `json:"httpEnabled"`
	}

	// Stack represents a Docker stack created via docker stack deploy
	Stack struct {
		// Stack Identifier
		ID StackID `json:"Id" example:"1"`
		// Stack name
		Name string `json:"Name" example:"myStack"`
		// Stack type. 1 for a Swarm stack, 2 for a Compose stack
		Type StackType `json:"Type" example:"2"`
		// Endpoint identifier. Reference the endpoint that will be used for deployment
		EndpointID EndpointID `json:"EndpointId" example:"1"`
		// Cluster identifier of the Swarm cluster where the stack is deployed
		SwarmID string `json:"SwarmId" example:"jpofkc0i9uo9wtx1zesuk649w"`
		// Path to the Stack file
		EntryPoint string `json:"EntryPoint" example:"docker-compose.yml"`
		// A list of environment variables used during stack deployment
		Env []Pair `json:"Env" example:""`
		//
		ResourceControl *ResourceControl `json:"ResourceControl" example:""`
		// Stack status (1 - active, 2 - inactive)
		Status StackStatus `json:"Status" example:"1"`
		// Path on disk to the repository hosting the Stack file
		ProjectPath string `example:"/data/compose/myStack_jpofkc0i9uo9wtx1zesuk649w"`
		// The date in unix time when stack was created
		CreationDate int64 `example:"1587399600"`
		// The username which created this stack
		CreatedBy string `example:"admin"`
		// The date in unix time when stack was last updated
		UpdateDate int64 `example:"1587399600"`
		// The username which last updated this stack
		UpdatedBy string `example:"bob"`
		// The git config of this stack
		GitConfig *gittypes.RepoConfig
	}

	// StackID represents a stack identifier (it must be composed of Name + "_" + SwarmID to create a unique identifier)
	StackID int

	// StackStatus represent a status for a stack
	StackStatus int

	// StackType represents the type of the stack (compose v2, stack deploy v3)
	StackType int

	// Status represents the application status
	Status struct {
		// Portainer API version
		Version string `json:"Version" example:"2.0.0"`
	}

	// Tag represents a tag that can be associated to a resource
	Tag struct {
		// Tag identifier
		ID TagID `example:"1"`
		// Tag name
		Name string `json:"Name" example:"org/acme"`
		// A set of endpoint ids that have this tag
		Endpoints map[EndpointID]bool `json:"Endpoints"`
		// A set of endpoint group ids that have this tag
		EndpointGroups map[EndpointGroupID]bool `json:"EndpointGroups"`
	}

	// TagID represents a tag identifier
	TagID int

	// Team represents a list of user accounts
	Team struct {
		// Team Identifier
		ID TeamID `json:"Id" example:"1"`
		// Team name
		Name string `json:"Name" example:"developers"`
	}

	// TeamAccessPolicies represent the association of an access policy and a team
	TeamAccessPolicies map[TeamID]AccessPolicy

	// TeamID represents a team identifier
	TeamID int

	// TeamMembership represents a membership association between a user and a team
	TeamMembership struct {
		// Membership Identifier
		ID TeamMembershipID `json:"Id" example:"1"`
		// User identifier
		UserID UserID `json:"UserID" example:"1"`
		// Team identifier
		TeamID TeamID `json:"TeamID" example:"1"`
		// Team role (1 for team leader and 2 for team member)
		Role MembershipRole `json:"Role" example:"1"`
	}

	// TeamMembershipID represents a team membership identifier
	TeamMembershipID int

	// TeamResourceAccess represents the level of control on a resource for a specific team
	TeamResourceAccess struct {
		TeamID      TeamID              `json:"TeamId"`
		AccessLevel ResourceAccessLevel `json:"AccessLevel"`
	}

	// Template represents an application template that can be used as an App Template
	// or an Edge template
	Template struct {
		// Mandatory container/stack fields
		// Template Identifier
		ID TemplateID `json:"Id" example:"1"`
		// Template type. Valid values are: 1 (container), 2 (Swarm stack) or 3 (Compose stack)
		Type TemplateType `json:"type" example:"1"`
		// Title of the template
		Title string `json:"title" example:"Nginx"`
		// Description of the template
		Description string `json:"description" example:"High performance web server"`
		// Whether the template should be available to administrators only
		AdministratorOnly bool `json:"administrator_only" example:"true"`

		// Mandatory container fields
		// Image associated to a container template. Mandatory for a container template
		Image string `json:"image" example:"nginx:latest"`

		// Mandatory stack fields
		Repository TemplateRepository `json:"repository"`

		// Mandatory Edge stack fields
		// Stack file used for this template
		StackFile string `json:"stackFile"`

		// Optional stack/container fields
		// Default name for the stack/container to be used on deployment
		Name string `json:"name,omitempty" example:"mystackname"`
		// URL of the template's logo
		Logo string `json:"logo,omitempty" example:"https://cloudinovasi.id/assets/img/logos/nginx.png"`
		// A list of environment variables used during the template deployment
		Env []TemplateEnv `json:"env,omitempty"`
		// A note that will be displayed in the UI. Supports HTML content
		Note string `json:"note,omitempty" example:"This is my <b>custom</b> template"`
		// Platform associated to the template.
		// Valid values are: 'linux', 'windows' or leave empty for multi-platform
		Platform string `json:"platform,omitempty" example:"linux"`
		// A list of categories associated to the template
		Categories []string `json:"categories,omitempty" example:"database"`

		// Optional container fields
		// The URL of a registry associated to the image for a container template
		Registry string `json:"registry,omitempty" example:"quay.io"`
		// The command that will be executed in a container template
		Command string `json:"command,omitempty" example:"ls -lah"`
		// Name of a network that will be used on container deployment if it exists inside the environment
		Network string `json:"network,omitempty" example:"mynet"`
		// A list of volumes used during the container template deployment
		Volumes []TemplateVolume `json:"volumes,omitempty"`
		// A list of ports exposed by the container
		Ports []string `json:"ports,omitempty" example:"8080:80/tcp"`
		// Container labels
		Labels []Pair `json:"labels,omitempty" example:""`
		// Whether the container should be started in privileged mode
		Privileged bool `json:"privileged,omitempty" example:"true"`
		// Whether the container should be started in
		// interactive mode (-i -t equivalent on the CLI)
		Interactive bool `json:"interactive,omitempty" example:"true"`
		// Container restart policy
		RestartPolicy string `json:"restart_policy,omitempty" example:"on-failure"`
		// Container hostname
		Hostname string `json:"hostname,omitempty" example:"mycontainer"`
	}

	// TemplateEnv represents a template environment variable configuration
	TemplateEnv struct {
		// name of the environment variable
		Name string `json:"name" example:"MYSQL_ROOT_PASSWORD"`
		// Text for the label that will be generated in the UI
		Label string `json:"label,omitempty" example:"Root password"`
		// Content of the tooltip that will be generated in the UI
		Description string `json:"description,omitempty" example:"MySQL root account password"`
		// Default value that will be set for the variable
		Default string `json:"default,omitempty" example:"default_value"`
		// If set to true, will not generate any input for this variable in the UI
		Preset bool `json:"preset,omitempty" example:"false"`
		// A list of name/value that will be used to generate a dropdown in the UI
		Select []TemplateEnvSelect `json:"select,omitempty"`
	}

	// TemplateEnvSelect represents text/value pair that will be displayed as a choice for the
	// template user
	TemplateEnvSelect struct {
		// Some text that will displayed as a choice
		Text string `json:"text" example:"text value"`
		// A value that will be associated to the choice
		Value string `json:"value" example:"value"`
		// Will set this choice as the default choice
		Default bool `json:"default" example:"false"`
	}

	// TemplateID represents a template identifier
	TemplateID int

	// TemplateRepository represents the git repository configuration for a template
	TemplateRepository struct {
		// URL of a git repository used to deploy a stack template. Mandatory for a Swarm/Compose stack template
		URL string `json:"url" example:"https://github.com/portainer/portainer-compose"`
		// Path to the stack file inside the git repository
		StackFile string `json:"stackfile" example:"./subfolder/docker-compose.yml"`
	}

	// TemplateType represents the type of a template
	TemplateType int

	// TemplateVolume represents a template volume configuration
	TemplateVolume struct {
		// Path inside the container
		Container string `json:"container" example:"/data"`
		// Path on the host
		Bind string `json:"bind,omitempty" example:"/tmp"`
		// Whether the volume used should be readonly
		ReadOnly bool `json:"readonly,omitempty" example:"true"`
	}

	// TLSConfiguration represents a TLS configuration
	TLSConfiguration struct {
		// Use TLS
		TLS bool `json:"TLS" example:"true"`
		// Skip the verification of the server TLS certificate
		TLSSkipVerify bool `json:"TLSSkipVerify" example:"false"`
		// Path to the TLS CA certificate file
		TLSCACertPath string `json:"TLSCACert,omitempty" example:"/data/tls/ca.pem"`
		// Path to the TLS client certificate file
		TLSCertPath string `json:"TLSCert,omitempty" example:"/data/tls/cert.pem"`
		// Path to the TLS client key file
		TLSKeyPath string `json:"TLSKey,omitempty" example:"/data/tls/key.pem"`
	}

	// TLSFileType represents a type of TLS file required to connect to a Docker endpoint.
	// It can be either a TLS CA file, a TLS certificate file or a TLS key file
	TLSFileType int

	// TokenData represents the data embedded in a JWT token
	TokenData struct {
		ID       UserID
		Username string
		Role     UserRole
	}

	// TunnelDetails represents information associated to a tunnel
	TunnelDetails struct {
		Status       string
		LastActivity time.Time
		Port         int
		Jobs         []EdgeJob
		Credentials  string
	}

	// TunnelServerInfo represents information associated to the tunnel server
	TunnelServerInfo struct {
		PrivateKeySeed string `json:"PrivateKeySeed"`
	}

	// User represents a user account
	User struct {
		// User Identifier
		ID       UserID `json:"Id" example:"1"`
		Username string `json:"Username" example:"bob"`
		Password string `json:"Password,omitempty" example:"passwd"`
		// User role (1 for administrator account and 2 for regular account)
		Role UserRole `json:"Role" example:"1"`

		// Deprecated fields
		// Deprecated in DBVersion == 25
		PortainerAuthorizations Authorizations         `json:"PortainerAuthorizations"`
		EndpointAuthorizations  EndpointAuthorizations `json:"EndpointAuthorizations"`
	}

	// UserAccessPolicies represent the association of an access policy and a user
	UserAccessPolicies map[UserID]AccessPolicy

	// UserID represents a user identifier
	UserID int

	// UserResourceAccess represents the level of control on a resource for a specific user
	UserResourceAccess struct {
		UserID      UserID              `json:"UserId"`
		AccessLevel ResourceAccessLevel `json:"AccessLevel"`
	}

	// UserRole represents the role of a user. It can be either an administrator
	// or a regular user
	UserRole int

	// Webhook represents a url webhook that can be used to update a service
	Webhook struct {
		// Webhook Identifier
		ID          WebhookID   `json:"Id" example:"1"`
		Token       string      `json:"Token"`
		ResourceID  string      `json:"ResourceId"`
		EndpointID  EndpointID  `json:"EndpointId"`
		WebhookType WebhookType `json:"Type"`
	}

	// WebhookID represents a webhook identifier.
	WebhookID int

	// WebhookType represents the type of resource a webhook is related to
	WebhookType int

	// CLIService represents a service for managing CLI
	CLIService interface {
		ParseFlags(version string) (*CLIFlags, error)
		ValidateFlags(flags *CLIFlags) error
	}

	// ComposeStackManager represents a service to manage Compose stacks
	ComposeStackManager interface {
		ComposeSyntaxMaxVersion() string
		NormalizeStackName(name string) string
		Up(stack *Stack, endpoint *Endpoint) error
		Down(stack *Stack, endpoint *Endpoint) error
	}

	// CryptoService represents a service for encrypting/hashing data
	CryptoService interface {
		Hash(data string) (string, error)
		CompareHashAndData(hash string, data string) error
	}

	// CustomTemplateService represents a service to manage custom templates
	CustomTemplateService interface {
		GetNextIdentifier() int
		CustomTemplates() ([]CustomTemplate, error)
		CustomTemplate(ID CustomTemplateID) (*CustomTemplate, error)
		CreateCustomTemplate(customTemplate *CustomTemplate) error
		UpdateCustomTemplate(ID CustomTemplateID, customTemplate *CustomTemplate) error
		DeleteCustomTemplate(ID CustomTemplateID) error
	}

	// DataStore defines the interface to manage the data
	DataStore interface {
		Open() error
		Init() error
		Close() error
		IsNew() bool
		MigrateData(force bool) error
		CheckCurrentEdition() error
		BackupTo(w io.Writer) error

		CustomTemplate() CustomTemplateService
		EdgeGroup() EdgeGroupService
		EdgeJob() EdgeJobService
		EdgeStack() EdgeStackService
		Endpoint() EndpointService
		EndpointGroup() EndpointGroupService
		EndpointRelation() EndpointRelationService
		Registry() RegistryService
		ResourceControl() ResourceControlService
		Role() RoleService
		Settings() SettingsService
		SSLSettings() SSLSettingsService
		Stack() StackService
		Tag() TagService
		TeamMembership() TeamMembershipService
		Team() TeamService
		TunnelServer() TunnelServerService
		User() UserService
		Version() VersionService
		Webhook() WebhookService
	}

	// DigitalSignatureService represents a service to manage digital signatures
	DigitalSignatureService interface {
		ParseKeyPair(private, public []byte) error
		GenerateKeyPair() ([]byte, []byte, error)
		EncodedPublicKey() string
		PEMHeaders() (string, string)
		CreateSignature(message string) (string, error)
	}

	// DockerSnapshotter represents a service used to create Docker endpoint snapshots
	DockerSnapshotter interface {
		CreateSnapshot(endpoint *Endpoint) (*DockerSnapshot, error)
	}

	// EdgeGroupService represents a service to manage Edge groups
	EdgeGroupService interface {
		EdgeGroups() ([]EdgeGroup, error)
		EdgeGroup(ID EdgeGroupID) (*EdgeGroup, error)
		CreateEdgeGroup(group *EdgeGroup) error
		UpdateEdgeGroup(ID EdgeGroupID, group *EdgeGroup) error
		DeleteEdgeGroup(ID EdgeGroupID) error
	}

	// EdgeJobService represents a service to manage Edge jobs
	EdgeJobService interface {
		EdgeJobs() ([]EdgeJob, error)
		EdgeJob(ID EdgeJobID) (*EdgeJob, error)
		CreateEdgeJob(edgeJob *EdgeJob) error
		UpdateEdgeJob(ID EdgeJobID, edgeJob *EdgeJob) error
		DeleteEdgeJob(ID EdgeJobID) error
		GetNextIdentifier() int
	}

	// EdgeStackService represents a service to manage Edge stacks
	EdgeStackService interface {
		EdgeStacks() ([]EdgeStack, error)
		EdgeStack(ID EdgeStackID) (*EdgeStack, error)
		CreateEdgeStack(edgeStack *EdgeStack) error
		UpdateEdgeStack(ID EdgeStackID, edgeStack *EdgeStack) error
		DeleteEdgeStack(ID EdgeStackID) error
		GetNextIdentifier() int
	}

	// EndpointService represents a service for managing endpoint data
	EndpointService interface {
		Endpoint(ID EndpointID) (*Endpoint, error)
		Endpoints() ([]Endpoint, error)
		CreateEndpoint(endpoint *Endpoint) error
		UpdateEndpoint(ID EndpointID, endpoint *Endpoint) error
		DeleteEndpoint(ID EndpointID) error
		Synchronize(toCreate, toUpdate, toDelete []*Endpoint) error
		GetNextIdentifier() int
	}

	// EndpointGroupService represents a service for managing endpoint group data
	EndpointGroupService interface {
		EndpointGroup(ID EndpointGroupID) (*EndpointGroup, error)
		EndpointGroups() ([]EndpointGroup, error)
		CreateEndpointGroup(group *EndpointGroup) error
		UpdateEndpointGroup(ID EndpointGroupID, group *EndpointGroup) error
		DeleteEndpointGroup(ID EndpointGroupID) error
	}

	// EndpointRelationService represents a service for managing endpoint relations data
	EndpointRelationService interface {
		EndpointRelation(EndpointID EndpointID) (*EndpointRelation, error)
		CreateEndpointRelation(endpointRelation *EndpointRelation) error
		UpdateEndpointRelation(EndpointID EndpointID, endpointRelation *EndpointRelation) error
		DeleteEndpointRelation(EndpointID EndpointID) error
	}

	// FileService represents a service for managing files
	FileService interface {
		GetFileContent(filePath string) ([]byte, error)
		Rename(oldPath, newPath string) error
		RemoveDirectory(directoryPath string) error
		StoreTLSFileFromBytes(folder string, fileType TLSFileType, data []byte) (string, error)
		GetPathForTLSFile(folder string, fileType TLSFileType) (string, error)
		DeleteTLSFile(folder string, fileType TLSFileType) error
		DeleteTLSFiles(folder string) error
		GetStackProjectPath(stackIdentifier string) string
		StoreStackFileFromBytes(stackIdentifier, fileName string, data []byte) (string, error)
		GetEdgeStackProjectPath(edgeStackIdentifier string) string
		StoreEdgeStackFileFromBytes(edgeStackIdentifier, fileName string, data []byte) (string, error)
		StoreRegistryManagementFileFromBytes(folder, fileName string, data []byte) (string, error)
		KeyPairFilesExist() (bool, error)
		StoreKeyPair(private, public []byte, privatePEMHeader, publicPEMHeader string) error
		LoadKeyPair() ([]byte, []byte, error)
		WriteJSONToFile(path string, content interface{}) error
		FileExists(path string) (bool, error)
		StoreEdgeJobFileFromBytes(identifier string, data []byte) (string, error)
		GetEdgeJobFolder(identifier string) string
		ClearEdgeJobTaskLogs(edgeJobID, taskID string) error
		GetEdgeJobTaskLogFileContent(edgeJobID, taskID string) (string, error)
		StoreEdgeJobTaskLogFileFromBytes(edgeJobID, taskID string, data []byte) error
		GetBinaryFolder() string
		StoreCustomTemplateFileFromBytes(identifier, fileName string, data []byte) (string, error)
		GetCustomTemplateProjectPath(identifier string) string
		GetTemporaryPath() (string, error)
		GetDatastorePath() string
		GetDefaultSSLCertsPath() (string, string)
		StoreSSLCertPair(cert, key []byte) (string, string, error)
		CopySSLCertPair(certPath, keyPath string) (string, string, error)
	}

	// GitService represents a service for managing Git
	GitService interface {
		CloneRepository(destination string, repositoryURL, referenceName, username, password string) error
	}

	// JWTService represents a service for managing JWT tokens
	JWTService interface {
		GenerateToken(data *TokenData) (string, error)
		GenerateTokenForOAuth(data *TokenData, expiryTime *time.Time) (string, error)
		ParseAndVerifyToken(token string) (*TokenData, error)
		SetUserSessionDuration(userSessionDuration time.Duration)
	}

	// KubeClient represents a service used to query a Kubernetes environment
	KubeClient interface {
		SetupUserServiceAccount(userID int, teamIDs []int) error
		GetServiceAccountBearerToken(userID int) (string, error)
		StartExecProcess(namespace, podName, containerName string, command []string, stdin io.Reader, stdout io.Writer) error
		GetNamespaceAccessPolicies() (map[string]K8sNamespaceAccessPolicy, error)
		UpdateNamespaceAccessPolicies(accessPolicies map[string]K8sNamespaceAccessPolicy) error
		DeleteRegistrySecret(registry *Registry, namespace string) error
		CreateRegistrySecret(registry *Registry, namespace string) error
		IsRegistrySecret(namespace, secretName string) (bool, error)
	}

	// KubernetesDeployer represents a service to deploy a manifest inside a Kubernetes endpoint
	KubernetesDeployer interface {
		Deploy(endpoint *Endpoint, data string, namespace string) (string, error)
		ConvertCompose(data string) ([]byte, error)
	}

	// KubernetesSnapshotter represents a service used to create Kubernetes endpoint snapshots
	KubernetesSnapshotter interface {
		CreateSnapshot(endpoint *Endpoint) (*KubernetesSnapshot, error)
	}

	// LDAPService represents a service used to authenticate users against a LDAP/AD
	LDAPService interface {
		AuthenticateUser(username, password string, settings *LDAPSettings) error
		TestConnectivity(settings *LDAPSettings) error
		GetUserGroups(username string, settings *LDAPSettings) ([]string, error)
	}

	// OAuthService represents a service used to authenticate users using OAuth
	OAuthService interface {
		Authenticate(code string, configuration *OAuthSettings) (string, *time.Time, error)
	}

	// RegistryService represents a service for managing registry data
	RegistryService interface {
		Registry(ID RegistryID) (*Registry, error)
		Registries() ([]Registry, error)
		CreateRegistry(registry *Registry) error
		UpdateRegistry(ID RegistryID, registry *Registry) error
		DeleteRegistry(ID RegistryID) error
	}

	// ResourceControlService represents a service for managing resource control data
	ResourceControlService interface {
		ResourceControl(ID ResourceControlID) (*ResourceControl, error)
		ResourceControlByResourceIDAndType(resourceID string, resourceType ResourceControlType) (*ResourceControl, error)
		ResourceControls() ([]ResourceControl, error)
		CreateResourceControl(rc *ResourceControl) error
		UpdateResourceControl(ID ResourceControlID, resourceControl *ResourceControl) error
		DeleteResourceControl(ID ResourceControlID) error
	}

	// ReverseTunnelService represents a service used to manage reverse tunnel connections.
	ReverseTunnelService interface {
		StartTunnelServer(addr, port string, snapshotService SnapshotService) error
		StopTunnelServer() error
		GenerateEdgeKey(url, host string, endpointIdentifier int) string
		SetTunnelStatusToActive(endpointID EndpointID)
		SetTunnelStatusToRequired(endpointID EndpointID) error
		SetTunnelStatusToIdle(endpointID EndpointID)
		GetTunnelDetails(endpointID EndpointID) *TunnelDetails
		AddEdgeJob(endpointID EndpointID, edgeJob *EdgeJob)
		RemoveEdgeJob(edgeJobID EdgeJobID)
	}

	// RoleService represents a service for managing user roles
	RoleService interface {
		Role(ID RoleID) (*Role, error)
		Roles() ([]Role, error)
		CreateRole(role *Role) error
		UpdateRole(ID RoleID, role *Role) error
	}

	// SettingsService represents a service for managing application settings
	SettingsService interface {
		Settings() (*Settings, error)
		UpdateSettings(settings *Settings) error
	}

	// Server defines the interface to serve the API
	Server interface {
		Start() error
	}

	// SSLSettingsService represents a service for managing application settings
	SSLSettingsService interface {
		Settings() (*SSLSettings, error)
		UpdateSettings(settings *SSLSettings) error
	}

	// StackService represents a service for managing stack data
	StackService interface {
		Stack(ID StackID) (*Stack, error)
		StackByName(name string) (*Stack, error)
		Stacks() ([]Stack, error)
		CreateStack(stack *Stack) error
		UpdateStack(ID StackID, stack *Stack) error
		DeleteStack(ID StackID) error
		GetNextIdentifier() int
	}

	// SnapshotService represents a service for managing endpoint snapshots
	SnapshotService interface {
		Start()
		Stop()
		SetSnapshotInterval(snapshotInterval string) error
		SnapshotEndpoint(endpoint *Endpoint) error
	}

	// SwarmStackManager represents a service to manage Swarm stacks
	SwarmStackManager interface {
		Login(registries []Registry, endpoint *Endpoint)
		Logout(endpoint *Endpoint) error
		Deploy(stack *Stack, prune bool, endpoint *Endpoint) error
		Remove(stack *Stack, endpoint *Endpoint) error
	}

	// TagService represents a service for managing tag data
	TagService interface {
		Tags() ([]Tag, error)
		Tag(ID TagID) (*Tag, error)
		CreateTag(tag *Tag) error
		UpdateTag(ID TagID, tag *Tag) error
		DeleteTag(ID TagID) error
	}

	// TeamService represents a service for managing user data
	TeamService interface {
		Team(ID TeamID) (*Team, error)
		TeamByName(name string) (*Team, error)
		Teams() ([]Team, error)
		CreateTeam(team *Team) error
		UpdateTeam(ID TeamID, team *Team) error
		DeleteTeam(ID TeamID) error
	}

	// TeamMembershipService represents a service for managing team membership data
	TeamMembershipService interface {
		TeamMembership(ID TeamMembershipID) (*TeamMembership, error)
		TeamMemberships() ([]TeamMembership, error)
		TeamMembershipsByUserID(userID UserID) ([]TeamMembership, error)
		TeamMembershipsByTeamID(teamID TeamID) ([]TeamMembership, error)
		CreateTeamMembership(membership *TeamMembership) error
		UpdateTeamMembership(ID TeamMembershipID, membership *TeamMembership) error
		DeleteTeamMembership(ID TeamMembershipID) error
		DeleteTeamMembershipByUserID(userID UserID) error
		DeleteTeamMembershipByTeamID(teamID TeamID) error
	}

	// TunnelServerService represents a service for managing data associated to the tunnel server
	TunnelServerService interface {
		Info() (*TunnelServerInfo, error)
		UpdateInfo(info *TunnelServerInfo) error
	}

	// UserService represents a service for managing user data
	UserService interface {
		User(ID UserID) (*User, error)
		UserByUsername(username string) (*User, error)
		Users() ([]User, error)
		UsersByRole(role UserRole) ([]User, error)
		CreateUser(user *User) error
		UpdateUser(ID UserID, user *User) error
		DeleteUser(ID UserID) error
	}

	// VersionService represents a service for managing version data
	VersionService interface {
		DBVersion() (int, error)
		Edition() (SoftwareEdition, error)
		InstanceID() (string, error)
		StoreDBVersion(version int) error
		StoreInstanceID(ID string) error
	}

	// WebhookService represents a service for managing webhook data.
	WebhookService interface {
		Webhooks() ([]Webhook, error)
		Webhook(ID WebhookID) (*Webhook, error)
		CreateWebhook(portainer *Webhook) error
		WebhookByResourceID(resourceID string) (*Webhook, error)
		WebhookByToken(token string) (*Webhook, error)
		DeleteWebhook(serviceID WebhookID) error
	}
)

const (
	// APIVersion is the version number of the Portainer API
	APIVersion = "2.6.0"
	// DBVersion is the version number of the Portainer database
	DBVersion = 32
	// ComposeSyntaxMaxVersion is a maximum supported version of the docker compose syntax
	ComposeSyntaxMaxVersion = "3.9"
	// AssetsServerURL represents the URL of the Portainer asset server
	AssetsServerURL = "https://portainer-io-assets.sfo2.digitaloceanspaces.com"
	// MessageOfTheDayURL represents the URL where Portainer MOTD message can be retrieved
	MessageOfTheDayURL = AssetsServerURL + "/motd.json"
	// VersionCheckURL represents the URL used to retrieve the latest version of Portainer
	VersionCheckURL = "https://api.github.com/repos/portainer/portainer/releases/latest"
	// PortainerAgentHeader represents the name of the header available in any agent response
	PortainerAgentHeader = "Portainer-Agent"
	// PortainerAgentEdgeIDHeader represent the name of the header containing the Edge ID associated to an agent/agent cluster
	PortainerAgentEdgeIDHeader = "X-PortainerAgent-EdgeID"
	// HTTPResponseAgentPlatform represents the name of the header containing the Agent platform
	HTTPResponseAgentPlatform = "Portainer-Agent-Platform"
	// PortainerAgentTargetHeader represent the name of the header containing the target node name
	PortainerAgentTargetHeader = "X-PortainerAgent-Target"
	// PortainerAgentSignatureHeader represent the name of the header containing the digital signature
	PortainerAgentSignatureHeader = "X-PortainerAgent-Signature"
	// PortainerAgentPublicKeyHeader represent the name of the header containing the public key
	PortainerAgentPublicKeyHeader = "X-PortainerAgent-PublicKey"
	// PortainerAgentKubernetesSATokenHeader represent the name of the header containing a Kubernetes SA token
	PortainerAgentKubernetesSATokenHeader = "X-PortainerAgent-SA-Token"
	// PortainerAgentSignatureMessage represents the message used to create a digital signature
	// to be used when communicating with an agent
	PortainerAgentSignatureMessage = "Portainer-App"
	// DefaultEdgeAgentCheckinIntervalInSeconds represents the default interval (in seconds) used by Edge agents to checkin with the Portainer instance
	DefaultEdgeAgentCheckinIntervalInSeconds = 5
	// DefaultTemplatesURL represents the URL to the official templates supported by Portainer
	DefaultTemplatesURL = "https://raw.githubusercontent.com/portainer/templates/master/templates-2.0.json"
	// DefaultUserSessionTimeout represents the default timeout after which the user session is cleared
	DefaultUserSessionTimeout = "8h"
)

const (
	_ AuthenticationMethod = iota
	// AuthenticationInternal represents the internal authentication method (authentication against Portainer API)
	AuthenticationInternal
	// AuthenticationLDAP represents the LDAP authentication method (authentication against a LDAP server)
	AuthenticationLDAP
	//AuthenticationOAuth represents the OAuth authentication method (authentication against a authorization server)
	AuthenticationOAuth
)

const (
	_ AgentPlatform = iota
	// AgentPlatformDocker represent the Docker platform (Standalone/Swarm)
	AgentPlatformDocker
	// AgentPlatformKubernetes represent the Kubernetes platform
	AgentPlatformKubernetes
)

const (
	_ EdgeJobLogsStatus = iota
	// EdgeJobLogsStatusIdle represents an idle log collection job
	EdgeJobLogsStatusIdle
	// EdgeJobLogsStatusPending represents a pending log collection job
	EdgeJobLogsStatusPending
	// EdgeJobLogsStatusCollected represents a completed log collection job
	EdgeJobLogsStatusCollected
)

const (
	_ CustomTemplatePlatform = iota
	// CustomTemplatePlatformLinux represents a custom template for linux
	CustomTemplatePlatformLinux
	// CustomTemplatePlatformWindows represents a custom template for windows
	CustomTemplatePlatformWindows
)

const (
	_ EdgeStackStatusType = iota
	//StatusOk represents a successfully deployed edge stack
	StatusOk
	//StatusError represents an edge endpoint which failed to deploy its edge stack
	StatusError
	//StatusAcknowledged represents an acknowledged edge stack
	StatusAcknowledged
)

const (
	_ EndpointExtensionType = iota
	// StoridgeEndpointExtension represents the Storidge extension
	StoridgeEndpointExtension
)

const (
	_ EndpointStatus = iota
	// EndpointStatusUp is used to represent an available endpoint
	EndpointStatusUp
	// EndpointStatusDown is used to represent an unavailable endpoint
	EndpointStatusDown
)

const (
	_ EndpointType = iota
	// DockerEnvironment represents an endpoint connected to a Docker environment
	DockerEnvironment
	// AgentOnDockerEnvironment represents an endpoint connected to a Portainer agent deployed on a Docker environment
	AgentOnDockerEnvironment
	// AzureEnvironment represents an endpoint connected to an Azure environment
	AzureEnvironment
	// EdgeAgentOnDockerEnvironment represents an endpoint connected to an Edge agent deployed on a Docker environment
	EdgeAgentOnDockerEnvironment
	// KubernetesLocalEnvironment represents an endpoint connected to a local Kubernetes environment
	KubernetesLocalEnvironment
	// AgentOnKubernetesEnvironment represents an endpoint connected to a Portainer agent deployed on a Kubernetes environment
	AgentOnKubernetesEnvironment
	// EdgeAgentOnKubernetesEnvironment represents an endpoint connected to an Edge agent deployed on a Kubernetes environment
	EdgeAgentOnKubernetesEnvironment
)

const (
	_ JobType = iota
	// SnapshotJobType is a system job used to create endpoint snapshots
	SnapshotJobType = 2
)

const (
	_ MembershipRole = iota
	// TeamLeader represents a leader role inside a team
	TeamLeader
	// TeamMember represents a member role inside a team
	TeamMember
)

const (
	_ SoftwareEdition = iota
	// PortainerCE represents the community edition of Portainer
	PortainerCE
	// PortainerBE represents the business edition of Portainer
	PortainerBE
	// PortainerEE represents the business edition of Portainer
	PortainerEE
)

const (
	_ RegistryType = iota
	// QuayRegistry represents a Quay.io registry
	QuayRegistry
	// AzureRegistry represents an ACR registry
	AzureRegistry
	// CustomRegistry represents a custom registry
	CustomRegistry
	// GitlabRegistry represents a gitlab registry
	GitlabRegistry
	// ProGetRegistry represents a proget registry
	ProGetRegistry
	// DockerHubRegistry represents a dockerhub registry
	DockerHubRegistry
)

const (
	_ ResourceAccessLevel = iota
	// ReadWriteAccessLevel represents an access level with read-write permissions on a resource
	ReadWriteAccessLevel
)

const (
	_ ResourceControlType = iota
	// ContainerResourceControl represents a resource control associated to a Docker container
	ContainerResourceControl
	// ServiceResourceControl represents a resource control associated to a Docker service
	ServiceResourceControl
	// VolumeResourceControl represents a resource control associated to a Docker volume
	VolumeResourceControl
	// NetworkResourceControl represents a resource control associated to a Docker network
	NetworkResourceControl
	// SecretResourceControl represents a resource control associated to a Docker secret
	SecretResourceControl
	// StackResourceControl represents a resource control associated to a stack composed of Docker services
	StackResourceControl
	// ConfigResourceControl represents a resource control associated to a Docker config
	ConfigResourceControl
	// CustomTemplateResourceControl represents a resource control associated to a custom template
	CustomTemplateResourceControl
	// ContainerGroupResourceControl represents a resource control associated to an Azure container group
	ContainerGroupResourceControl
)

const (
	_ StackType = iota
	// DockerSwarmStack represents a stack managed via docker stack
	DockerSwarmStack
	// DockerComposeStack represents a stack managed via docker-compose
	DockerComposeStack
	// KubernetesStack represents a stack managed via kubectl
	KubernetesStack
)

// StackStatus represents a status for a stack
const (
	_ StackStatus = iota
	StackStatusActive
	StackStatusInactive
)

const (
	_ TemplateType = iota
	// ContainerTemplate represents a container template
	ContainerTemplate
	// SwarmStackTemplate represents a template used to deploy a Swarm stack
	SwarmStackTemplate
	// ComposeStackTemplate represents a template used to deploy a Compose stack
	ComposeStackTemplate
	// EdgeStackTemplate represents a template used to deploy an Edge stack
	EdgeStackTemplate
)

const (
	// TLSFileCA represents a TLS CA certificate file
	TLSFileCA TLSFileType = iota
	// TLSFileCert represents a TLS certificate file
	TLSFileCert
	// TLSFileKey represents a TLS key file
	TLSFileKey
)

const (
	_ UserRole = iota
	// AdministratorRole represents an administrator user role
	AdministratorRole
	// StandardUserRole represents a regular user role
	StandardUserRole
)

const (
	_ WebhookType = iota
	// ServiceWebhook is a webhook for restarting a docker service
	ServiceWebhook
)

const (
	// EdgeAgentIdle represents an idle state for a tunnel connected to an Edge endpoint.
	EdgeAgentIdle string = "IDLE"
	// EdgeAgentManagementRequired represents a required state for a tunnel connected to an Edge endpoint
	EdgeAgentManagementRequired string = "REQUIRED"
	// EdgeAgentActive represents an active state for a tunnel connected to an Edge endpoint
	EdgeAgentActive string = "ACTIVE"
)

// represents an authorization type
const (
	OperationDockerContainerArchiveInfo         Authorization = "DockerContainerArchiveInfo"
	OperationDockerContainerList                Authorization = "DockerContainerList"
	OperationDockerContainerExport              Authorization = "DockerContainerExport"
	OperationDockerContainerChanges             Authorization = "DockerContainerChanges"
	OperationDockerContainerInspect             Authorization = "DockerContainerInspect"
	OperationDockerContainerTop                 Authorization = "DockerContainerTop"
	OperationDockerContainerLogs                Authorization = "DockerContainerLogs"
	OperationDockerContainerStats               Authorization = "DockerContainerStats"
	OperationDockerContainerAttachWebsocket     Authorization = "DockerContainerAttachWebsocket"
	OperationDockerContainerArchive             Authorization = "DockerContainerArchive"
	OperationDockerContainerCreate              Authorization = "DockerContainerCreate"
	OperationDockerContainerPrune               Authorization = "DockerContainerPrune"
	OperationDockerContainerKill                Authorization = "DockerContainerKill"
	OperationDockerContainerPause               Authorization = "DockerContainerPause"
	OperationDockerContainerUnpause             Authorization = "DockerContainerUnpause"
	OperationDockerContainerRestart             Authorization = "DockerContainerRestart"
	OperationDockerContainerStart               Authorization = "DockerContainerStart"
	OperationDockerContainerStop                Authorization = "DockerContainerStop"
	OperationDockerContainerWait                Authorization = "DockerContainerWait"
	OperationDockerContainerResize              Authorization = "DockerContainerResize"
	OperationDockerContainerAttach              Authorization = "DockerContainerAttach"
	OperationDockerContainerExec                Authorization = "DockerContainerExec"
	OperationDockerContainerRename              Authorization = "DockerContainerRename"
	OperationDockerContainerUpdate              Authorization = "DockerContainerUpdate"
	OperationDockerContainerPutContainerArchive Authorization = "DockerContainerPutContainerArchive"
	OperationDockerContainerDelete              Authorization = "DockerContainerDelete"
	OperationDockerImageList                    Authorization = "DockerImageList"
	OperationDockerImageSearch                  Authorization = "DockerImageSearch"
	OperationDockerImageGetAll                  Authorization = "DockerImageGetAll"
	OperationDockerImageGet                     Authorization = "DockerImageGet"
	OperationDockerImageHistory                 Authorization = "DockerImageHistory"
	OperationDockerImageInspect                 Authorization = "DockerImageInspect"
	OperationDockerImageLoad                    Authorization = "DockerImageLoad"
	OperationDockerImageCreate                  Authorization = "DockerImageCreate"
	OperationDockerImagePrune                   Authorization = "DockerImagePrune"
	OperationDockerImagePush                    Authorization = "DockerImagePush"
	OperationDockerImageTag                     Authorization = "DockerImageTag"
	OperationDockerImageDelete                  Authorization = "DockerImageDelete"
	OperationDockerImageCommit                  Authorization = "DockerImageCommit"
	OperationDockerImageBuild                   Authorization = "DockerImageBuild"
	OperationDockerNetworkList                  Authorization = "DockerNetworkList"
	OperationDockerNetworkInspect               Authorization = "DockerNetworkInspect"
	OperationDockerNetworkCreate                Authorization = "DockerNetworkCreate"
	OperationDockerNetworkConnect               Authorization = "DockerNetworkConnect"
	OperationDockerNetworkDisconnect            Authorization = "DockerNetworkDisconnect"
	OperationDockerNetworkPrune                 Authorization = "DockerNetworkPrune"
	OperationDockerNetworkDelete                Authorization = "DockerNetworkDelete"
	OperationDockerVolumeList                   Authorization = "DockerVolumeList"
	OperationDockerVolumeInspect                Authorization = "DockerVolumeInspect"
	OperationDockerVolumeCreate                 Authorization = "DockerVolumeCreate"
	OperationDockerVolumePrune                  Authorization = "DockerVolumePrune"
	OperationDockerVolumeDelete                 Authorization = "DockerVolumeDelete"
	OperationDockerExecInspect                  Authorization = "DockerExecInspect"
	OperationDockerExecStart                    Authorization = "DockerExecStart"
	OperationDockerExecResize                   Authorization = "DockerExecResize"
	OperationDockerSwarmInspect                 Authorization = "DockerSwarmInspect"
	OperationDockerSwarmUnlockKey               Authorization = "DockerSwarmUnlockKey"
	OperationDockerSwarmInit                    Authorization = "DockerSwarmInit"
	OperationDockerSwarmJoin                    Authorization = "DockerSwarmJoin"
	OperationDockerSwarmLeave                   Authorization = "DockerSwarmLeave"
	OperationDockerSwarmUpdate                  Authorization = "DockerSwarmUpdate"
	OperationDockerSwarmUnlock                  Authorization = "DockerSwarmUnlock"
	OperationDockerNodeList                     Authorization = "DockerNodeList"
	OperationDockerNodeInspect                  Authorization = "DockerNodeInspect"
	OperationDockerNodeUpdate                   Authorization = "DockerNodeUpdate"
	OperationDockerNodeDelete                   Authorization = "DockerNodeDelete"
	OperationDockerServiceList                  Authorization = "DockerServiceList"
	OperationDockerServiceInspect               Authorization = "DockerServiceInspect"
	OperationDockerServiceLogs                  Authorization = "DockerServiceLogs"
	OperationDockerServiceCreate                Authorization = "DockerServiceCreate"
	OperationDockerServiceUpdate                Authorization = "DockerServiceUpdate"
	OperationDockerServiceDelete                Authorization = "DockerServiceDelete"
	OperationDockerSecretList                   Authorization = "DockerSecretList"
	OperationDockerSecretInspect                Authorization = "DockerSecretInspect"
	OperationDockerSecretCreate                 Authorization = "DockerSecretCreate"
	OperationDockerSecretUpdate                 Authorization = "DockerSecretUpdate"
	OperationDockerSecretDelete                 Authorization = "DockerSecretDelete"
	OperationDockerConfigList                   Authorization = "DockerConfigList"
	OperationDockerConfigInspect                Authorization = "DockerConfigInspect"
	OperationDockerConfigCreate                 Authorization = "DockerConfigCreate"
	OperationDockerConfigUpdate                 Authorization = "DockerConfigUpdate"
	OperationDockerConfigDelete                 Authorization = "DockerConfigDelete"
	OperationDockerTaskList                     Authorization = "DockerTaskList"
	OperationDockerTaskInspect                  Authorization = "DockerTaskInspect"
	OperationDockerTaskLogs                     Authorization = "DockerTaskLogs"
	OperationDockerPluginList                   Authorization = "DockerPluginList"
	OperationDockerPluginPrivileges             Authorization = "DockerPluginPrivileges"
	OperationDockerPluginInspect                Authorization = "DockerPluginInspect"
	OperationDockerPluginPull                   Authorization = "DockerPluginPull"
	OperationDockerPluginCreate                 Authorization = "DockerPluginCreate"
	OperationDockerPluginEnable                 Authorization = "DockerPluginEnable"
	OperationDockerPluginDisable                Authorization = "DockerPluginDisable"
	OperationDockerPluginPush                   Authorization = "DockerPluginPush"
	OperationDockerPluginUpgrade                Authorization = "DockerPluginUpgrade"
	OperationDockerPluginSet                    Authorization = "DockerPluginSet"
	OperationDockerPluginDelete                 Authorization = "DockerPluginDelete"
	OperationDockerSessionStart                 Authorization = "DockerSessionStart"
	OperationDockerDistributionInspect          Authorization = "DockerDistributionInspect"
	OperationDockerBuildPrune                   Authorization = "DockerBuildPrune"
	OperationDockerBuildCancel                  Authorization = "DockerBuildCancel"
	OperationDockerPing                         Authorization = "DockerPing"
	OperationDockerInfo                         Authorization = "DockerInfo"
	OperationDockerEvents                       Authorization = "DockerEvents"
	OperationDockerSystem                       Authorization = "DockerSystem"
	OperationDockerVersion                      Authorization = "DockerVersion"

	OperationDockerAgentPing         Authorization = "DockerAgentPing"
	OperationDockerAgentList         Authorization = "DockerAgentList"
	OperationDockerAgentHostInfo     Authorization = "DockerAgentHostInfo"
	OperationDockerAgentBrowseDelete Authorization = "DockerAgentBrowseDelete"
	OperationDockerAgentBrowseGet    Authorization = "DockerAgentBrowseGet"
	OperationDockerAgentBrowseList   Authorization = "DockerAgentBrowseList"
	OperationDockerAgentBrowsePut    Authorization = "DockerAgentBrowsePut"
	OperationDockerAgentBrowseRename Authorization = "DockerAgentBrowseRename"

	OperationPortainerDockerHubInspect        Authorization = "PortainerDockerHubInspect"
	OperationPortainerDockerHubUpdate         Authorization = "PortainerDockerHubUpdate"
	OperationPortainerEndpointGroupCreate     Authorization = "PortainerEndpointGroupCreate"
	OperationPortainerEndpointGroupList       Authorization = "PortainerEndpointGroupList"
	OperationPortainerEndpointGroupDelete     Authorization = "PortainerEndpointGroupDelete"
	OperationPortainerEndpointGroupInspect    Authorization = "PortainerEndpointGroupInspect"
	OperationPortainerEndpointGroupUpdate     Authorization = "PortainerEndpointGroupEdit"
	OperationPortainerEndpointGroupAccess     Authorization = "PortainerEndpointGroupAccess "
	OperationPortainerEndpointList            Authorization = "PortainerEndpointList"
	OperationPortainerEndpointInspect         Authorization = "PortainerEndpointInspect"
	OperationPortainerEndpointCreate          Authorization = "PortainerEndpointCreate"
	OperationPortainerEndpointExtensionAdd    Authorization = "PortainerEndpointExtensionAdd"
	OperationPortainerEndpointJob             Authorization = "PortainerEndpointJob"
	OperationPortainerEndpointSnapshots       Authorization = "PortainerEndpointSnapshots"
	OperationPortainerEndpointSnapshot        Authorization = "PortainerEndpointSnapshot"
	OperationPortainerEndpointUpdate          Authorization = "PortainerEndpointUpdate"
	OperationPortainerEndpointUpdateAccess    Authorization = "PortainerEndpointUpdateAccess"
	OperationPortainerEndpointDelete          Authorization = "PortainerEndpointDelete"
	OperationPortainerEndpointExtensionRemove Authorization = "PortainerEndpointExtensionRemove"
	OperationPortainerExtensionList           Authorization = "PortainerExtensionList"
	OperationPortainerExtensionInspect        Authorization = "PortainerExtensionInspect"
	OperationPortainerExtensionCreate         Authorization = "PortainerExtensionCreate"
	OperationPortainerExtensionUpdate         Authorization = "PortainerExtensionUpdate"
	OperationPortainerExtensionDelete         Authorization = "PortainerExtensionDelete"
	OperationPortainerMOTD                    Authorization = "PortainerMOTD"
	OperationPortainerRegistryList            Authorization = "PortainerRegistryList"
	OperationPortainerRegistryInspect         Authorization = "PortainerRegistryInspect"
	OperationPortainerRegistryCreate          Authorization = "PortainerRegistryCreate"
	OperationPortainerRegistryConfigure       Authorization = "PortainerRegistryConfigure"
	OperationPortainerRegistryUpdate          Authorization = "PortainerRegistryUpdate"
	OperationPortainerRegistryUpdateAccess    Authorization = "PortainerRegistryUpdateAccess"
	OperationPortainerRegistryDelete          Authorization = "PortainerRegistryDelete"
	OperationPortainerResourceControlCreate   Authorization = "PortainerResourceControlCreate"
	OperationPortainerResourceControlUpdate   Authorization = "PortainerResourceControlUpdate"
	OperationPortainerResourceControlDelete   Authorization = "PortainerResourceControlDelete"
	OperationPortainerRoleList                Authorization = "PortainerRoleList"
	OperationPortainerRoleInspect             Authorization = "PortainerRoleInspect"
	OperationPortainerRoleCreate              Authorization = "PortainerRoleCreate"
	OperationPortainerRoleUpdate              Authorization = "PortainerRoleUpdate"
	OperationPortainerRoleDelete              Authorization = "PortainerRoleDelete"
	OperationPortainerScheduleList            Authorization = "PortainerScheduleList"
	OperationPortainerScheduleInspect         Authorization = "PortainerScheduleInspect"
	OperationPortainerScheduleFile            Authorization = "PortainerScheduleFile"
	OperationPortainerScheduleTasks           Authorization = "PortainerScheduleTasks"
	OperationPortainerScheduleCreate          Authorization = "PortainerScheduleCreate"
	OperationPortainerScheduleUpdate          Authorization = "PortainerScheduleUpdate"
	OperationPortainerScheduleDelete          Authorization = "PortainerScheduleDelete"
	OperationPortainerSettingsInspect         Authorization = "PortainerSettingsInspect"
	OperationPortainerSettingsUpdate          Authorization = "PortainerSettingsUpdate"
	OperationPortainerSettingsLDAPCheck       Authorization = "PortainerSettingsLDAPCheck"
	OperationPortainerStackList               Authorization = "PortainerStackList"
	OperationPortainerStackInspect            Authorization = "PortainerStackInspect"
	OperationPortainerStackFile               Authorization = "PortainerStackFile"
	OperationPortainerStackCreate             Authorization = "PortainerStackCreate"
	OperationPortainerStackMigrate            Authorization = "PortainerStackMigrate"
	OperationPortainerStackUpdate             Authorization = "PortainerStackUpdate"
	OperationPortainerStackDelete             Authorization = "PortainerStackDelete"
	OperationPortainerTagList                 Authorization = "PortainerTagList"
	OperationPortainerTagCreate               Authorization = "PortainerTagCreate"
	OperationPortainerTagDelete               Authorization = "PortainerTagDelete"
	OperationPortainerTeamMembershipList      Authorization = "PortainerTeamMembershipList"
	OperationPortainerTeamMembershipCreate    Authorization = "PortainerTeamMembershipCreate"
	OperationPortainerTeamMembershipUpdate    Authorization = "PortainerTeamMembershipUpdate"
	OperationPortainerTeamMembershipDelete    Authorization = "PortainerTeamMembershipDelete"
	OperationPortainerTeamList                Authorization = "PortainerTeamList"
	OperationPortainerTeamInspect             Authorization = "PortainerTeamInspect"
	OperationPortainerTeamMemberships         Authorization = "PortainerTeamMemberships"
	OperationPortainerTeamCreate              Authorization = "PortainerTeamCreate"
	OperationPortainerTeamUpdate              Authorization = "PortainerTeamUpdate"
	OperationPortainerTeamDelete              Authorization = "PortainerTeamDelete"
	OperationPortainerTemplateList            Authorization = "PortainerTemplateList"
	OperationPortainerTemplateInspect         Authorization = "PortainerTemplateInspect"
	OperationPortainerTemplateCreate          Authorization = "PortainerTemplateCreate"
	OperationPortainerTemplateUpdate          Authorization = "PortainerTemplateUpdate"
	OperationPortainerTemplateDelete          Authorization = "PortainerTemplateDelete"
	OperationPortainerUploadTLS               Authorization = "PortainerUploadTLS"
	OperationPortainerUserList                Authorization = "PortainerUserList"
	OperationPortainerUserInspect             Authorization = "PortainerUserInspect"
	OperationPortainerUserMemberships         Authorization = "PortainerUserMemberships"
	OperationPortainerUserCreate              Authorization = "PortainerUserCreate"
	OperationPortainerUserUpdate              Authorization = "PortainerUserUpdate"
	OperationPortainerUserUpdatePassword      Authorization = "PortainerUserUpdatePassword"
	OperationPortainerUserDelete              Authorization = "PortainerUserDelete"
	OperationPortainerWebsocketExec           Authorization = "PortainerWebsocketExec"
	OperationPortainerWebhookList             Authorization = "PortainerWebhookList"
	OperationPortainerWebhookCreate           Authorization = "PortainerWebhookCreate"
	OperationPortainerWebhookDelete           Authorization = "PortainerWebhookDelete"

	OperationIntegrationStoridgeAdmin Authorization = "IntegrationStoridgeAdmin"

	OperationDockerUndefined      Authorization = "DockerUndefined"
	OperationDockerAgentUndefined Authorization = "DockerAgentUndefined"
	OperationPortainerUndefined   Authorization = "PortainerUndefined"

	EndpointResourcesAccess Authorization = "EndpointResourcesAccess"
)

const (
	AzurePathContainerGroups = "/subscriptions/*/providers/Microsoft.ContainerInstance/containerGroups"
	AzurePathContainerGroup  = "/subscriptions/*/resourceGroups/*/providers/Microsoft.ContainerInstance/containerGroups/*"
)

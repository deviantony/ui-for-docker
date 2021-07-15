package testhelpers

import (
	"io"

	portainer "github.com/portainer/portainer/api"
)

type datastore struct {
	customTemplate   portainer.CustomTemplateService
	edgeGroup        portainer.EdgeGroupService
	edgeJob          portainer.EdgeJobService
	edgeStack        portainer.EdgeStackService
	endpoint         portainer.EndpointService
	endpointGroup    portainer.EndpointGroupService
	endpointRelation portainer.EndpointRelationService
	registry         portainer.RegistryService
	resourceControl  portainer.ResourceControlService
	role             portainer.RoleService
	sslSettings      portainer.SSLSettingsService
	settings         portainer.SettingsService
	stack            portainer.StackService
	tag              portainer.TagService
	teamMembership   portainer.TeamMembershipService
	team             portainer.TeamService
	tunnelServer     portainer.TunnelServerService
	user             portainer.UserService
	version          portainer.VersionService
	webhook          portainer.WebhookService
}

func (d *datastore) BackupTo(io.Writer) error                            { return nil }
func (d *datastore) Open() error                                         { return nil }
func (d *datastore) Init() error                                         { return nil }
func (d *datastore) Close() error                                        { return nil }
func (d *datastore) CheckCurrentEdition() error                          { return nil }
func (d *datastore) IsNew() bool                                         { return false }
func (d *datastore) MigrateData(force bool) error                        { return nil }
func (d *datastore) RollbackToCE() error                                 { return nil }
func (d *datastore) CustomTemplate() portainer.CustomTemplateService     { return d.customTemplate }
func (d *datastore) EdgeGroup() portainer.EdgeGroupService               { return d.edgeGroup }
func (d *datastore) EdgeJob() portainer.EdgeJobService                   { return d.edgeJob }
func (d *datastore) EdgeStack() portainer.EdgeStackService               { return d.edgeStack }
func (d *datastore) Endpoint() portainer.EndpointService                 { return d.endpoint }
func (d *datastore) EndpointGroup() portainer.EndpointGroupService       { return d.endpointGroup }
func (d *datastore) EndpointRelation() portainer.EndpointRelationService { return d.endpointRelation }
func (d *datastore) Registry() portainer.RegistryService                 { return d.registry }
func (d *datastore) ResourceControl() portainer.ResourceControlService   { return d.resourceControl }
func (d *datastore) Role() portainer.RoleService                         { return d.role }
func (d *datastore) Settings() portainer.SettingsService                 { return d.settings }
func (d *datastore) SSLSettings() portainer.SSLSettingsService           { return d.sslSettings }
func (d *datastore) Stack() portainer.StackService                       { return d.stack }
func (d *datastore) Tag() portainer.TagService                           { return d.tag }
func (d *datastore) TeamMembership() portainer.TeamMembershipService     { return d.teamMembership }
func (d *datastore) Team() portainer.TeamService                         { return d.team }
func (d *datastore) TunnelServer() portainer.TunnelServerService         { return d.tunnelServer }
func (d *datastore) User() portainer.UserService                         { return d.user }
func (d *datastore) Version() portainer.VersionService                   { return d.version }
func (d *datastore) Webhook() portainer.WebhookService                   { return d.webhook }

type datastoreOption = func(d *datastore)

// NewDatastore creates new instance of datastore.
// Will apply options before returning, opts will be applied from left to right.
func NewDatastore(options ...datastoreOption) *datastore {
	d := datastore{}
	for _, o := range options {
		o(&d)
	}
	return &d
}

type stubUserService struct {
	users []portainer.User
}

func (s *stubUserService) User(ID portainer.UserID) (*portainer.User, error)       { return nil, nil }
func (s *stubUserService) UserByUsername(username string) (*portainer.User, error) { return nil, nil }
func (s *stubUserService) Users() ([]portainer.User, error)                        { return s.users, nil }
func (s *stubUserService) UsersByRole(role portainer.UserRole) ([]portainer.User, error) {
	return s.users, nil
}
func (s *stubUserService) CreateUser(user *portainer.User) error                      { return nil }
func (s *stubUserService) UpdateUser(ID portainer.UserID, user *portainer.User) error { return nil }
func (s *stubUserService) DeleteUser(ID portainer.UserID) error                       { return nil }

// WithUsers datastore option that will instruct datastore to return provided users
func WithUsers(us []portainer.User) datastoreOption {
	return func(d *datastore) {
		d.user = &stubUserService{users: us}
	}
}

type stubEdgeJobService struct {
	jobs []portainer.EdgeJob
}

func (s *stubEdgeJobService) EdgeJobs() ([]portainer.EdgeJob, error) { return s.jobs, nil }
func (s *stubEdgeJobService) EdgeJob(ID portainer.EdgeJobID) (*portainer.EdgeJob, error) {
	return nil, nil
}
func (s *stubEdgeJobService) CreateEdgeJob(edgeJob *portainer.EdgeJob) error { return nil }
func (s *stubEdgeJobService) UpdateEdgeJob(ID portainer.EdgeJobID, edgeJob *portainer.EdgeJob) error {
	return nil
}
func (s *stubEdgeJobService) DeleteEdgeJob(ID portainer.EdgeJobID) error { return nil }
func (s *stubEdgeJobService) GetNextIdentifier() int                     { return 0 }

// WithEdgeJobs option will instruct datastore to return provided jobs
func WithEdgeJobs(js []portainer.EdgeJob) datastoreOption {
	return func(d *datastore) {
		d.edgeJob = &stubEdgeJobService{jobs: js}
	}
}

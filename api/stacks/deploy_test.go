package stacks

import (
	"errors"
	"io/ioutil"
	"strings"
	"testing"

	portainer "github.com/portainer/portainer/api"
	bolt "github.com/portainer/portainer/api/bolt/bolttest"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/stretchr/testify/assert"
)

type gitService struct {
	cloneErr error
	id       string
}

func (g *gitService) CloneRepository(destination, repositoryURL, referenceName, username, password string) error {
	return g.cloneErr
}

func (g *gitService) LatestCommitID(repositoryURL, referenceName, username, password string) (string, error) {
	return g.id, nil
}

type noopDeployer struct{}

func (s *noopDeployer) DeploySwarmStack(stack *portainer.Stack, endpoint *portainer.Endpoint, registries []portainer.Registry, prune bool) error {
	return nil
}

func (s *noopDeployer) DeployComposeStack(stack *portainer.Stack, endpoint *portainer.Endpoint, registries []portainer.Registry) error {
	return nil
}

func Test_redeployWhenChanged_FailsWhenCannotFindStack(t *testing.T) {
	store, teardown := bolt.MustNewTestStore(true)
	defer teardown()

	err := RedeployWhenChanged(1, nil, store, nil)
	assert.Error(t, err)
	assert.Truef(t, strings.HasPrefix(err.Error(), "failed to get the stack"), "it isn't an error we expected: %v", err.Error())
}

func Test_redeployWhenChanged_DoesNothingWhenNotAGitBasedStack(t *testing.T) {
	store, teardown := bolt.MustNewTestStore(true)
	defer teardown()

	err := store.Stack().CreateStack(&portainer.Stack{ID: 1})
	assert.NoError(t, err, "failed to create a test stack")

	err = RedeployWhenChanged(1, nil, store, &gitService{nil, ""})
	assert.NoError(t, err)
}

func Test_redeployWhenChanged_DoesNothingWhenNoGitChanges(t *testing.T) {
	store, teardown := bolt.MustNewTestStore(true)
	defer teardown()

	tmpDir, _ := ioutil.TempDir("", "stack")

	err := store.Stack().CreateStack(&portainer.Stack{
		ID:          1,
		ProjectPath: tmpDir,
		GitConfig: &gittypes.RepoConfig{
			URL:           "url",
			ReferenceName: "ref",
			ConfigHash:    "oldHash",
		}})
	assert.NoError(t, err, "failed to create a test stack")

	err = RedeployWhenChanged(1, nil, store, &gitService{nil, "oldHash"})
	assert.NoError(t, err)
}

func Test_redeployWhenChanged_FailsWhenCannotClone(t *testing.T) {
	cloneErr := errors.New("failed to clone")
	store, teardown := bolt.MustNewTestStore(true)
	defer teardown()

	err := store.Stack().CreateStack(&portainer.Stack{
		ID: 1,
		GitConfig: &gittypes.RepoConfig{
			URL:           "url",
			ReferenceName: "ref",
			ConfigHash:    "oldHash",
		}})
	assert.NoError(t, err, "failed to create a test stack")

	err = RedeployWhenChanged(1, nil, store, &gitService{cloneErr, "newHash"})
	assert.Error(t, err)
	assert.ErrorIs(t, err, cloneErr, "should failed to clone but didn't, check test setup")
}

func Test_redeployWhenChanged(t *testing.T) {
	store, teardown := bolt.MustNewTestStore(true)
	defer teardown()

	tmpDir, _ := ioutil.TempDir("", "stack")

	err := store.Endpoint().CreateEndpoint(&portainer.Endpoint{ID: 1})
	assert.NoError(t, err, "error creating endpoint")

	username := "user"
	err = store.User().CreateUser(&portainer.User{Username: username, Role: portainer.AdministratorRole})
	assert.NoError(t, err, "error creating a user")

	stack := portainer.Stack{
		ID:          1,
		EndpointID:  1,
		ProjectPath: tmpDir,
		UpdatedBy:   username,
		GitConfig: &gittypes.RepoConfig{
			URL:           "url",
			ReferenceName: "ref",
			ConfigHash:    "oldHash",
		}}
	err = store.Stack().CreateStack(&stack)
	assert.NoError(t, err, "failed to create a test stack")

	t.Run("can deploy docker compose stack", func(t *testing.T) {
		stack.Type = portainer.DockerComposeStack
		store.Stack().UpdateStack(stack.ID, &stack)

		err = RedeployWhenChanged(1, &noopDeployer{}, store, &gitService{nil, "newHash"})
		assert.NoError(t, err)
	})

	t.Run("can deploy docker swarm stack", func(t *testing.T) {
		stack.Type = portainer.DockerSwarmStack
		store.Stack().UpdateStack(stack.ID, &stack)

		err = RedeployWhenChanged(1, &noopDeployer{}, store, &gitService{nil, "newHash"})
		assert.NoError(t, err)
	})

	t.Run("can NOT deploy kube stack", func(t *testing.T) {
		stack.Type = portainer.KubernetesStack
		store.Stack().UpdateStack(stack.ID, &stack)

		err = RedeployWhenChanged(1, &noopDeployer{}, store, &gitService{nil, "newHash"})
		assert.EqualError(t, err, "cannot update stack, type 3 is unsupported")
	})
}

func Test_getUserRegistries(t *testing.T) {
	store, teardown := bolt.MustNewTestStore(true)
	defer teardown()

	endpointID := 123

	admin := portainer.User{ID: 1, Username: "admin", Role: portainer.AdministratorRole}
	err := store.User().CreateUser(&admin)
	assert.NoError(t, err, "error creating an admin")

	user := portainer.User{ID: 2, Username: "user", Role: portainer.StandardUserRole}
	err = store.User().CreateUser(&user)
	assert.NoError(t, err, "error creating a user")

	team := portainer.Team{ID: 1, Name: "team"}

	store.TeamMembership().CreateTeamMembership(&portainer.TeamMembership{
		ID:     1,
		UserID: user.ID,
		TeamID: team.ID,
		Role:   portainer.TeamMember,
	})

	registryReachableByUser := portainer.Registry{
		ID: 1,
		RegistryAccesses: portainer.RegistryAccesses{
			portainer.EndpointID(endpointID): {
				UserAccessPolicies: map[portainer.UserID]portainer.AccessPolicy{
					user.ID: {RoleID: portainer.RoleID(portainer.StandardUserRole)},
				},
			},
		},
	}
	err = store.Registry().CreateRegistry(&registryReachableByUser)
	assert.NoError(t, err, "couldn't create a registry")

	registryReachableByTeam := portainer.Registry{
		ID: 2,
		RegistryAccesses: portainer.RegistryAccesses{
			portainer.EndpointID(endpointID): {
				TeamAccessPolicies: map[portainer.TeamID]portainer.AccessPolicy{
					team.ID: {RoleID: portainer.RoleID(portainer.StandardUserRole)},
				},
			},
		},
	}
	err = store.Registry().CreateRegistry(&registryReachableByTeam)
	assert.NoError(t, err, "couldn't create a registry")

	registryRestricted := portainer.Registry{
		ID: 3,
		RegistryAccesses: portainer.RegistryAccesses{
			portainer.EndpointID(endpointID): {
				UserAccessPolicies: map[portainer.UserID]portainer.AccessPolicy{
					user.ID + 100: {RoleID: portainer.RoleID(portainer.StandardUserRole)},
				},
			},
		},
	}
	err = store.Registry().CreateRegistry(&registryRestricted)
	assert.NoError(t, err, "couldn't create a registry")

	t.Run("admin should has access to all registries", func(t *testing.T) {
		registries, err := getUserRegistries(store, admin.Username, portainer.EndpointID(endpointID))
		assert.NoError(t, err)
		assert.ElementsMatch(t, []portainer.Registry{registryReachableByUser, registryReachableByTeam, registryRestricted}, registries)
	})

	t.Run("regular user has access to registries allowed to him and/or his team", func(t *testing.T) {
		registries, err := getUserRegistries(store, user.Username, portainer.EndpointID(endpointID))
		assert.NoError(t, err)
		assert.ElementsMatch(t, []portainer.Registry{registryReachableByUser, registryReachableByTeam}, registries)
	})
}

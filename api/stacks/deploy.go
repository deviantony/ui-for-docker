package stacks

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/security"
)

func RedeployWhenChanged(stackID portainer.StackID, deployer StackDeployer, datastore portainer.DataStore, gitService portainer.GitService) error {
	stack, err := datastore.Stack().Stack(stackID)
	if err != nil {
		return errors.WithMessagef(err, "failed to get the stack %v", stackID)
	}

	if stack.GitConfig == nil {
		return nil // do nothing if it isn't a git-based stack
	}

	username, password := "", ""
	if stack.GitConfig.Authentication != nil {
		username, password = stack.GitConfig.Authentication.Username, stack.GitConfig.Authentication.Password
	}

	newHash, err := gitService.LatestCommitID(stack.GitConfig.URL, stack.GitConfig.ReferenceName, username, password)
	if err != nil {
		return errors.WithMessagef(err, "failed to fetch latest commit id of the stack %v", stack.ID)
	}

	if strings.EqualFold(newHash, string(stack.GitConfig.ConfigHash)) {
		return nil
	}

	cloneParams := &cloneRepositoryParameters{
		url:   stack.GitConfig.URL,
		ref:   stack.GitConfig.ReferenceName,
		toDir: stack.ProjectPath,
	}
	if stack.GitConfig.Authentication != nil {
		cloneParams.auth = &gitAuth{
			username: username,
			password: password,
		}
	}

	if err := cloneGitRepository(gitService, cloneParams); err != nil {
		return errors.WithMessagef(err, "failed to do a fresh clone of the stack %v", stack.ID)
	}

	endpoint, err := datastore.Endpoint().Endpoint(stack.EndpointID)
	if err != nil {
		return errors.WithMessagef(err, "failed to find the endpoint %v associated to the stack %v", stack.EndpointID, stack.ID)
	}

	author := stack.UpdatedBy
	if author == "" {
		author = stack.CreatedBy
	}

	registries, err := getUserRegistries(datastore, author, endpoint.ID)
	if err != nil {
		return err
	}

	switch stack.Type {
	case portainer.DockerComposeStack:
		err := deployer.DeployComposeStack(stack, endpoint, registries)
		if err != nil {
			return errors.WithMessagef(err, "failed to deploy a docker compose stack %v", stackID)
		}
	case portainer.DockerSwarmStack:
		err := deployer.DeploySwarmStack(stack, endpoint, registries, true)
		if err != nil {
			return errors.WithMessagef(err, "failed to deploy a docker compose stack %v", stackID)
		}
	default:
		return errors.Errorf("cannot update stack, type %v is unsupported", stack.Type)
	}

	stack.UpdateDate = time.Now().Unix()
	stack.GitConfig.ConfigHash = newHash
	if err := datastore.Stack().UpdateStack(stack.ID, stack); err != nil {
		return errors.WithMessagef(err, "failed to update the stack %v", stack.ID)
	}

	return nil
}

func getUserRegistries(datastore portainer.DataStore, authorUsername string, endpointID portainer.EndpointID) ([]portainer.Registry, error) {
	registries, err := datastore.Registry().Registries()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to retrieve registries from the database")
	}

	user, err := datastore.User().UserByUsername(authorUsername)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to fetch a stack's author [%s]", authorUsername)
	}

	if user.Role == portainer.AdministratorRole {
		return registries, nil
	}

	userMemberships, err := datastore.TeamMembership().TeamMembershipsByUserID(user.ID)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to fetch memberships of the stack author [%s]", authorUsername)
	}

	filteredRegistries := make([]portainer.Registry, 0, len(registries))
	for _, registry := range registries {
		if security.AuthorizedRegistryAccess(&registry, user, userMemberships, endpointID) {
			filteredRegistries = append(filteredRegistries, registry)
		}
	}

	return filteredRegistries, nil
}

type cloneRepositoryParameters struct {
	url   string
	ref   string
	toDir string
	auth  *gitAuth
}

type gitAuth struct {
	username string
	password string
}

func cloneGitRepository(gitService portainer.GitService, cloneParams *cloneRepositoryParameters) error {
	if cloneParams.auth != nil {
		return gitService.CloneRepository(cloneParams.toDir, cloneParams.url, cloneParams.ref, cloneParams.auth.username, cloneParams.auth.password)
	}
	return gitService.CloneRepository(cloneParams.toDir, cloneParams.url, cloneParams.ref, "", "")
}

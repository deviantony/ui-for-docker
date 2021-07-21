package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/docker/pkg/ioutils"
	"github.com/stretchr/testify/assert"
)

func TestService_ClonePrivateRepository_GitHub(t *testing.T) {
	ensureIntegrationTest(t)

	accessToken := getRequiredValue(t, "GITHUB_PAT")
	username := getRequiredValue(t, "GITHUB_USERNAME")
	service := NewService()

	dst, err := ioutils.TempDir("", "clone")
	assert.NoError(t, err)
	defer os.RemoveAll(dst)

	repositoryUrl := "https://github.com/portainer/private-test-repository.git"
	err = service.CloneRepository(dst, repositoryUrl, "refs/heads/main", username, accessToken)
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(dst, "README.md"))
}

func TestService_LatestCommitID_GitHub(t *testing.T) {
	ensureIntegrationTest(t)

	accessToken := getRequiredValue(t, "GITHUB_PAT")
	username := getRequiredValue(t, "GITHUB_USERNAME")
	service := NewService()

	repositoryUrl := "https://github.com/portainer/private-test-repository.git"
	id, err := service.LatestCommitID(repositoryUrl, "refs/heads/main", username, accessToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, id, "cannot guarantee commit id, but it should be not empty")
}

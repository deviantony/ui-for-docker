package git

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/docker/pkg/ioutils"
	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/assert"
)

func TestService_ClonePublicRepository_Azure(t *testing.T) {
	ensureIntegrationTest(t)

	pat := getRequiredValue(t, "AZURE_DEVOPS_PAT")
	service := NewService()

	type args struct {
		repositoryURLFormat string
		referenceName       string
		username            string
		password            string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Clone Azure DevOps repo branch",
			args: args{
				repositoryURLFormat: "https://:%s@portainer.visualstudio.com/Playground/_git/dev_integration",
				referenceName:       "refs/heads/main",
				username:            "",
				password:            pat,
			},
			wantErr: false,
		},
		{
			name: "Clone Azure DevOps repo tag",
			args: args{
				repositoryURLFormat: "https://:%s@portainer.visualstudio.com/Playground/_git/dev_integration",
				referenceName:       "refs/tags/v1.1",
				username:            "",
				password:            pat,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst, err := ioutils.TempDir("", "clone")
			assert.NoError(t, err)
			defer os.RemoveAll(dst)
			repositoryUrl := fmt.Sprintf(tt.args.repositoryURLFormat, tt.args.password)
			err = service.CloneRepository(dst, repositoryUrl, tt.args.referenceName, "", "")
			assert.NoError(t, err)
			assert.FileExists(t, filepath.Join(dst, "README.md"))
		})
	}
}

func TestService_ClonePrivateRepository_Azure(t *testing.T) {
	ensureIntegrationTest(t)

	pat := getRequiredValue(t, "AZURE_DEVOPS_PAT")
	service := NewService()

	dst, err := ioutils.TempDir("", "clone")
	assert.NoError(t, err)
	defer os.RemoveAll(dst)

	repositoryUrl := "https://portainer.visualstudio.com/Playground/_git/dev_integration"
	err = service.CloneRepository(dst, repositoryUrl, "refs/heads/main", "", pat)
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(dst, "README.md"))
}

func TestService_LatestCommitID_Azure(t *testing.T) {
	ensureIntegrationTest(t)

	pat := getRequiredValue(t, "AZURE_DEVOPS_PAT")
	service := NewService()

	repositoryUrl := "https://portainer.visualstudio.com/Playground/_git/dev_integration"
	id, err := service.LatestCommitID(repositoryUrl, "refs/heads/main", "", pat)
	assert.NoError(t, err)
	assert.NotEmpty(t, id, "cannot guarantee commit id, but it should be not empty")
}

func getRequiredValue(t *testing.T, name string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		t.Fatalf("can't find required env var \"%s\"", name)
	}
	return value
}

func ensureIntegrationTest(t *testing.T) {
	if _, ok := os.LookupEnv("INTEGRATION_TEST"); !ok {
		t.Skip("skip an integration test")
	}
}

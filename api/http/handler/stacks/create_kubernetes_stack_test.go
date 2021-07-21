package stacks

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

type git struct {
	content string
}

func (g *git) CloneRepository(destination string, repositoryURL, referenceName, username, password string) error {
	return g.ClonePublicRepository(repositoryURL, referenceName, destination)
}
func (g *git) ClonePublicRepository(repositoryURL string, referenceName string, destination string) error {
	return ioutil.WriteFile(path.Join(destination, "deployment.yml"), []byte(g.content), 0755)
}
func (g *git) ClonePrivateRepositoryWithBasicAuth(repositoryURL, referenceName string, destination, username, password string) error {
	return g.ClonePublicRepository(repositoryURL, referenceName, destination)
}

func (g *git) LatestCommitID(repositoryURL, referenceName, username, password string) (string, error) {
	return "", nil
}

func TestCloneAndConvertGitRepoFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "kube-create-stack")
	assert.NoError(t, err, "failed to create a tmp dir")
	defer os.RemoveAll(dir)

	content := `apiVersion: apps/v1
	kind: Deployment
	metadata:
		name: nginx-deployment
		labels:
			app: nginx
	spec:
		replicas: 3
		selector:
			matchLabels:
				app: nginx
		template:
			metadata:
				labels:
					app: nginx
			spec:
				containers:
				- name: nginx
					image: nginx:1.14.2
					ports:
					- containerPort: 80`

	h := &Handler{
		GitService: &git{
			content: content,
		},
	}
	gitInfo := &kubernetesGitDeploymentPayload{
		FilePathInRepository: "deployment.yml",
	}
	fileContent, err := h.cloneManifestContentFromGitRepo(gitInfo, dir)
	assert.NoError(t, err, "failed to clone or convert the file from Git repo")
	assert.Equal(t, content, fileContent)
}

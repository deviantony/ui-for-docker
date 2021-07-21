package cli

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	clientV1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

// compareYAMLStrings will compare 2 strings by stripping tabs, newlines and whitespaces from both strings
func compareYAMLStrings(in1, in2 string) int {
	r := strings.NewReplacer("\t", "", "\n", "", " ", "")
	in1 = r.Replace(in1)
	in2 = r.Replace(in2)
	return strings.Compare(in1, in2)
}

func Test_GenerateYAML(t *testing.T) {
	resourceYAMLTests := []struct {
		title    string
		resource runtime.Object
		wantYAML string
	}{
		{
			title: "Config",
			resource: &clientV1.Config{
				APIVersion:     "v1",
				Kind:           "Config",
				CurrentContext: "portainer-ctx",
				Contexts: []clientV1.NamedContext{
					{
						Name: "portainer-ctx",
						Context: clientV1.Context{
							AuthInfo: "test-user",
							Cluster:  "portainer-cluster",
						},
					},
				},
				Clusters: []clientV1.NamedCluster{
					{
						Name: "portainer-cluster",
						Cluster: clientV1.Cluster{
							Server:                "localhost",
							InsecureSkipTLSVerify: true,
						},
					},
				},
				AuthInfos: []clientV1.NamedAuthInfo{
					{
						Name: "test-user",
						AuthInfo: clientV1.AuthInfo{
							Token: "test-token",
						},
					},
				},
			},
			wantYAML: `
			apiVersion: v1
			clusters:
			- cluster:
					insecure-skip-tls-verify: true
					server: localhost
				name: portainer-cluster
			contexts:
			- context:
					cluster: portainer-cluster
					user: test-user
				name: portainer-ctx
			current-context: portainer-ctx
			kind: Config
			preferences: {}
			users:
			- name: test-user
				user:
					token: test-token
			`,
		},
	}

	for _, ryt := range resourceYAMLTests {
		t.Run(ryt.title, func(t *testing.T) {
			yaml, err := GenerateYAML(ryt.resource)
			if err != nil {
				t.Errorf("generateYamlConfig failed; err=%s", err)
			}

			if compareYAMLStrings(yaml, ryt.wantYAML) != 0 {
				t.Errorf("generateYamlConfig failed;\ngot=\n%s\nwant=\n%s", yaml, ryt.wantYAML)
			}
		})
	}
}

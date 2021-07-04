package gittypes

// RepoConfig represents a configuration for a repo
type RepoConfig struct {
	// The repo url
	URL string `example:"https://github.com/portainer/portainer-ee.git"`
	// The reference name
	ReferenceName string `example:"refs/heads/branch_name"`
	// Path to where the config file is in this url/refName
	ConfigFilePath string `example:"docker-compose.yml"`
}

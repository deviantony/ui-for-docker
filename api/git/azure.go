package git

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/portainer/portainer/api/archive"
)

const (
	azureDevOpsHost        = "dev.azure.com"
	visualStudioHostSuffix = ".visualstudio.com"
)

func isAzureUrl(s string) bool {
	return strings.Contains(s, azureDevOpsHost) ||
		strings.Contains(s, visualStudioHostSuffix)
}

type azureOptions struct {
	organisation, project, repository string
	// a user may pass credentials in a repository URL,
	// for example https://<username>:<password>@<domain>/<path>
	username, password string
}

type azureDownloader struct {
	client  *http.Client
	baseUrl string
}

func NewAzureDownloader(client *http.Client) *azureDownloader {
	return &azureDownloader{
		client:  client,
		baseUrl: "https://dev.azure.com",
	}
}

func (a *azureDownloader) download(ctx context.Context, destination string, options cloneOptions) error {
	zipFilepath, err := a.downloadZipFromAzureDevOps(ctx, options)
	if err != nil {
		return errors.Wrap(err, "failed to download a zip file from Azure DevOps")
	}
	defer os.Remove(zipFilepath)

	err = archive.UnzipFile(zipFilepath, destination)
	if err != nil {
		return errors.Wrap(err, "failed to unzip file")
	}

	return nil
}

func (a *azureDownloader) downloadZipFromAzureDevOps(ctx context.Context, options cloneOptions) (string, error) {
	config, err := parseUrl(options.repositoryUrl)
	if err != nil {
		return "", errors.WithMessage(err, "failed to parse url")
	}
	downloadUrl, err := a.buildDownloadUrl(config, options.referenceName)
	if err != nil {
		return "", errors.WithMessage(err, "failed to build download url")
	}
	zipFile, err := ioutil.TempFile("", "azure-git-repo-*.zip")
	if err != nil {
		return "", errors.WithMessage(err, "failed to create temp file")
	}
	defer zipFile.Close()

	req, err := http.NewRequestWithContext(ctx, "GET", downloadUrl, nil)
	if options.username != "" || options.password != "" {
		req.SetBasicAuth(options.username, options.password)
	} else if config.username != "" || config.password != "" {
		req.SetBasicAuth(config.username, config.password)
	}

	if err != nil {
		return "", errors.WithMessage(err, "failed to create a new HTTP request")
	}

	res, err := a.client.Do(req)
	if err != nil {
		return "", errors.WithMessage(err, "failed to make an HTTP request")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download zip with a status \"%v\"", res.Status)
	}

	_, err = io.Copy(zipFile, res.Body)
	if err != nil {
		return "", errors.WithMessage(err, "failed to save HTTP response to a file")
	}
	return zipFile.Name(), nil
}

func (a *azureDownloader) latestCommitID(ctx context.Context, options fetchOptions) (string, error) {
	config, err := parseUrl(options.repositoryUrl)
	if err != nil {
		return "", errors.WithMessage(err, "failed to parse url")
	}

	refsUrl, err := a.buildRefsUrl(config, options.referenceName)
	if err != nil {
		return "", errors.WithMessage(err, "failed to build azure refs url")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", refsUrl, nil)
	if options.username != "" || options.password != "" {
		req.SetBasicAuth(options.username, options.password)
	} else if config.username != "" || config.password != "" {
		req.SetBasicAuth(config.username, config.password)
	}

	if err != nil {
		return "", errors.WithMessage(err, "failed to create a new HTTP request")
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return "", errors.WithMessage(err, "failed to make an HTTP request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get repository refs with a status \"%v\"", resp.Status)
	}

	var refs struct {
		Value []struct {
			Name     string `json:"name"`
			ObjectId string `json:"objectId"`
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&refs); err != nil {
		return "", errors.Wrap(err, "could not parse Azure Refs response")
	}

	for _, ref := range refs.Value {
		if strings.EqualFold(ref.Name, options.referenceName) {
			return ref.ObjectId, nil
		}
	}

	return "", errors.Errorf("could not find ref %q in the repository", options.referenceName)
}

func parseUrl(rawUrl string) (*azureOptions, error) {
	if strings.HasPrefix(rawUrl, "https://") || strings.HasPrefix(rawUrl, "http://") {
		return parseHttpUrl(rawUrl)
	}
	if strings.HasPrefix(rawUrl, "git@ssh") {
		return parseSshUrl(rawUrl)
	}
	if strings.HasPrefix(rawUrl, "ssh://") {
		r := []rune(rawUrl)
		return parseSshUrl(string(r[6:])) // remove the prefix
	}

	return nil, errors.Errorf("supported url schemes are https and ssh; recevied URL %s rawUrl", rawUrl)
}

var expectedSshUrl = "git@ssh.dev.azure.com:v3/Organisation/Project/Repository"

func parseSshUrl(rawUrl string) (*azureOptions, error) {
	path := strings.Split(rawUrl, "/")

	unexpectedUrlErr := errors.Errorf("want url %s, got %s", expectedSshUrl, rawUrl)
	if len(path) != 4 {
		return nil, unexpectedUrlErr
	}
	return &azureOptions{
		organisation: path[1],
		project:      path[2],
		repository:   path[3],
	}, nil
}

const expectedAzureDevOpsHttpUrl = "https://Organisation@dev.azure.com/Organisation/Project/_git/Repository"
const expectedVisualStudioHttpUrl = "https://organisation.visualstudio.com/project/_git/repository"

func parseHttpUrl(rawUrl string) (*azureOptions, error) {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse HTTP url")
	}

	opt := azureOptions{}
	switch {
	case u.Host == azureDevOpsHost:
		path := strings.Split(u.Path, "/")
		if len(path) != 5 {
			return nil, errors.Errorf("want url %s, got %s", expectedAzureDevOpsHttpUrl, u)
		}
		opt.organisation = path[1]
		opt.project = path[2]
		opt.repository = path[4]
	case strings.HasSuffix(u.Host, visualStudioHostSuffix):
		path := strings.Split(u.Path, "/")
		if len(path) != 4 {
			return nil, errors.Errorf("want url %s, got %s", expectedVisualStudioHttpUrl, u)
		}
		opt.organisation = strings.TrimSuffix(u.Host, visualStudioHostSuffix)
		opt.project = path[1]
		opt.repository = path[3]
	default:
		return nil, errors.Errorf("unknown azure host in url \"%s\"", rawUrl)
	}

	opt.username = u.User.Username()
	opt.password, _ = u.User.Password()

	return &opt, nil
}

func (a *azureDownloader) buildDownloadUrl(config *azureOptions, referenceName string) (string, error) {
	rawUrl := fmt.Sprintf("%s/%s/%s/_apis/git/repositories/%s/items",
		a.baseUrl,
		url.PathEscape(config.organisation),
		url.PathEscape(config.project),
		url.PathEscape(config.repository))
	u, err := url.Parse(rawUrl)

	if err != nil {
		return "", errors.Wrapf(err, "failed to parse download url path %s", rawUrl)
	}
	q := u.Query()
	// scopePath=/&download=true&versionDescriptor.version=main&$format=zip&recursionLevel=full&api-version=6.0
	q.Set("scopePath", "/")
	q.Set("download", "true")
	q.Set("versionDescriptor.versionType", getVersionType(referenceName))
	q.Set("versionDescriptor.version", formatReferenceName(referenceName))
	q.Set("$format", "zip")
	q.Set("recursionLevel", "full")
	q.Set("api-version", "6.0")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (a *azureDownloader) buildRefsUrl(config *azureOptions, referenceName string) (string, error) {
	rawUrl := fmt.Sprintf("%s/%s/%s/_apis/git/repositories/%s/refs",
		a.baseUrl,
		url.PathEscape(config.organisation),
		url.PathEscape(config.project),
		url.PathEscape(config.repository))
	u, err := url.Parse(rawUrl)

	if err != nil {
		return "", errors.Wrapf(err, "failed to parse refs url path %s", rawUrl)
	}

	// filterContains=main&api-version=6.0
	q := u.Query()
	q.Set("filterContains", formatReferenceName(referenceName))
	q.Set("api-version", "6.0")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

const (
	branchPrefix = "refs/heads/"
	tagPrefix    = "refs/tags/"
)

func formatReferenceName(name string) string {
	if strings.HasPrefix(name, branchPrefix) {
		return strings.TrimPrefix(name, branchPrefix)
	}
	if strings.HasPrefix(name, tagPrefix) {
		return strings.TrimPrefix(name, tagPrefix)
	}
	return name
}

func getVersionType(name string) string {
	if strings.HasPrefix(name, branchPrefix) {
		return "branch"
	}
	if strings.HasPrefix(name, tagPrefix) {
		return "tag"
	}
	return "commit"
}

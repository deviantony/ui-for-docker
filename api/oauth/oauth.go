package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"

	portainer "github.com/portainer/portainer/api"
)

// Service represents a service used to authenticate users against an authorization server
type Service struct{}

// NewService returns a pointer to a new instance of this service
func NewService() *Service {
	return &Service{}
}

// Authenticate takes an access code and exchanges it for an access token from portainer OAuthSettings token endpoint.
// On success, it will then return the username and token expiry time associated to authenticated user by fetching this information
// from the resource server and matching it with the user identifier setting.
func (*Service) Authenticate(code string, configuration *portainer.OAuthSettings) (string, error) {
	token, err := getOAuthToken(code, configuration)
	if err != nil {
		log.Printf("[DEBUG] - Failed retrieving access token: %v", err)
		return "", err
	}
	username, err := getUsername(token.AccessToken, configuration)
	if err != nil {
		log.Printf("[DEBUG] - Failed retrieving oauth user name: %v", err)
		return "", err
	}
	return username, nil
}

func getOAuthToken(code string, configuration *portainer.OAuthSettings) (*oauth2.Token, error) {
	unescapedCode, err := url.QueryUnescape(code)
	if err != nil {
		return nil, err
	}

	config := buildConfig(configuration)
	token, err := config.Exchange(context.Background(), unescapedCode)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func getUsername(token string, configuration *portainer.OAuthSettings) (string, error) {
	req, err := http.NewRequest("GET", configuration.ResourceURI, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", &oauth2.RetrieveError{
			Response: resp,
			Body:     body,
		}
	}

	content, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return "", err
	}

	if content == "application/x-www-form-urlencoded" || content == "text/plain" {
		values, err := url.ParseQuery(string(body))
		if err != nil {
			return "", err
		}

		username := values.Get(configuration.UserIdentifier)
		if username == "" {
			return username, &oauth2.RetrieveError{
				Response: resp,
				Body:     body,
			}
		}

		return username, nil
	}

	var datamap map[string]interface{}
	if err = json.Unmarshal(body, &datamap); err != nil {
		return "", err
	}

	username, ok := datamap[configuration.UserIdentifier].(string)
	if ok && username != "" {
		return username, nil
	}

	if !ok {
		username, ok := datamap[configuration.UserIdentifier].(float64)
		if ok && username != 0 {
			return fmt.Sprint(int(username)), nil
		}
	}

	return "", &oauth2.RetrieveError{
		Response: resp,
		Body:     body,
	}
}

func buildConfig(configuration *portainer.OAuthSettings) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		AuthURL:  configuration.AuthorizationURI,
		TokenURL: configuration.AccessTokenURI,
	}

	return &oauth2.Config{
		ClientID:     configuration.ClientID,
		ClientSecret: configuration.ClientSecret,
		Endpoint:     endpoint,
		RedirectURL:  configuration.RedirectURI,
		Scopes:       []string{configuration.Scopes},
	}
}

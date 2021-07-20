package customtemplates

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/internal/authorization"
)

// @id CustomTemplateCreate
// @summary Create a custom template
// @description Create a custom template.
// @description **Access policy**: authenticated
// @tags custom_templates
// @security jwt
// @accept json,multipart/form-data
// @produce json
// @param method query string true "method for creating template" Enums(string, file, repository)
// @param body_string body customTemplateFromFileContentPayload false "Required when using method=string"
// @param body_repository body customTemplateFromGitRepositoryPayload false "Required when using method=repository"
// @param Title formData string false "Title of the template. required when method is file"
// @param Description formData string false "Description of the template. required when method is file"
// @param Note formData string false "A note that will be displayed in the UI. Supports HTML content"
// @param Platform formData int false "Platform associated to the template (1 - 'linux', 2 - 'windows'). required when method is file" Enums(1,2)
// @param Type formData int false "Type of created stack (1 - swarm, 2 - compose), required when method is file" Enums(1,2)
// @param file formData file false "required when method is file"
// @success 200 {object} portainer.CustomTemplate
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /custom_templates [post]
func (handler *Handler) customTemplateCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	method, err := request.RetrieveQueryParameter(r, "method", false)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid query parameter: method", err}
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve user details from authentication token", err}
	}

	customTemplate, err := handler.createCustomTemplate(method, r)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to create custom template", err}
	}

	customTemplate.CreatedByUserID = tokenData.ID

	customTemplates, err := handler.DataStore.CustomTemplate().CustomTemplates()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve custom templates from the database", err}
	}

	for _, existingTemplate := range customTemplates {
		if existingTemplate.Title == customTemplate.Title {
			return &httperror.HandlerError{http.StatusInternalServerError, "Template name must be unique", errors.New("Template name must be unique")}
		}
	}

	err = handler.DataStore.CustomTemplate().CreateCustomTemplate(customTemplate)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to create custom template", err}
	}

	resourceControl := authorization.NewPrivateResourceControl(strconv.Itoa(int(customTemplate.ID)), portainer.CustomTemplateResourceControl, tokenData.ID)

	err = handler.DataStore.ResourceControl().CreateResourceControl(resourceControl)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist resource control inside the database", err}
	}

	customTemplate.ResourceControl = resourceControl

	return response.JSON(w, customTemplate)
}

func (handler *Handler) createCustomTemplate(method string, r *http.Request) (*portainer.CustomTemplate, error) {
	switch method {
	case "string":
		return handler.createCustomTemplateFromFileContent(r)
	case "repository":
		return handler.createCustomTemplateFromGitRepository(r)
	case "file":
		return handler.createCustomTemplateFromFileUpload(r)
	}
	return nil, errors.New("Invalid value for query parameter: method. Value must be one of: string, repository or file")
}

type customTemplateFromFileContentPayload struct {
	// URL of the template's logo
	Logo string `example:"https://cloudinovasi.id/assets/img/logos/nginx.png"`
	// Title of the template
	Title string `example:"Nginx" validate:"required"`
	// Description of the template
	Description string `example:"High performance web server" validate:"required"`
	// A note that will be displayed in the UI. Supports HTML content
	Note string `example:"This is my <b>custom</b> template"`
	// Platform associated to the template.
	// Valid values are: 1 - 'linux', 2 - 'windows'
	Platform portainer.CustomTemplatePlatform `example:"1" enums:"1,2" validate:"required"`
	// Type of created stack (1 - swarm, 2 - compose)
	Type portainer.StackType `example:"1" enums:"1,2" validate:"required"`
	// Content of stack file
	FileContent string `validate:"required"`
}

func (payload *customTemplateFromFileContentPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Title) {
		return errors.New("Invalid custom template title")
	}
	if govalidator.IsNull(payload.Description) {
		return errors.New("Invalid custom template description")
	}
	if govalidator.IsNull(payload.FileContent) {
		return errors.New("Invalid file content")
	}
	if payload.Platform != portainer.CustomTemplatePlatformLinux && payload.Platform != portainer.CustomTemplatePlatformWindows {
		return errors.New("Invalid custom template platform")
	}
	if payload.Type != portainer.DockerSwarmStack && payload.Type != portainer.DockerComposeStack {
		return errors.New("Invalid custom template type")
	}
	return nil
}

func (handler *Handler) createCustomTemplateFromFileContent(r *http.Request) (*portainer.CustomTemplate, error) {
	var payload customTemplateFromFileContentPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return nil, err
	}

	customTemplateID := handler.DataStore.CustomTemplate().GetNextIdentifier()
	customTemplate := &portainer.CustomTemplate{
		ID:          portainer.CustomTemplateID(customTemplateID),
		Title:       payload.Title,
		EntryPoint:  filesystem.ComposeFileDefaultName,
		Description: payload.Description,
		Note:        payload.Note,
		Platform:    (payload.Platform),
		Type:        (payload.Type),
		Logo:        payload.Logo,
	}

	templateFolder := strconv.Itoa(customTemplateID)
	projectPath, err := handler.FileService.StoreCustomTemplateFileFromBytes(templateFolder, customTemplate.EntryPoint, []byte(payload.FileContent))
	if err != nil {
		return nil, err
	}
	customTemplate.ProjectPath = projectPath

	return customTemplate, nil
}

type customTemplateFromGitRepositoryPayload struct {
	// URL of the template's logo
	Logo string `example:"https://cloudinovasi.id/assets/img/logos/nginx.png"`
	// Title of the template
	Title string `example:"Nginx" validate:"required"`
	// Description of the template
	Description string `example:"High performance web server" validate:"required"`
	// A note that will be displayed in the UI. Supports HTML content
	Note string `example:"This is my <b>custom</b> template"`
	// Platform associated to the template.
	// Valid values are: 1 - 'linux', 2 - 'windows'
	Platform portainer.CustomTemplatePlatform `example:"1" enums:"1,2" validate:"required"`
	// Type of created stack (1 - swarm, 2 - compose)
	Type portainer.StackType `example:"1" enums:"1,2" validate:"required"`

	// URL of a Git repository hosting the Stack file
	RepositoryURL string `example:"https://github.com/openfaas/faas" validate:"required"`
	// Reference name of a Git repository hosting the Stack file
	RepositoryReferenceName string `example:"refs/heads/master"`
	// Use basic authentication to clone the Git repository
	RepositoryAuthentication bool `example:"true"`
	// Username used in basic authentication. Required when RepositoryAuthentication is true.
	RepositoryUsername string `example:"myGitUsername"`
	// Password used in basic authentication. Required when RepositoryAuthentication is true.
	RepositoryPassword string `example:"myGitPassword"`
	// Path to the Stack file inside the Git repository
	ComposeFilePathInRepository string `example:"docker-compose.yml" default:"docker-compose.yml"`
}

func (payload *customTemplateFromGitRepositoryPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Title) {
		return errors.New("Invalid custom template title")
	}
	if govalidator.IsNull(payload.Description) {
		return errors.New("Invalid custom template description")
	}
	if govalidator.IsNull(payload.RepositoryURL) || !govalidator.IsURL(payload.RepositoryURL) {
		return errors.New("Invalid repository URL. Must correspond to a valid URL format")
	}
	if payload.RepositoryAuthentication && (govalidator.IsNull(payload.RepositoryUsername) || govalidator.IsNull(payload.RepositoryPassword)) {
		return errors.New("Invalid repository credentials. Username and password must be specified when authentication is enabled")
	}
	if govalidator.IsNull(payload.ComposeFilePathInRepository) {
		payload.ComposeFilePathInRepository = filesystem.ComposeFileDefaultName
	}
	if payload.Platform != portainer.CustomTemplatePlatformLinux && payload.Platform != portainer.CustomTemplatePlatformWindows {
		return errors.New("Invalid custom template platform")
	}
	if payload.Type != portainer.DockerSwarmStack && payload.Type != portainer.DockerComposeStack {
		return errors.New("Invalid custom template type")
	}
	return nil
}

func (handler *Handler) createCustomTemplateFromGitRepository(r *http.Request) (*portainer.CustomTemplate, error) {
	var payload customTemplateFromGitRepositoryPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return nil, err
	}

	customTemplateID := handler.DataStore.CustomTemplate().GetNextIdentifier()
	customTemplate := &portainer.CustomTemplate{
		ID:          portainer.CustomTemplateID(customTemplateID),
		Title:       payload.Title,
		EntryPoint:  payload.ComposeFilePathInRepository,
		Description: payload.Description,
		Note:        payload.Note,
		Platform:    payload.Platform,
		Type:        payload.Type,
		Logo:        payload.Logo,
	}

	projectPath := handler.FileService.GetCustomTemplateProjectPath(strconv.Itoa(customTemplateID))
	customTemplate.ProjectPath = projectPath

	repositoryUsername := payload.RepositoryUsername
	repositoryPassword := payload.RepositoryPassword
	if !payload.RepositoryAuthentication {
		repositoryUsername = ""
		repositoryPassword = ""
	}

	err = handler.GitService.CloneRepository(projectPath, payload.RepositoryURL, payload.RepositoryReferenceName, repositoryUsername, repositoryPassword)
	if err != nil {
		return nil, err
	}

	return customTemplate, nil
}

type customTemplateFromFileUploadPayload struct {
	Logo        string
	Title       string
	Description string
	Note        string
	Platform    portainer.CustomTemplatePlatform
	Type        portainer.StackType
	FileContent []byte
}

func (payload *customTemplateFromFileUploadPayload) Validate(r *http.Request) error {
	title, err := request.RetrieveMultiPartFormValue(r, "Title", false)
	if err != nil {
		return errors.New("Invalid custom template title")
	}
	payload.Title = title

	description, err := request.RetrieveMultiPartFormValue(r, "Description", false)
	if err != nil {
		return errors.New("Invalid custom template description")
	}

	payload.Description = description

	note, _ := request.RetrieveMultiPartFormValue(r, "Note", true)
	payload.Note = note

	platform, _ := request.RetrieveNumericMultiPartFormValue(r, "Platform", true)
	templatePlatform := portainer.CustomTemplatePlatform(platform)
	if templatePlatform != portainer.CustomTemplatePlatformLinux && templatePlatform != portainer.CustomTemplatePlatformWindows {
		return errors.New("Invalid custom template platform")
	}
	payload.Platform = templatePlatform

	typeNumeral, _ := request.RetrieveNumericMultiPartFormValue(r, "Type", true)
	templateType := portainer.StackType(typeNumeral)
	if templateType != portainer.DockerComposeStack && templateType != portainer.DockerSwarmStack {
		return errors.New("Invalid custom template type")
	}
	payload.Type = templateType

	composeFileContent, _, err := request.RetrieveMultiPartFormFile(r, "File")
	if err != nil {
		return errors.New("Invalid Compose file. Ensure that the Compose file is uploaded correctly")
	}
	payload.FileContent = composeFileContent

	return nil
}

func (handler *Handler) createCustomTemplateFromFileUpload(r *http.Request) (*portainer.CustomTemplate, error) {
	payload := &customTemplateFromFileUploadPayload{}
	err := payload.Validate(r)
	if err != nil {
		return nil, err
	}

	customTemplateID := handler.DataStore.CustomTemplate().GetNextIdentifier()
	customTemplate := &portainer.CustomTemplate{
		ID:          portainer.CustomTemplateID(customTemplateID),
		Title:       payload.Title,
		Description: payload.Description,
		Note:        payload.Note,
		Platform:    payload.Platform,
		Type:        payload.Type,
		Logo:        payload.Logo,
		EntryPoint:  filesystem.ComposeFileDefaultName,
	}

	templateFolder := strconv.Itoa(customTemplateID)
	projectPath, err := handler.FileService.StoreCustomTemplateFileFromBytes(templateFolder, customTemplate.EntryPoint, []byte(payload.FileContent))
	if err != nil {
		return nil, err
	}
	customTemplate.ProjectPath = projectPath

	return customTemplate, nil
}

import angular from 'angular';
import _ from 'lodash-es';

import { AccessControlFormData } from '../../../components/accessControlForm/porAccessControlFormModel';

angular
  .module('portainer.app')
  .controller('CreateStackController', function (
    $scope,
    $state,
    $async,
    $window,
    ModalService,
    StackService,
    Authentication,
    Notifications,
    FormValidator,
    ResourceControlService,
    FormHelper,
    EndpointProvider,
    StackHelper,
    ContainerHelper,
    CustomTemplateService,
    ContainerService
  ) {
    $scope.formValues = {
      Name: '',
      StackFileContent: '',
      StackFile: null,
      RepositoryURL: '',
      RepositoryReferenceName: '',
      RepositoryAuthentication: false,
      RepositoryUsername: '',
      RepositoryPassword: '',
      Env: [],
      ComposeFilePathInRepository: 'docker-compose.yml',
      AccessControlData: new AccessControlFormData(),
    };

    $scope.state = {
      Method: 'editor',
      formValidationError: '',
      actionInProgress: false,
      StackType: null,
      editorYamlValidationError: '',
      uploadYamlValidationError: '',
      isEditorDirty: false,
    };

    $window.onbeforeunload = () => {
      if ($scope.state.Method === 'editor' && $scope.formValues.StackFileContent && $scope.state.isEditorDirty) {
        return '';
      }
    };

    $scope.$on('$destroy', function() {
      $scope.state.isEditorDirty = false;
    })

    $scope.onChangeFormValues = onChangeFormValues;

    $scope.addEnvironmentVariable = function () {
      $scope.formValues.Env.push({ name: '', value: '' });
    };

    $scope.removeEnvironmentVariable = function (index) {
      $scope.formValues.Env.splice(index, 1);
    };

    function validateForm(accessControlData, isAdmin) {
      $scope.state.formValidationError = '';
      var error = '';
      error = FormValidator.validateAccessControl(accessControlData, isAdmin);

      if (error) {
        $scope.state.formValidationError = error;
        return false;
      }
      return true;
    }

    function createSwarmStack(name, method) {
      var env = FormHelper.removeInvalidEnvVars($scope.formValues.Env);
      const endpointId = +$state.params.endpointId;

      if (method === 'template' || method === 'editor') {
        var stackFileContent = $scope.formValues.StackFileContent;
        return StackService.createSwarmStackFromFileContent(name, stackFileContent, env, endpointId);
      }

      if (method === 'upload') {
        var stackFile = $scope.formValues.StackFile;
        return StackService.createSwarmStackFromFileUpload(name, stackFile, env, endpointId);
      }

      if (method === 'repository') {
        var repositoryOptions = {
          RepositoryURL: $scope.formValues.RepositoryURL,
          RepositoryReferenceName: $scope.formValues.RepositoryReferenceName,
          ComposeFilePathInRepository: $scope.formValues.ComposeFilePathInRepository,
          RepositoryAuthentication: $scope.formValues.RepositoryAuthentication,
          RepositoryUsername: $scope.formValues.RepositoryUsername,
          RepositoryPassword: $scope.formValues.RepositoryPassword,
        };
        return StackService.createSwarmStackFromGitRepository(name, repositoryOptions, env, endpointId);
      }
    }

    function createComposeStack(name, method) {
      var env = FormHelper.removeInvalidEnvVars($scope.formValues.Env);
      const endpointId = +$state.params.endpointId;

      if (method === 'editor' || method === 'template') {
        var stackFileContent = $scope.formValues.StackFileContent;
        return StackService.createComposeStackFromFileContent(name, stackFileContent, env, endpointId);
      } else if (method === 'upload') {
        var stackFile = $scope.formValues.StackFile;
        return StackService.createComposeStackFromFileUpload(name, stackFile, env, endpointId);
      } else if (method === 'repository') {
        var repositoryOptions = {
          RepositoryURL: $scope.formValues.RepositoryURL,
          RepositoryReferenceName: $scope.formValues.RepositoryReferenceName,
          ComposeFilePathInRepository: $scope.formValues.ComposeFilePathInRepository,
          RepositoryAuthentication: $scope.formValues.RepositoryAuthentication,
          RepositoryUsername: $scope.formValues.RepositoryUsername,
          RepositoryPassword: $scope.formValues.RepositoryPassword,
        };
        return StackService.createComposeStackFromGitRepository(name, repositoryOptions, env, endpointId);
      }
    }

    $scope.handleEnvVarChange = handleEnvVarChange;
    function handleEnvVarChange(value) {
      $scope.formValues.Env = value;
    }

    $scope.deployStack = function () {
      var name = $scope.formValues.Name;
      var method = $scope.state.Method;

      var accessControlData = $scope.formValues.AccessControlData;
      var userDetails = Authentication.getUserDetails();
      var isAdmin = Authentication.isAdmin();

      if (method === 'editor' && $scope.formValues.StackFileContent === '') {
        $scope.state.formValidationError = 'Stack file content must not be empty';
        return;
      }

      if (!validateForm(accessControlData, isAdmin)) {
        return;
      }

      var type = $scope.state.StackType;
      var action = createSwarmStack;
      if (type === 2) {
        action = createComposeStack;
      }
      $scope.state.actionInProgress = true;
      action(name, method)
        .then(function success(data) {
          if (data.data) {
            data = data.data;
          }
          const userId = userDetails.ID;
          const resourceControl = data.ResourceControl;
          return ResourceControlService.applyResourceControl(userId, accessControlData, resourceControl);
        })
        .then(function success() {
          Notifications.success('Stack successfully deployed');
          $scope.state.isEditorDirty = false;
          $state.go('docker.stacks');
        })
        .catch(function error(err) {
          Notifications.error('Deployment error', err, 'Unable to deploy stack');
        })
        .finally(function final() {
          $scope.state.actionInProgress = false;
        });
    };

    $scope.editorUpdate = function (cm) {
      $scope.formValues.StackFileContent = cm.getValue();
      $scope.state.editorYamlValidationError = StackHelper.validateYAML($scope.formValues.StackFileContent, $scope.containerNames);
      $scope.state.isEditorDirty = true;
    };

    async function onFileLoadAsync(event) {
      $scope.state.uploadYamlValidationError = StackHelper.validateYAML(event.target.result, $scope.containerNames);
    }

    function onFileLoad(event) {
      return $async(onFileLoadAsync, event);
    }

    $scope.uploadFile = function (file) {
      $scope.formValues.StackFile = file;

      if (file) {
        const temporaryFileReader = new FileReader();
        temporaryFileReader.fileName = file.name;
        temporaryFileReader.onload = onFileLoad;
        temporaryFileReader.readAsText(file);
      }
    };

    $scope.onChangeTemplate = async function onChangeTemplate(template) {
      try {
        $scope.formValues.StackFileContent = undefined;
        $scope.selectedTemplate = template;
        $scope.formValues.StackFileContent = await CustomTemplateService.customTemplateFile(template.Id);
      } catch (err) {
        Notifications.error('Failure', err, 'Unable to retrieve Custom Template file');
      }
    };

    async function initView() {
      var endpointMode = $scope.applicationState.endpoint.mode;
      $scope.state.StackType = 2;
      if (endpointMode.provider === 'DOCKER_SWARM_MODE' && endpointMode.role === 'MANAGER') {
        $scope.state.StackType = 1;
      }

      try {
        const templates = await CustomTemplateService.customTemplates($scope.state.StackType);
        $scope.templates = _.map(templates, (template) => ({ ...template, label: `${template.Title} - ${template.Description}` }));
      } catch (err) {
        Notifications.error('Failure', err, 'Unable to retrieve Custom Templates');
      }

      try {
        const endpoint = EndpointProvider.currentEndpoint();
        $scope.composeSyntaxMaxVersion = endpoint.ComposeSyntaxMaxVersion;
      } catch (err) {
        Notifications.error('Failure', err, 'Unable to retrieve the ComposeSyntaxMaxVersion');
      }

      try {
        $scope.containers = await ContainerService.containers();
        $scope.containerNames = ContainerHelper.getContainerNames($scope.containers);
      } catch (err) {
        Notifications.error('Failure', err, 'Unable to retrieve Containers');
      }
    }

    this.uiCanExit = async function () {
      if ($scope.state.Method === 'editor' && $scope.formValues.StackFileContent && $scope.state.isEditorDirty) {
        return ModalService.confirmWebEditorDiscard();
      }
    };

    initView();

    function onChangeFormValues(newValues) {
      $scope.formValues = newValues;
    }
  });

import _ from 'lodash-es';
import { AccessControlFormData } from 'Portainer/components/accessControlForm/porAccessControlFormModel';

class CustomTemplatesViewController {
  /* @ngInject */
  constructor(
    $anchorScroll,
    $async,
    $rootScope,
    $state,
    Authentication,
    CustomTemplateService,
    EndpointProvider,
    FormValidator,
    ModalService,
    NetworkService,
    Notifications,
    ResourceControlService,
    StackService,
    StateManager
  ) {
    this.$anchorScroll = $anchorScroll;
    this.$async = $async;
    this.$rootScope = $rootScope;
    this.$state = $state;
    this.Authentication = Authentication;
    this.CustomTemplateService = CustomTemplateService;
    this.EndpointProvider = EndpointProvider;
    this.FormValidator = FormValidator;
    this.ModalService = ModalService;
    this.NetworkService = NetworkService;
    this.Notifications = Notifications;
    this.ResourceControlService = ResourceControlService;
    this.StateManager = StateManager;
    this.StackService = StackService;

    this.DOCKER_STANDALONE = 'DOCKER_STANDALONE';
    this.DOCKER_SWARM_MODE = 'DOCKER_SWARM_MODE';

    this.state = {
      selectedTemplate: null,
      showAdvancedOptions: false,
      formValidationError: '',
      actionInProgress: false,
      isEditorVisible: false,
      deployable: false,
    };

    this.currentUser = {
      isAdmin: false,
      id: null,
    };

    this.formValues = {
      network: '',
      name: '',
      fileContent: '',
      AccessControlData: new AccessControlFormData(),
    };

    this.getTemplates = this.getTemplates.bind(this);
    this.getTemplatesAsync = this.getTemplatesAsync.bind(this);
    this.removeTemplates = this.removeTemplates.bind(this);
    this.removeTemplatesAsync = this.removeTemplatesAsync.bind(this);
    this.validateForm = this.validateForm.bind(this);
    this.createStack = this.createStack.bind(this);
    this.createStackAsync = this.createStackAsync.bind(this);
    this.selectTemplate = this.selectTemplate.bind(this);
    this.selectTemplateAsync = this.selectTemplateAsync.bind(this);
    this.unselectTemplate = this.unselectTemplate.bind(this);
    this.unselectTemplateAsync = this.unselectTemplateAsync.bind(this);
    this.getNetworks = this.getNetworks.bind(this);
    this.getNetworksAsync = this.getNetworksAsync.bind(this);
    this.confirmDelete = this.confirmDelete.bind(this);
    this.confirmDeleteAsync = this.confirmDeleteAsync.bind(this);
    this.editorUpdate = this.editorUpdate.bind(this);
    this.isEditAllowed = this.isEditAllowed.bind(this);
  }

  isEditAllowed(template) {
    return this.currentUser.isAdmin || this.currentUser.id === template.CreatedByUserId;
  }

  getTemplates(endpointMode) {
    return this.$async(this.getTemplatesAsync, endpointMode);
  }
  async getTemplatesAsync({ provider, role }) {
    try {
      let stackType = 2;
      if (provider === 'DOCKER_SWARM_MODE' && role === 'MANAGER') {
        stackType = 1;
      }
      this.templates = await this.CustomTemplateService.customTemplates(stackType);
    } catch (err) {
      this.Notifications.error('Failed loading templates', err, 'Unable to load custom templates');
    }
  }

  removeTemplates(templates) {
    return this.$async(this.removeTemplatesAsync, templates);
  }
  async removeTemplatesAsync(templates) {
    for (let template of templates) {
      try {
        await this.CustomTemplateService.remove(template.id);
        this.Notifications.success('Removed template successfully');
        _.remove(this.templates, template);
      } catch (err) {
        this.Notifications.error('Failed removing template', err, 'Unable to remove custom template');
      }
    }
  }

  validateForm(accessControlData, isAdmin) {
    this.state.formValidationError = '';
    const error = this.FormValidator.validateAccessControl(accessControlData, isAdmin);

    if (error) {
      this.state.formValidationError = error;
      return false;
    }
    return true;
  }

  createStack() {
    return this.$async(this.createStackAsync);
  }
  async createStackAsync() {
    const userId = this.currentUser.id;
    const accessControlData = this.formValues.AccessControlData;

    if (!this.validateForm(accessControlData, this.currentUser.isAdmin)) {
      return;
    }
    const stackName = this.formValues.name;

    const endpointId = this.EndpointProvider.endpointID();

    this.state.actionInProgress = true;

    try {
      const file = this.formValues.fileContent;
      const createAction = this.state.selectedTemplate.Type === 1 ? this.StackService.createSwarmStackFromFileContent : this.StackService.createComposeStackFromFileContent;
      const { ResourceControl: resourceControl } = await createAction(stackName, file, [], endpointId);
      await this.ResourceControlService.applyResourceControl(userId, accessControlData, resourceControl);
      this.Notifications.success('Stack successfully deployed');
      this.$state.go('docker.stacks');
    } catch (err) {
      this.Notifications.error('Deployment error', err, 'Failed to deploy stack');
    } finally {
      this.state.actionInProgress = false;
    }
  }

  unselectTemplate(template) {
    // wrapping unselect with async to make a digest cycle run between unselect to select
    return this.$async(this.unselectTemplateAsync, template);
  }
  async unselectTemplateAsync(template) {
    template.Selected = false;
    this.state.selectedTemplate = null;

    this.formValues = {
      network: '',
      name: '',
      fileContent: '',
      AccessControlData: new AccessControlFormData(),
    };
  }

  selectTemplate(template) {
    return this.$async(this.selectTemplateAsync, template);
  }
  async selectTemplateAsync(template) {
    if (this.state.selectedTemplate) {
      await this.unselectTemplate(this.state.selectedTemplate);
    }

    template.Selected = true;

    this.formValues.network = _.find(this.availableNetworks, function (o) {
      return o.Name === 'bridge';
    });

    this.formValues.name = template.Title ? template.Title : '';
    this.state.selectedTemplate = template;
    const applicationState = this.StateManager.getState();
    this.state.deployable = this.isDeployable(applicationState.endpoint, template.Type);
    this.$anchorScroll('view-top');

    const file = await this.CustomTemplateService.customTemplateFile(template.Id);
    this.formValues.fileContent = file;
  }

  isDeployable(endpoint, templateType) {
    let deployable = false;
    switch (templateType) {
      case 1:
        deployable = endpoint.mode.provider === this.DOCKER_SWARM_MODE;
        break;
      case 2:
        deployable = endpoint.mode.provider === this.DOCKER_STANDALONE;
        break;
    }

    return deployable;
  }

  getNetworks(provider, apiVersion) {
    return this.$async(this.getNetworksAsync, provider, apiVersion);
  }
  async getNetworksAsync(provider, apiVersion) {
    try {
      const networks = await this.NetworkService.networks(
        provider === 'DOCKER_STANDALONE' || provider === 'DOCKER_SWARM_MODE',
        false,
        provider === 'DOCKER_SWARM_MODE' && apiVersion >= 1.25
      );
      this.availableNetworks = networks;
    } catch (err) {
      this.Notifications.error('Failure', err, 'Failed to load networks.');
    }
  }

  confirmDelete(templateId) {
    return this.$async(this.confirmDeleteAsync, templateId);
  }
  async confirmDeleteAsync(templateId) {
    const confirmed = await this.ModalService.confirmDeletionAsync('Are you sure that you want to delete this template?');
    if (!confirmed) {
      return;
    }

    try {
      await this.CustomTemplateService.remove(templateId);
      _.remove(this.templates, { Id: templateId });
    } catch (err) {
      this.Notifications.error('Failure', err, 'Failed to delete template');
    }
  }

  editorUpdate(cm) {
    this.formValues.fileContent = cm.getValue();
  }

  $onInit() {
    const applicationState = this.StateManager.getState();

    const {
      endpoint: { mode: endpointMode },
      apiVersion,
    } = applicationState;

    this.getTemplates(endpointMode);
    this.getNetworks(endpointMode.provider, apiVersion);

    this.currentUser.isAdmin = this.Authentication.isAdmin();
    const user = this.Authentication.getUserDetails();
    this.currentUser.id = user.ID;
  }
}

export default CustomTemplatesViewController;

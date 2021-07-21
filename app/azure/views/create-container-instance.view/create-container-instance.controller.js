import { ContainerGroupDefaultModel } from '@/azure/models/container_group';

export default class CreateContainerInstanceViewController {
  /* @ngInject */
  constructor($async, $q, $state, AzureService, Notifications, Authentication, ResourceControlService, FormValidator) {
    Object.assign(this, { $async, $q, $state, AzureService, Notifications, Authentication, ResourceControlService, FormValidator });
    this.allResourceGroups = [];
    this.allProviders = [];

    this.state = {
      actionInProgress: false,
      selectedSubscription: null,
      selectedResourceGroup: null,
      formValidationError: '',
    };

    this.changeSubscription = this.changeSubscription.bind(this);
    this.addPortBinding = this.addPortBinding.bind(this);
    this.removePortBinding = this.removePortBinding.bind(this);
    this.create = this.create.bind(this);
    this.updateResourceGroupsAndLocations = this.updateResourceGroupsAndLocations.bind(this);
    this.applyResourceControl = this.applyResourceControl.bind(this);
    this.validateForm = this.validateForm.bind(this);
  }

  changeSubscription() {
    const selectedSubscription = this.state.selectedSubscription;
    this.updateResourceGroupsAndLocations(selectedSubscription, this.allResourceGroups, this.allProviders);
  }

  addPortBinding() {
    this.model.Ports.push({ host: '', container: '', protocol: 'TCP' });
  }

  removePortBinding(index) {
    this.model.Ports.splice(index, 1);
  }

  create() {
    return this.$async(async () => {
      const model = this.model;
      const subscriptionId = this.state.selectedSubscription.Id;
      const resourceGroupName = this.state.selectedResourceGroup.Name;

      this.state.formValidationError = this.validateForm(model);
      if (this.state.formValidationError) {
        return false;
      }

      this.state.actionInProgress = true;
      try {
        const resourceGroup = await this.AzureService.createContainerGroup(model, subscriptionId, resourceGroupName);
        await this.applyResourceControl(resourceGroup);
        this.Notifications.success('Container successfully created', model.Name);
        this.$state.go('azure.containerinstances');
      } catch (err) {
        const newErr = err.data ? err.data.error : err;
        this.Notifications.error('Failure', newErr, 'Unable to create container');
      } finally {
        this.state.actionInProgress = false;
      }
    });
  }

  updateResourceGroupsAndLocations(subscription, resourceGroups, providers) {
    this.state.selectedResourceGroup = resourceGroups[subscription.Id][0];
    this.resourceGroups = resourceGroups[subscription.Id];

    const currentSubLocations = providers[subscription.Id].Locations;
    this.model.Location = currentSubLocations[0];
    this.locations = currentSubLocations;
  }

  async $onInit() {
    this.model = new ContainerGroupDefaultModel();

    try {
      const subscriptions = await this.AzureService.subscriptions();
      this.state.selectedSubscription = subscriptions[0];
      this.subscriptions = subscriptions;

      const [resourceGroups, containerInstancesProviders] = await Promise.all([
        this.AzureService.resourceGroups(subscriptions),
        this.AzureService.containerInstanceProvider(subscriptions),
      ]);

      this.allResourceGroups = resourceGroups;
      this.allProviders = containerInstancesProviders;
      const selectedSubscription = this.state.selectedSubscription;
      this.updateResourceGroupsAndLocations(selectedSubscription, resourceGroups, containerInstancesProviders);
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve Azure resources');
    }
  }

  applyResourceControl(newResourceGroup) {
    const userId = this.Authentication.getUserDetails().ID;
    const resourceControl = newResourceGroup.Portainer.ResourceControl;
    const accessControlData = this.model.AccessControlData;

    return this.ResourceControlService.applyResourceControl(userId, accessControlData, resourceControl);
  }

  validateForm(model) {
    if (!model.Ports || !model.Ports.length || model.Ports.every((port) => !port.host || !port.container)) {
      return 'At least one port binding is required';
    }

    const error = this.FormValidator.validateAccessControl(model.AccessControlData, this.Authentication.isAdmin());
    if (error !== '') {
      return error;
    }

    return null;
  }
}

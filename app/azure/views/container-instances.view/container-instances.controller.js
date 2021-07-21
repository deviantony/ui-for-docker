import _ from 'lodash-es';

export default class ContainerInstancesViewController {
  /* @ngInject */
  constructor($async, $state, AzureService, Notifications) {
    Object.assign(this, { $async, $state, AzureService, Notifications });

    this.containerGroups = [];

    this.deleteAction = this.deleteAction.bind(this);
  }

  deleteAction(selectedItems) {
    return this.$async(async () => {
      for (let item of selectedItems) {
        try {
          await this.AzureService.deleteContainerGroup(item.Id);
          this.Notifications.success('Container group successfully removed', item.Name);

          _.remove(this.containerGroups, item);
        } catch (err) {
          this.Notifications.error('Failure', err, 'Unable to remove container group');
        }
      }
    });
  }

  async $onInit() {
    try {
      const subscriptions = await this.AzureService.subscriptions();
      const containerGroups = await this.AzureService.containerGroups(subscriptions);
      this.containerGroups = this.AzureService.aggregate(containerGroups);
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to load container groups');
    }
  }
}

export default class DashboardController {
  /* @ngInject */
  constructor(AzureService, Notifications) {
    Object.assign(this, { AzureService, Notifications });
  }

  async $onInit() {
    try {
      const subscriptions = await this.AzureService.subscriptions();
      this.subscriptions = subscriptions;

      const resourceGroups = this.AzureService.resourceGroups(subscriptions);
      this.resourceGroups = this.AzureService.aggregate(resourceGroups);
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to load dashboard data');
    }
  }
}

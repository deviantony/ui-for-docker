import angular from 'angular';

class KubernetesNodesController {
  /* @ngInject */
  constructor($async, Notifications, KubernetesNodeService) {
    this.$async = $async;
    this.Notifications = Notifications;
    this.KubernetesNodeService = KubernetesNodeService;

    this.getNodes = this.getNodes.bind(this);
    this.getNodesAsync = this.getNodesAsync.bind(this);
  }

  async getNodesAsync() {
    try {
      this.nodes = await this.KubernetesNodeService.nodes();
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve nodes');
    }
  }

  getNodes() {
    return this.$async(this.getNodesAsync);
  }

  async $onInit() {
    this.getNodes();
  }
}

export default KubernetesNodesController;
angular.module('portainer.kubernetes').controller('KubernetesNodesController', KubernetesNodesController);

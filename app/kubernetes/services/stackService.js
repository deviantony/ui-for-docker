import _ from 'lodash-es';
import angular from 'angular';

class KubernetesStackService {
  /* @ngInject */
  constructor($async, KubernetesApplicationService) {
    this.$async = $async;
    this.KubernetesApplicationService = KubernetesApplicationService;

    this.getAllAsync = this.getAllAsync.bind(this);
  }

  /**
   * GET
   */
  async getAllAsync(namespace) {
    const applications = await this.KubernetesApplicationService.get(namespace);
    const stacks = _.map(applications, (item) => item.StackName);
    return _.uniq(_.without(stacks, '-'));
  }

  get(namespace) {
    return this.$async(this.getAllAsync, namespace);
  }
}

export default KubernetesStackService;
angular.module('portainer.kubernetes').service('KubernetesStackService', KubernetesStackService);

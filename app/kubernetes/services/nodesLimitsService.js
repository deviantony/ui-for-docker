import angular from 'angular';
import PortainerError from 'Portainer/error';
import { KubernetesNodesLimits } from 'Kubernetes/models/nodes-limits/models';

class KubernetesNodesLimitsService {
  /* @ngInject */
  constructor($async, KubernetesNodesLimits) {
    this.$async = $async;
    this.KubernetesNodesLimits = KubernetesNodesLimits;

    this.getAsync = this.getAsync.bind(this);
  }

  /**
   * GET
   */
  async getAsync() {
    try {
      const nodesLimits = await this.KubernetesNodesLimits().get().$promise;
      return new KubernetesNodesLimits(nodesLimits.data);
    } catch (err) {
      throw new PortainerError('Unable to retrieve nodes limits', err);
    }
  }

  get() {
    return this.$async(this.getAsync);
  }
}

export default KubernetesNodesLimitsService;
angular.module('portainer.kubernetes').service('KubernetesNodesLimitsService', KubernetesNodesLimitsService);

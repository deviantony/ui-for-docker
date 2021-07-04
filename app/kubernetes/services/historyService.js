import angular from 'angular';
import PortainerError from 'Portainer/error';

import KubernetesHistoryHelper from 'Kubernetes/helpers/history';
import { KubernetesApplicationTypes } from 'Kubernetes/models/application/models';

class KubernetesHistoryService {
  /* @ngInject */
  constructor($async, KubernetesReplicaSetService, KubernetesControllerRevisionService) {
    this.$async = $async;
    this.KubernetesReplicaSetService = KubernetesReplicaSetService;
    this.KubernetesControllerRevisionService = KubernetesControllerRevisionService;

    this.getAsync = this.getAsync.bind(this);
  }

  /**
   * GET
   */
  async getAsync(application) {
    try {
      const namespace = application.ResourcePool;
      let rawRevisions;

      switch (application.ApplicationType) {
        case KubernetesApplicationTypes.DEPLOYMENT:
          rawRevisions = await this.KubernetesReplicaSetService.get(namespace);
          break;
        case KubernetesApplicationTypes.DAEMONSET:
          rawRevisions = await this.KubernetesControllerRevisionService.get(namespace);
          break;
        case KubernetesApplicationTypes.STATEFULSET:
          rawRevisions = await this.KubernetesControllerRevisionService.get(namespace);
          break;
        case KubernetesApplicationTypes.POD:
          rawRevisions = [];
          break;
        default:
          throw new PortainerError('Unable to determine which association to use for history');
      }
      if (rawRevisions.length) {
        const [currentRevision, revisionsList] = KubernetesHistoryHelper.getRevisions(rawRevisions, application);
        application.CurrentRevision = currentRevision;
        application.Revisions = revisionsList;
      }

      return application;
    } catch (err) {
      throw new PortainerError('', err);
    }
  }

  get(application) {
    return this.$async(this.getAsync, application);
  }
}

export default KubernetesHistoryService;
angular.module('portainer.kubernetes').service('KubernetesHistoryService', KubernetesHistoryService);

import angular from 'angular';
import _ from 'lodash-es';

import KubernetesVolumeConverter from 'Kubernetes/converters/volume';

class KubernetesVolumeService {
  /* @ngInject */
  constructor($async, KubernetesResourcePoolService, KubernetesApplicationService, KubernetesPersistentVolumeClaimService) {
    this.$async = $async;
    this.KubernetesResourcePoolService = KubernetesResourcePoolService;
    this.KubernetesApplicationService = KubernetesApplicationService;
    this.KubernetesPersistentVolumeClaimService = KubernetesPersistentVolumeClaimService;

    this.getAsync = this.getAsync.bind(this);
    this.getAllAsync = this.getAllAsync.bind(this);
    this.deleteAsync = this.deleteAsync.bind(this);
  }

  /**
   * GET
   */
  async getAsync(namespace, name) {
    const [pvc, pool] = await Promise.all([this.KubernetesPersistentVolumeClaimService.get(namespace, name), this.KubernetesResourcePoolService.get(namespace)]);
    return KubernetesVolumeConverter.pvcToVolume(pvc, pool);
  }

  async getAllAsync(namespace) {
    const data = await this.KubernetesResourcePoolService.get(namespace);
    const pools = data instanceof Array ? data : [data];
    const res = await Promise.all(
      _.map(pools, async (pool) => {
        const pvcs = await this.KubernetesPersistentVolumeClaimService.get(pool.Namespace.Name);
        return _.map(pvcs, (pvc) => KubernetesVolumeConverter.pvcToVolume(pvc, pool));
      })
    );
    return _.flatten(res);
  }

  get(namespace, name) {
    if (name) {
      return this.$async(this.getAsync, namespace, name);
    }
    return this.$async(this.getAllAsync, namespace);
  }

  /**
   * DELETE
   */
  async deleteAsync(volume) {
    await this.KubernetesPersistentVolumeClaimService.delete(volume.PersistentVolumeClaim);
  }

  delete(volume) {
    return this.$async(this.deleteAsync, volume);
  }
}

export default KubernetesVolumeService;
angular.module('portainer.kubernetes').service('KubernetesVolumeService', KubernetesVolumeService);

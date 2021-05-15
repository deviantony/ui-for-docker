import _ from 'lodash-es';

import angular from 'angular';
import KubernetesResourcePoolConverter from 'Kubernetes/converters/resourcePool';
import KubernetesResourceQuotaHelper from 'Kubernetes/helpers/resourceQuotaHelper';

/* @ngInject */
export function KubernetesResourcePoolService($async, KubernetesNamespaceService, KubernetesResourceQuotaService, KubernetesIngressService) {
  return {
    get,
    create,
    patch,
    delete: _delete,
  };

  async function getOne(name) {
    const namespace = await KubernetesNamespaceService.get(name);
    const [quotaAttempt] = await Promise.allSettled([KubernetesResourceQuotaService.get(name, KubernetesResourceQuotaHelper.generateResourceQuotaName(name))]);
    const pool = KubernetesResourcePoolConverter.apiToResourcePool(namespace);
    if (quotaAttempt.status === 'fulfilled') {
      pool.Quota = quotaAttempt.value;
      pool.Yaml += '---\n' + quotaAttempt.value.Yaml;
    }
    return pool;
  }

  async function getAll() {
    const namespaces = await KubernetesNamespaceService.get();
    const pools = await Promise.all(
      _.map(namespaces, async (namespace) => {
        const name = namespace.Name;
        const [quotaAttempt] = await Promise.allSettled([KubernetesResourceQuotaService.get(name, KubernetesResourceQuotaHelper.generateResourceQuotaName(name))]);
        const pool = KubernetesResourcePoolConverter.apiToResourcePool(namespace);
        if (quotaAttempt.status === 'fulfilled') {
          pool.Quota = quotaAttempt.value;
          pool.Yaml += '---\n' + quotaAttempt.value.Yaml;
        }
        return pool;
      })
    );
    return pools;
  }

  function get(name) {
    if (name) {
      return $async(getOne, name);
    }
    return $async(getAll);
  }

  function create(formValues) {
    return $async(async () => {
      const [namespace, quota, ingresses] = KubernetesResourcePoolConverter.formValuesToResourcePool(formValues);
      await KubernetesNamespaceService.create(namespace);

      if (quota) {
        await KubernetesResourceQuotaService.create(quota);
      }
      const ingressPromises = _.map(ingresses, (i) => KubernetesIngressService.create(i));
      await Promise.all(ingressPromises);
    });
  }

  function patch(oldFormValues, newFormValues) {
    return $async(async () => {
      const [oldNamespace, oldQuota, oldIngresses] = KubernetesResourcePoolConverter.formValuesToResourcePool(oldFormValues);
      const [newNamespace, newQuota, newIngresses] = KubernetesResourcePoolConverter.formValuesToResourcePool(newFormValues);
      void oldNamespace, newNamespace;

      if (oldQuota && newQuota) {
        await KubernetesResourceQuotaService.patch(oldQuota, newQuota);
      } else if (!oldQuota && newQuota) {
        await KubernetesResourceQuotaService.create(newQuota);
      } else if (oldQuota && !newQuota) {
        await KubernetesResourceQuotaService.delete(oldQuota);
      }

      const create = _.filter(newIngresses, (ing) => !_.find(oldIngresses, { Name: ing.Name }));
      const del = _.filter(oldIngresses, (ing) => !_.find(newIngresses, { Name: ing.Name }));
      const patch = _.without(newIngresses, ...create);

      const createPromises = _.map(create, (i) => KubernetesIngressService.create(i));
      const delPromises = _.map(del, (i) => KubernetesIngressService.delete(i));
      const patchPromises = _.map(patch, (ing) => {
        const old = _.find(oldIngresses, { Name: ing.Name });
        ing.Paths = angular.copy(old.Paths);
        ing.PreviousHost = old.Host;
        return KubernetesIngressService.patch(old, ing);
      });

      const promises = _.flatten([createPromises, delPromises, patchPromises]);
      await Promise.all(promises);
    });
  }

  function _delete(pool) {
    return $async(async () => {
      await KubernetesNamespaceService.delete(pool.Namespace);
    });
  }
}

angular.module('portainer.kubernetes').service('KubernetesResourcePoolService', KubernetesResourcePoolService);

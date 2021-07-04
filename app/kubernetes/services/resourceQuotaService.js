import _ from 'lodash-es';
import angular from 'angular';
import PortainerError from 'Portainer/error';
import { KubernetesCommonParams } from 'Kubernetes/models/common/params';
import KubernetesResourceQuotaConverter from 'Kubernetes/converters/resourceQuota';

/* @ngInject */
export function KubernetesResourceQuotaService($async, KubernetesResourceQuotas) {
  return {
    get,
    create,
    patch,
    delete: _delete,
  };

  async function getOne(namespace, name) {
    try {
      const params = new KubernetesCommonParams();
      params.id = name;
      const [raw, yaml] = await Promise.all([KubernetesResourceQuotas(namespace).get(params).$promise, KubernetesResourceQuotas(namespace).getYaml(params).$promise]);
      return KubernetesResourceQuotaConverter.apiToResourceQuota(raw, yaml);
    } catch (err) {
      throw new PortainerError('Unable to retrieve resource quota', err);
    }
  }

  async function getAll(namespace) {
    try {
      const data = await KubernetesResourceQuotas(namespace).get().$promise;
      return _.map(data.items, (item) => KubernetesResourceQuotaConverter.apiToResourceQuota(item));
    } catch (err) {
      throw new PortainerError('Unable to retrieve resource quotas', err);
    }
  }

  function get(namespace, name) {
    if (name) {
      return $async(getOne, namespace, name);
    }
    return $async(getAll, namespace);
  }

  function create(quota) {
    return $async(async () => {
      try {
        const payload = KubernetesResourceQuotaConverter.createPayload(quota);
        const namespace = payload.metadata.namespace;
        const params = {};
        const data = await KubernetesResourceQuotas(namespace).create(params, payload).$promise;
        return KubernetesResourceQuotaConverter.apiToResourceQuota(data);
      } catch (err) {
        throw new PortainerError('Unable to create quota', err);
      }
    });
  }

  function patch(oldQuota, newQuota) {
    return $async(async () => {
      try {
        const params = new KubernetesCommonParams();
        params.id = newQuota.Name;
        const namespace = newQuota.Namespace;
        const payload = KubernetesResourceQuotaConverter.patchPayload(oldQuota, newQuota);
        if (!payload.length) {
          return;
        }
        const data = await KubernetesResourceQuotas(namespace).patch(params, payload).$promise;
        return data;
      } catch (err) {
        throw new PortainerError('Unable to update resource quota', err);
      }
    });
  }

  function _delete(quota) {
    return $async(async () => {
      try {
        const params = new KubernetesCommonParams();
        params.id = quota.Name;
        await KubernetesResourceQuotas(quota.Namespace).delete(params).$promise;
      } catch (err) {
        throw new PortainerError('Unable to delete quota', err);
      }
    });
  }
}

angular.module('portainer.kubernetes').service('KubernetesResourceQuotaService', KubernetesResourceQuotaService);

import angular from 'angular';
import PortainerError from 'Portainer/error';
import { KubernetesCommonParams } from 'Kubernetes/models/common/params';

class KubernetesMetricsService {
  /* @ngInject */
  constructor($async, KubernetesMetrics) {
    this.$async = $async;
    this.KubernetesMetrics = KubernetesMetrics;

    this.capabilitiesAsync = this.capabilitiesAsync.bind(this);

    this.getPodAsync = this.getPodAsync.bind(this);
    this.getNodeAsync = this.getNodeAsync.bind(this);

    this.getPodsAsync = this.getPodsAsync.bind(this);
    this.getNodesAsync = this.getNodesAsync.bind(this);
  }

  /**
   * GET
   */
  async capabilitiesAsync(endpointID) {
    try {
      await this.KubernetesMetrics().capabilities({ endpointId: endpointID }).$promise;
    } catch (err) {
      throw new PortainerError('Unable to retrieve metrics', err);
    }
  }

  capabilities(endpointID) {
    return this.$async(this.capabilitiesAsync, endpointID);
  }

  /**
   * Stats of Node
   *
   * @param {string} nodeName
   */
  async getNodeAsync(nodeName) {
    try {
      const params = new KubernetesCommonParams();
      params.id = nodeName;
      const data = await this.KubernetesMetrics().getNode(params).$promise;
      return data;
    } catch (err) {
      throw new PortainerError('Unable to retrieve node stats', err);
    }
  }

  getNode(nodeName) {
    return this.$async(this.getNodeAsync, nodeName);
  }

  /**
   * Stats
   *
   * @param {string} namespace
   * @param {string} podName
   */
  async getPodAsync(namespace, podName) {
    try {
      const params = new KubernetesCommonParams();
      params.id = podName;
      const data = await this.KubernetesMetrics(namespace).getPod(params).$promise;
      return data;
    } catch (err) {
      throw new PortainerError('Unable to retrieve pod stats', err);
    }
  }

  getPod(namespace, podName) {
    return this.$async(this.getPodAsync, namespace, podName);
  }

  /**
   * Stats of Nodes in cluster
   *
   * @param {string} endpointID
   */
  async getNodesAsync(endpointID) {
    try {
      const data = await this.KubernetesMetrics().getNodes({ endpointId: endpointID }).$promise;
      return data;
    } catch (err) {
      throw new PortainerError('Unable to retrieve nodes stats', err);
    }
  }

  getNodes(endpointID) {
    return this.$async(this.getNodesAsync, endpointID);
  }

  /**
   * Stats of Pods in a namespace
   *
   * @param {string} namespace
   */
  async getPodsAsync(namespace) {
    try {
      const data = await this.KubernetesMetrics(namespace).getPods().$promise;
      return data;
    } catch (err) {
      throw new PortainerError('Unable to retrieve pod stats', err);
    }
  }

  getPods(namespace) {
    return this.$async(this.getPodsAsync, namespace);
  }
}

export default KubernetesMetricsService;
angular.module('portainer.kubernetes').service('KubernetesMetricsService', KubernetesMetricsService);

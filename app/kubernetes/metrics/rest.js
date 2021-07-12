angular.module('portainer.kubernetes').factory('KubernetesMetrics', [
  '$resource',
  'API_ENDPOINT_ENDPOINTS',
  'EndpointProvider',
  function KubernetesMetrics($resource, API_ENDPOINT_ENDPOINTS, EndpointProvider) {
    'use strict';
    return function (namespace) {
      const url = API_ENDPOINT_ENDPOINTS + '/:endpointId/kubernetes/apis/metrics.k8s.io/v1beta1';
      const podUrl = `${url}${namespace ? '/namespaces/:namespace' : ''}/pods/:id`;

      return $resource(
        url,
        {
          endpointId: EndpointProvider.endpointID,
          namespace: namespace,
        },
        {
          capabilities: { method: 'GET' },
          getPod: {
            method: 'GET',
            url: podUrl,
          },
          getNode: {
            method: 'GET',
            url: `${url}/nodes/:id`,
          },
          getPods: {
            method: 'GET',
            url: `${url}/namespaces/:namespace/pods`,
          },
          getNodes: {
            method: 'GET',
            url: `${url}/nodes`,
          },
        }
      );
    };
  },
]);

angular.module('portainer.kubernetes')
.factory('KubernetesSecrets', ['$resource', 'API_ENDPOINT_ENDPOINTS', 'EndpointProvider',
  function KubernetesSecretsFactory($resource, API_ENDPOINT_ENDPOINTS, EndpointProvider) {
    'use strict';
    return $resource(API_ENDPOINT_ENDPOINTS + '/:endpointId/kubernetes/api/v1/namespaces/:namespace/secrets/:id/:action',
      {
        endpointId: EndpointProvider.endpointID
      },
      {
        create: { method: 'POST', params: { namespace: '@metadata.namespace' } },
      });
  }]);
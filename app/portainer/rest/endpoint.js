import getEndpointsTotalCount from './transform/getEndpointsTotalCount';

angular.module('portainer.app').factory('Endpoints', [
  '$resource',
  'API_ENDPOINT_ENDPOINTS',
  function EndpointsFactory($resource, API_ENDPOINT_ENDPOINTS) {
    'use strict';
    return $resource(
      API_ENDPOINT_ENDPOINTS + '/:id/:action',
      {},
      {
        query: {
          method: 'GET',
          params: { start: '@start', limit: '@limit', search: '@search', groupId: '@groupId' },
          transformResponse: getEndpointsTotalCount,
        },
        get: { method: 'GET', params: { id: '@id' } },
        update: { method: 'PUT', params: { id: '@id' } },
        deassociate: { method: 'DELETE', params: { id: '@id', action: 'association' } },
        updateAccess: { method: 'PUT', params: { id: '@id', action: 'access' } },
        remove: { method: 'DELETE', params: { id: '@id' } },
        snapshots: { method: 'POST', params: { action: 'snapshot' } },
        snapshot: { method: 'POST', params: { id: '@id', action: 'snapshot' } },
        status: { method: 'GET', params: { id: '@id', action: 'status' } },
        updateSecuritySettings: { method: 'PUT', params: { id: '@id', action: 'settings' } },
        dockerhubLimits: {
          method: 'GET',
          url: `${API_ENDPOINT_ENDPOINTS}/:id/dockerhub/:registryId`,
        },
        registries: {
          method: 'GET',
          url: `${API_ENDPOINT_ENDPOINTS}/:id/registries`,
          params: { id: '@id', namespace: '@namespace' },
          isArray: true,
        },
        registry: {
          url: `${API_ENDPOINT_ENDPOINTS}/:id/registries/:registryId`,
          method: 'GET',
          params: { id: '@id', namespace: '@namespace', registryId: '@registryId' },
        },
        updateRegistryAccess: {
          method: 'PUT',
          url: `${API_ENDPOINT_ENDPOINTS}/:id/registries/:registryId`,
          params: { id: '@id', registryId: '@registryId' },
        },
      }
    );
  },
]);

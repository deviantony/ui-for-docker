angular.module('portainer.kubernetes').factory('KubernetesNodesLimits', [
  '$resource',
  'API_ENDPOINT_KUBERNETES',
  'EndpointProvider',
  function KubernetesNodesLimitsFactory($resource, API_ENDPOINT_KUBERNETES, EndpointProvider) {
    'use strict';
    return function () {
      const url = API_ENDPOINT_KUBERNETES + '/:endpointId/nodes_limits';
      return $resource(
        url,
        {
          endpointId: EndpointProvider.endpointID,
        },
        {
          get: {
            method: 'GET',
            ignoreLoadingBar: true,
            transformResponse: (data) => ({ data: JSON.parse(data) }),
          },
        }
      );
    };
  },
]);

angular.module('portainer.kubernetes').factory('KubernetesConfig', [
  '$http',
  'EndpointProvider',
  'API_ENDPOINT_KUBERNETES',
  function KubernetesConfigFactory($http, EndpointProvider, API_ENDPOINT_KUBERNETES) {
    'use strict';

    return { get };

    async function get() {
      const endpointID = EndpointProvider.endpointID();
      return $http({
        method: 'GET',
        url: `${API_ENDPOINT_KUBERNETES}/${endpointID}/config`,
        responseType: 'blob',
        headers: {
          Accept: 'text/yaml',
        },
      });
    }
  },
]);

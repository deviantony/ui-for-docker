import angular from 'angular';

angular.module('portainer.kubernetes').factory('KubernetesConfig', KubernetesConfigFactory);

/* @ngInject */
function KubernetesConfigFactory($http, EndpointProvider, API_ENDPOINT_KUBERNETES) {
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
}

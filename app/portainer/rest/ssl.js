import angular from 'angular';

const API_ENDPOINT_SSL = 'api/ssl';

angular.module('portainer.app').factory('SSL', SSLFactory);

/* @ngInject */
function SSLFactory($resource) {
  return $resource(
    API_ENDPOINT_SSL,
    {},
    {
      get: { method: 'GET' },
      upload: { method: 'PUT' },
    }
  );
}

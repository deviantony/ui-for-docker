angular.module('portainer.app').factory('LDAP', [
  '$resource',
  'API_ENDPOINT_LDAP',
  function SettingsFactory($resource, API_ENDPOINT_LDAP) {
    'use strict';
    return $resource(
      API_ENDPOINT_LDAP + '/:action',
      {},
      {
        adminGroups: { method: 'POST', isArray: true, params: { action: 'admin-groups' } },
      }
    );
  },
]);

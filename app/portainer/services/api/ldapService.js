angular.module('portainer.app').factory('LDAPService', [
  '$q',
  'LDAP',
  function LdapServiceFactory($q, LDAP) {
    'use strict';
    var service = {};

    service.adminGroups = function (ldapSettings) {
      var deferred = $q.defer();
      LDAP.adminGroups({ ldapSettings })
        .$promise.then(function success(data) {
          const userGroups = data.map((Name) => ({ Name, selected: ldapSettings.AdminGroups && ldapSettings.AdminGroups.includes(Name) ? true : false }));
          deferred.resolve(userGroups);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve admin grous', err: err });
        });

      return deferred.promise;
    };

    return service;
  },
]);

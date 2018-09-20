angular.module('portainer.app')
.factory('DeploykeyService', ['$q', 'Deploykeys', function DeploykeyServiceFactory($q, Deploykeys) {
  'use strict';
  var service = {};

  service.deploykeys = function() {
    var deferred = $q.defer();
    Deploykeys.query().$promise
    .then(function success(data) {
      var deploykeys = data.map(function (item) {
        return new DeploykeyViewModel(item);
      });            
      deferred.resolve(deploykeys);
    })
    .catch(function error(err) {
      deferred.reject({msg: 'Unable to retrieve keys', err: err});
    });
    return deferred.promise;
  };
  
  service.createNewkeyNames = function() {
    var deferred = $q.defer();
    Deploykeys.query().$promise
    .then(function success(data) {
      var deploykeys = data.map(function (item) {
        return item.Name;
      });      
      deferred.resolve(deploykeys);
    })
    .catch(function error(err) {
      deferred.reject({msg: 'Unable to retrieve keys', err: err});
    });
    return deferred.promise;
  };

  service.createNewdeploykey = function(name,UserName) {
    var payload = {
      Name: name,
      userName: UserName
    };
    return Deploykeys.create({}, payload).$promise;
  };

  service.deleteNewdeploykey = function(id) {
    return Deploykeys.remove({id: id}).$promise;
  };

  return service;
}]);

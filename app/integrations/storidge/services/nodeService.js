import { StoridgeNodeModel, StoridgeNodeDetailedModel } from '../models/node';

angular.module('portainer.integrations.storidge').factory('StoridgeNodeService', [
  '$q',
  'Storidge',
  function StoridgeNodeServiceFactory($q, Storidge) {
    'use strict';
    var service = {};

    service.nodes = function () {
      var deferred = $q.defer();

      Storidge.queryNodes()
        .$promise.then(function success(data) {
          var nodeData = data.nodes;
          var nodes = [];

          for (var key in nodeData) {
            if (Object.prototype.hasOwnProperty.call(nodeData, key)) {
              nodes.push(new StoridgeNodeModel(key, nodeData[key]));
            }
          }

          deferred.resolve(nodes);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve Storidge nodes', err: err });
        });

      return deferred.promise;
    };

    service.node = function (id) {
      var deferred = $q.defer();

      Storidge.getNode({ id: id })
        .$promise.then(function success(data) {
          var node = new StoridgeNodeDetailedModel(data.name, data.properties);
          deferred.resolve(node);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve Storidge node', err: err });
        });

      return deferred.promise;
    };

    service.add = function () {
      return Storidge.addNode().$promise;
    };

    service.cordon = function (id) {
      return Storidge.cordonNode({ id: id }).$promise;
    };

    service.uncordon = function (id) {
      return Storidge.uncordonNode({ id: id }).$promise;
    };

    service.remove = function (id) {
      return Storidge.removeNode({ id: id }).$promise;
    };

    return service;
  },
]);

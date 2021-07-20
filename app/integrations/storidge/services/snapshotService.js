import { StoridgeSnapshotModel } from '../models/snapshot';

angular.module('portainer.integrations.storidge').factory('StoridgeSnapshotService', [
  '$q',
  'Storidge',
  function StoridgeSnapshotServiceFactory($q, Storidge) {
    'use strict';
    var service = {};

    service.snapshots = snapshots;
    service.snapshot = snapshot;
    service.create = create;
    service.remove = remove;

    function snapshots(volumeId) {
      var deferred = $q.defer();

      Storidge.querySnapshots({ id: volumeId })
        .$promise.then(function success(data) {
          var snapshotsData = data.snapshots;
          let snapshotsArray = [];
          for (const key in snapshotsData) {
            if (Object.prototype.hasOwnProperty.call(snapshotsData, key)) {
              snapshotsArray.push(snapshotsData[key]);
            }
          }
          var snapshots = snapshotsArray.map(function (snapshot) {
            return new StoridgeSnapshotModel(snapshot);
          });
          deferred.resolve(snapshots);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve Storidge snapshots', err: err });
        });

      return deferred.promise;
    }

    function snapshot(id) {
      var deferred = $q.defer();

      Storidge.getSnapshot({ id: id })
        .$promise.then(function success(data) {
          var snapshot = new StoridgeSnapshotModel(data.snapshot);
          deferred.resolve(snapshot);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve Storidge snapshot', err: err });
        });

      return deferred.promise;
    }

    function create(volumeId, description) {
      var deferred = $q.defer();
      Storidge.createSnapshot({ id: volumeId, opts: { description: description } })
        .$promise.then(function success(data) {
          deferred.resolve(data);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to create Storidge volume snapshot', err: err });
        });

      return deferred.promise;
    }

    function remove(id) {
      var deferred = $q.defer();

      Storidge.removeSnapshot({ id: id })
        .$promise.then(function success() {
          deferred.resolve();
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to remove Storidge volume snapshot', err: err });
        });

      return deferred.promise;
    }

    return service;
  },
]);

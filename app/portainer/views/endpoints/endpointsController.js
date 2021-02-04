import angular from 'angular';

angular.module('portainer.app').controller('EndpointsController', EndpointsController);

function EndpointsController($q, $scope, $state, $async, EndpointService, GroupService, EndpointHelper, Notifications) {
  $scope.removeAction = removeAction;

  function removeAction(endpoints) {
    return $async(removeActionAsync, endpoints);
  }

  async function removeActionAsync(endpoints) {
    for (let endpoint of endpoints) {
      try {
        await EndpointService.deleteEndpoint(endpoint.Id);

        Notifications.success('Endpoint successfully removed', endpoint.Name);
      } catch (err) {
        Notifications.error('Failure', err, 'Unable to remove endpoint');
      }
    }

    $state.reload();
  }

  $scope.getPaginatedEndpoints = getPaginatedEndpoints;
  function getPaginatedEndpoints(lastId, limit, search) {
    const deferred = $q.defer();
    $q.all({
      endpoints: EndpointService.endpoints(lastId, limit, { search }),
      groups: GroupService.groups(),
    })
      .then(function success(data) {
        var endpoints = data.endpoints.value;
        var groups = data.groups;
        EndpointHelper.mapGroupToEndpoint(endpoints, groups);
        deferred.resolve({ endpoints: endpoints, totalCount: data.endpoints.totalCount });
      })
      .catch(function error(err) {
        Notifications.error('Failure', err, 'Unable to retrieve endpoint information');
      });
    return deferred.promise;
  }
}

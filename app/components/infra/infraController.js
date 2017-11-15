angular.module('infra', [])
.controller('InfraController', ['$interval', '$q', '$scope', '$state', 'EndpointProvider', 'OrcaEndpointService', 'ProjectService', 'EndpointService', 'InfraService', 'SystemService', 'NodeService', 'Pagination', 'Notifications', 'StateManager', 'Authentication',
function ($interval, $q, $scope, $state, EndpointProvider, OrcaEndpointService, ProjectService, EndpointService, InfraService, SystemService, NodeService, Pagination, Notifications, StateManager, Authentication) {
  $scope.state = {};
  $scope.state.pagination_count = Pagination.getPaginationCount('swarms');
  $scope.state.nonswarms_pagination_count = Pagination.getPaginationCount('nonswarms');
  $scope.state.selectedItemCount = 0;
  $scope.sortType = 'name';
  $scope.sortReverse = false;
  $scope.sortNonSwarmType = 'name';
  $scope.sortNonSwarmReverse = false;

  var statePromise;

  $scope.order = function(sortType) {
    $scope.sortReverse = ($scope.sortType === sortType) ? !$scope.sortReverse : false;
    $scope.sortType = sortType;
  };

  $scope.orderNonSwarm = function(sortType) {
    $scope.sortNonSwarmReverse = ($scope.sortNonSwarmType === sortType) ? !$scope.sortNonSwarmReverse : false;
    $scope.sortNonSwarmType = sortType;
  };

  $scope.changePaginationCount = function() {
    Pagination.setPaginationCount('swarms', $scope.state.pagination_count);
  };

  $scope.changeNonSwarmsPaginationCount = function() {
    Pagination.setPaginationCount('nonswarms', $scope.state.nonswarms_pagination_count);
  };

  $scope.selectItems = function (allSelected) {
    angular.forEach($scope.state.filteredSwarms, function (swarm) {
      if (swarm.Checked !== allSelected) {
        swarm.Checked = allSelected;
        $scope.selectItem(swarm);
      }
    });
  };

  $scope.selectItem = function (item) {
    if (item.Checked) {
      $scope.state.selectedItemCount++;
    } else {
      $scope.state.selectedItemCount--;
    }
  };

  $scope.discoverEndpoints = function() {
    // TODO: hardcode to ec2-us-east-1 for now -- otherwise get/iterate list of available providers
    var provider = 'ec2-us-east-1'

    $scope.spinner = true;

    OrcaEndpointService.discover(provider)
    .then(function success(data) {
      // console.log("Running task to create new project...")
    })
    .then(function success() {
      Notifications.success('Endpoint discovery successfully completed');
    })
    .catch(function error(err) {
      Notifications.error('Failure', err, 'An error occured during endpoint discovery for provider ' + provider);
    })
    .finally(function final() {
      $scope.spinner = false;
    });

    $scope.spinner = false;
  };

  $scope.reloadEndpoints = function() {
    // TODO: hardcode to ec2-us-east-1 for now -- otherwise get/iterate list of available providers
    var provider = 'ec2-us-east-1'

    $scope.spinner = true;

    var foundNewEndpoints = [];
    OrcaEndpointService.endpoints(provider)
    .then(function success(data) {
      var endpoints = data;
      if (endpoints != null) {
           EndpointService.endpoints()
           .then(function success(data) {
                var existingEndpoints = data;
                // Analyze found EPs for NEW EPs only
                for (var i = 0; i < endpoints.length; i++) {
                    var ep = endpoints[i];
                    foundExisting = false;
                    for (var j = 0; j < existingEndpoints.length; j++) {
                        var oldEp = existingEndpoints[j];
                        // TODO: update existing?
                        if ("tcp://" + ep.DockerURL == oldEp.URL) {
                            foundExisting = true;
                            break;
                        }
                    }

                    if (foundExisting == false) {
                        // TODO: handle errors
                        EndpointService.createRemoteEndpoint(ep.Name, ep.DockerURL, "tcp://" + ep.DockerURL, false, true, true, null, null, null)
                        .then(function success(data) {
                            foundNewEndpoints.push(ep);
                            $scope.newEndpoints = foundNewEndpoints;
                            Notifications.success('Endpoint created', ep.Name);
                            //$state.reload();
                        });
                    }
                }
          });
      }
      //$scope.newEndpoints = foundNewEndpoints;
    })
    .catch(function error(err) {
      $scope.newEndpoints = [];
      Notifications.error('Failure', err, 'Unable to reload endpoints for provider ' + provider);
    })
    .finally(function final() {
      $scope.spinner = false;
    });
  };

  $scope.statusAction = function () {
    // TODO: hardcode to ec2-us-east-1 for now -- otherwise get/iterate list of available providers
    var provider = 'ec2-us-east-1'

    $scope.spinner = true;

    ProjectService.operationStatus(provider)
        .then(function success(data) {
          if (data.Name == provider + ": ") {
            $scope.operationStatus = "";
            $scope.spinner = false;
          } else {
            $scope.spinner = true;
            $scope.operationStatus = data;
          }

          if ($scope.operationStatus != "") {
              ProjectService.messageStatus(provider)
                .then(function success(data) {
                  var messages = [];
                  var errors = [];
                  for (var i = 0; i < data.length; i++) {
                      var entry = data[i];
                      if (entry.Name != '' && entry.Messages != [] && entry.Messages != '') {
                        messages.push({Name: entry.Name, Messages: entry.Messages});
                      }
                      if (entry.Name != '' && entry.Errors != [] && entry.Errors != '') {
                        errors.push({Name: entry.Name, Errors: entry.Errors});
                      }
                  }
                  if (messages != [] && messages != '' && messages != null) {
                    $scope.messages = messages;
                  } else {
                    delete $scope.messages;
                  }
                  if (errors != [] && errors != '' && errors != null) {
                    $scope.errors = errors;
                  } else {
                    delete $scope.errors;
                  }
                })
                .catch(function error(err) {
                  Notifications.error('Failure', err, 'Unable to get Orca message status');
                });
            }
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to get Orca operation status');
        })
        .finally(function final() {
          $scope.spinner = false;
        });
  };

  function initView() {
    $scope.spinner = true;
    $scope.applicationState.infra = true;
    $scope.operationStatus = "";

    EndpointProvider.setSwitchedEndpointID("");

    // TODO: add re-discover or refresh option later
    var tmpSwarms = InfraService.getSwarms();
    var tmpNonSwarms = InfraService.getNonSwarms();
    if (tmpSwarms.length == 0 && InfraService.getDataLoading() == false) {
        InfraService.setDataLoading(true);
        EndpointService.endpoints()
        .then(function success(data) {
          $scope.endpoints = data;
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to retrieve endpoints');
          $scope.endpoints = [];
        })
        .finally(function final() {
          InfraService.getEndpointStates($scope.endpoints)
          .then(function success(data) {
            var foundSwarms = [];
            var foundNonSwarms = [];
            for (var i = 0; i < data.length; i++) {
                var epEntry = data[i];

                // TODO: improve this mapping, likely on initial EP setup...
                for (var j = 0; j < $scope.endpoints.length; j++) {
                    var oldEpEntry = $scope.endpoints[j];
                    if (oldEpEntry.Id == epEntry.id) {
                        epEntry.name = oldEpEntry.Name;
                        break;
                    }
                }

                if (epEntry.provider == "DOCKER_SWARM_MODE") {
                    foundSwarms.push(epEntry);
                } else {
                    foundNonSwarms.push(epEntry);
                }
            }
            $scope.swarms = foundSwarms;
            $scope.nonswarms = foundNonSwarms;
            InfraService.setSwarms(foundSwarms);
            InfraService.setNonSwarms(foundNonSwarms);
            InfraService.setDataLoading(false);
            $state.reload();
          });
        });
    } else {
        $scope.swarms = tmpSwarms;
        $scope.nonswarms = tmpNonSwarms;
    }

    // Reset cached EP list
    InfraService.setCachedSwarmEndpoints([]);

    $scope.spinner = false;
    statusPromise = $interval($scope.statusAction, 5000);
  }

  $scope.$on('$destroy', function() {
      $interval.cancel(statusPromise);
  });

  initView();
}]);

angular
  .module('portainer.app')
  .controller('HomeController', function (
    $q,
    $scope,
    $state,
    TagService,
    Authentication,
    EndpointService,
    EndpointHelper,
    GroupService,
    Notifications,
    EndpointProvider,
    StateManager,
    ModalService,
    MotdService
  ) {
    $scope.state = {
      connectingToEdgeEndpoint: false,
    };

    $scope.goToEdit = function (id) {
      $state.go('portainer.endpoints.endpoint', { id: id });
    };

    $scope.goToDashboard = function (endpoint) {
      if (endpoint.Type === 3) {
        $state.go('azure.dashboard', { endpointId: endpoint.Id });
        return;
      }
      if (endpoint.Type === 4 || endpoint.Type === 7) {
        if (!endpoint.EdgeID) {
          $state.go('portainer.endpoints.endpoint', { id: endpoint.Id });
          return;
        }
        $scope.state.connectingToEdgeEndpoint = true;
      }
      if (endpoint.Type === 5 || endpoint.Type === 6 || endpoint.Type === 7) {
        $state.go('kubernetes.dashboard', { endpointId: endpoint.Id });
        return;
      }
      $state.go('docker.dashboard', { endpointId: endpoint.Id });
    };

    $scope.dismissImportantInformation = function (hash) {
      StateManager.dismissImportantInformation(hash);
    };

    $scope.dismissInformationPanel = function (id) {
      StateManager.dismissInformationPanel(id);
    };

    $scope.triggerSnapshot = function () {
      ModalService.confirmEndpointSnapshot(function (result) {
        if (!result) {
          return;
        }
        triggerSnapshot();
      });
    };

    function triggerSnapshot() {
      EndpointService.snapshotEndpoints()
        .then(function success() {
          Notifications.success('Success', 'Endpoints updated');
          $state.reload();
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'An error occured during endpoint snapshot');
        });
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
          EndpointProvider.setEndpoints(endpoints);
          deferred.resolve({ endpoints: endpoints, totalCount: data.endpoints.totalCount });
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to retrieve endpoint information');
        });
      return deferred.promise;
    }

    async function initView() {
      $scope.isAdmin = Authentication.isAdmin();

      MotdService.motd().then(function success(data) {
        $scope.motd = data;
      });

      try {
        const [{ totalCount, endpoints }, tags] = await Promise.all([getPaginatedEndpoints(0, 100), TagService.tags()]);
        $scope.tags = tags;

        $scope.totalCount = totalCount;
        if (totalCount > 100) {
          $scope.endpoints = [];
        } else {
          $scope.endpoints = endpoints;
        }
      } catch (err) {
        Notifications.error('Failed loading page data', err);
      }
    }

    initView();
  });

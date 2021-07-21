import moment from 'moment';

angular.module('portainer.app').factory('StateManager', [
  '$q',
  '$async',
  'SystemService',
  'InfoHelper',
  'LocalStorage',
  'SettingsService',
  'StatusService',
  'APPLICATION_CACHE_VALIDITY',
  'AgentPingService',
  '$analytics',
  function StateManagerFactory($q, $async, SystemService, InfoHelper, LocalStorage, SettingsService, StatusService, APPLICATION_CACHE_VALIDITY, AgentPingService, $analytics) {
    var manager = {};

    var state = {
      loading: true,
      application: {},
      endpoint: {},
      UI: {
        dismissedInfoPanels: {},
        dismissedInfoHash: '',
      },
      extensions: [],
    };

    manager.setVersionInfo = function (versionInfo) {
      state.application.versionStatus = versionInfo;
      LocalStorage.storeApplicationState(state.application);
    };

    manager.dismissInformationPanel = function (id) {
      state.UI.dismissedInfoPanels[id] = true;
      LocalStorage.storeUIState(state.UI);
    };

    manager.dismissImportantInformation = function (hash) {
      state.UI.dismissedInfoHash = hash;
      LocalStorage.storeUIState(state.UI);
    };

    manager.getState = function () {
      return state;
    };

    manager.clean = function () {
      state.endpoint = {};
      state.application = {};
    };

    manager.updateLogo = function (logoURL) {
      state.application.logo = logoURL;
      LocalStorage.storeApplicationState(state.application);
    };

    manager.updateSnapshotInterval = function (interval) {
      state.application.snapshotInterval = interval;
      LocalStorage.storeApplicationState(state.application);
    };

    manager.updateEnableEdgeComputeFeatures = function updateEnableEdgeComputeFeatures(enableEdgeComputeFeatures) {
      state.application.enableEdgeComputeFeatures = enableEdgeComputeFeatures;
      LocalStorage.storeApplicationState(state.application);
    };

    manager.updateEnableTelemetry = function updateEnableTelemetry(enableTelemetry) {
      state.application.enableTelemetry = enableTelemetry;
      $analytics.setOptOut(!enableTelemetry);
      LocalStorage.storeApplicationState(state.application);
    };

    function assignStateFromStatusAndSettings(status, settings) {
      state.application.version = status.Version;
      state.application.edition = status.Edition;
      state.application.instanceId = status.InstanceID;

      state.application.enableTelemetry = settings.EnableTelemetry;
      state.application.logo = settings.LogoURL;
      state.application.snapshotInterval = settings.SnapshotInterval;
      state.application.enableEdgeComputeFeatures = settings.EnableEdgeComputeFeatures;
      state.application.validity = moment().unix();
    }

    function loadApplicationState() {
      var deferred = $q.defer();

      $q.all({
        settings: SettingsService.publicSettings(),
        status: StatusService.status(),
      })
        .then(function success(data) {
          var status = data.status;
          var settings = data.settings;
          assignStateFromStatusAndSettings(status, settings);
          LocalStorage.storeApplicationState(state.application);
          deferred.resolve(state);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve server settings and status', err: err });
        });

      return deferred.promise;
    }

    manager.initialize = initialize;
    async function initialize() {
      return $async(async () => {
        const UIState = LocalStorage.getUIState();
        if (UIState) {
          state.UI = UIState;
        }

        const endpointState = LocalStorage.getEndpointState();
        if (endpointState) {
          state.endpoint = endpointState;
        }

        const applicationState = LocalStorage.getApplicationState();
        if (isAppStateValid(applicationState)) {
          state.application = applicationState;
        } else {
          await loadApplicationState();
        }

        state.loading = false;
        $analytics.setPortainerStatus(state.application.instanceId, state.application.version);
        $analytics.setOptOut(!state.application.enableTelemetry);
        return state;
      });
    }

    function isAppStateValid(appState) {
      if (!appState || !appState.validity) {
        return false;
      }
      const now = moment().unix();
      const cacheValidity = now - appState.validity;
      return cacheValidity < APPLICATION_CACHE_VALIDITY;
    }

    function assignExtensions(endpointExtensions) {
      var extensions = [];

      for (var i = 0; i < endpointExtensions.length; i++) {
        var extension = endpointExtensions[i];
        if (extension.Type === 1) {
          extensions.push('storidge');
        }
      }

      return extensions;
    }

    manager.updateEndpointState = function (endpoint, extensions) {
      var deferred = $q.defer();

      if (endpoint.Type === 3) {
        state.endpoint.name = endpoint.Name;
        state.endpoint.mode = { provider: 'AZURE' };
        LocalStorage.storeEndpointState(state.endpoint);
        deferred.resolve();
        return deferred.promise;
      } else if (endpoint.Type === 5 || endpoint.Type === 6 || endpoint.Type === 7) {
        state.endpoint.name = endpoint.Name;
        state.endpoint.mode = { provider: 'KUBERNETES' };
        LocalStorage.storeEndpointState(state.endpoint);
        deferred.resolve();
        return deferred.promise;
      }

      const reload = endpoint.Status === 1 || !endpoint.Snaphosts || !endpoint.Snaphosts.length || !endpoint.Snapshots[0].SnapshotRaw;

      $q.all({
        version: reload ? SystemService.version() : $q.when(endpoint.Snapshots[0].SnapshotRaw.Version),
        info: reload ? SystemService.info() : $q.when(endpoint.Snapshots[0].SnapshotRaw.Info),
      })
        .then(function success(data) {
          var endpointMode = InfoHelper.determineEndpointMode(data.info, endpoint.Type);
          var endpointAPIVersion = parseFloat(data.version.ApiVersion);
          state.endpoint.mode = endpointMode;
          state.endpoint.name = endpoint.Name;
          state.endpoint.type = endpoint.Type;
          state.endpoint.apiVersion = endpointAPIVersion;
          state.endpoint.extensions = assignExtensions(extensions);

          if (endpointMode.agentProxy && endpoint.Status === 1) {
            return AgentPingService.ping().then(function onPingSuccess(data) {
              state.endpoint.agentApiVersion = data.version;
            });
          }
        })
        .then(function () {
          LocalStorage.storeEndpointState(state.endpoint);
          deferred.resolve();
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to connect to the Docker endpoint', err: err });
        })
        .finally(function final() {
          state.loading = false;
        });

      return deferred.promise;
    };

    manager.getAgentApiVersion = function getAgentApiVersion() {
      return state.endpoint.agentApiVersion;
    };

    return manager;
  },
]);

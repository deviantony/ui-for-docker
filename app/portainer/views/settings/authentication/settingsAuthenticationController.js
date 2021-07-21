import _ from 'lodash-es';
angular.module('portainer.app').controller('SettingsAuthenticationController', [
  '$q',
  '$scope',
  '$state',
  'Notifications',
  'SettingsService',
  'FileUploadService',
  'TeamService',
  'LDAPService',
  function ($q, $scope, $state, Notifications, SettingsService, FileUploadService, TeamService, LDAPService) {
    const DEFAULT_GROUP_FILTER = '(objectClass=groupOfNames)';

    $scope.state = {
      successfulConnectivityCheck: false,
      failedConnectivityCheck: false,
      uploadInProgress: false,
      connectivityCheckInProgress: false,
      actionInProgress: false,
      availableUserSessionTimeoutOptions: [
        {
          key: '1 hour',
          value: '1h',
        },
        {
          key: '4 hours',
          value: '4h',
        },
        {
          key: '8 hours',
          value: '8h',
        },
        {
          key: '24 hours',
          value: '24h',
        },
        { key: '1 week', value: `${24 * 7}h` },
        { key: '1 month', value: `${24 * 30}h` },
        { key: '6 months', value: `${24 * 30 * 6}h` },
        { key: '1 year', value: `${24 * 30 * 12}h` },
      ],
      enableAssignAdminGroup: false,
    };

    $scope.formValues = {
      UserSessionTimeout: $scope.state.availableUserSessionTimeoutOptions[0],
      TLSCACert: '',
      LDAPSettings: {
        AnonymousMode: true,
        ReaderDN: '',
        URL: '',
        TLSConfig: {
          TLS: false,
          TLSSkipVerify: false,
        },
        StartTLS: false,
        SearchSettings: [
          {
            BaseDN: '',
            Filter: '',
            UserNameAttribute: '',
          },
        ],
        GroupSearchSettings: [
          {
            GroupBaseDN: '',
            GroupFilter: '',
            GroupAttribute: '',
          },
        ],
        AdminGroupSearchSettings: [
          {
            GroupBaseDN: '',
            GroupFilter: '',
            GroupAttribute: '',
          },
        ],
        AdminAutoPopulate: false,
        AutoCreateUsers: true,
      },
      selectedAdminGroups: '',
    };

    $scope.groups = null;

    $scope.searchAdminGroups = async function () {
      const settings = {
        ...$scope.settings.LDAPSettings,
        AdminGroupSearchSettings: $scope.settings.LDAPSettings.AdminGroupSearchSettings.map((search) => ({ ...search, GroupFilter: search.GroupFilter || DEFAULT_GROUP_FILTER })),
      };

      $scope.groups = await LDAPService.adminGroups(settings);
      $scope.state.enableAssignAdminGroup = $scope.groups && $scope.groups.length > 0;
    };

    $scope.isOauthEnabled = function isOauthEnabled() {
      return $scope.settings && $scope.settings.AuthenticationMethod === 3;
    };

    $scope.addSearchConfiguration = function () {
      $scope.formValues.LDAPSettings.SearchSettings.push({ BaseDN: '', UserNameAttribute: '', Filter: '' });
    };

    $scope.removeSearchConfiguration = function (index) {
      $scope.formValues.LDAPSettings.SearchSettings.splice(index, 1);
    };

    $scope.addGroupSearchConfiguration = function () {
      $scope.formValues.LDAPSettings.GroupSearchSettings.push({ GroupBaseDN: '', GroupAttribute: '', GroupFilter: '' });
    };

    $scope.removeGroupSearchConfiguration = function (index) {
      $scope.formValues.LDAPSettings.GroupSearchSettings.splice(index, 1);
    };

    $scope.addAdminGroupSearchConfiguration = function () {
      $scope.formValues.LDAPSettings.GroupSearchSettings.push({ GroupBaseDN: '', GroupAttribute: '', GroupFilter: '' });
    };

    $scope.removeAdminGroupSearchConfiguration = function (index) {
      $scope.formValues.LDAPSettings.GroupSearchSettings.splice(index, 1);
    };

    $scope.LDAPConnectivityCheck = function () {
      var settings = angular.copy($scope.settings);
      var TLSCAFile = $scope.formValues.TLSCACert !== settings.LDAPSettings.TLSConfig.TLSCACert ? $scope.formValues.TLSCACert : null;

      if ($scope.formValues.LDAPSettings.AnonymousMode) {
        settings.LDAPSettings['ReaderDN'] = '';
        settings.LDAPSettings['Password'] = '';
      }

      var uploadRequired = ($scope.formValues.LDAPSettings.TLSConfig.TLS || $scope.formValues.LDAPSettings.StartTLS) && !$scope.formValues.LDAPSettings.TLSConfig.TLSSkipVerify;
      $scope.state.uploadInProgress = uploadRequired;

      $scope.state.connectivityCheckInProgress = true;
      $q.when(!uploadRequired || FileUploadService.uploadLDAPTLSFiles(TLSCAFile, null, null))
        .then(function success() {
          addLDAPDefaultPort(settings, $scope.formValues.LDAPSettings.TLSConfig.TLS);
          return SettingsService.checkLDAPConnectivity(settings);
        })
        .then(function success() {
          $scope.state.failedConnectivityCheck = false;
          $scope.state.successfulConnectivityCheck = true;
          Notifications.success('Connection to LDAP successful');
        })
        .catch(function error(err) {
          $scope.state.failedConnectivityCheck = true;
          $scope.state.successfulConnectivityCheck = false;
          Notifications.error('Failure', err, 'Connection to LDAP failed');
        })
        .finally(function final() {
          $scope.state.uploadInProgress = false;
          $scope.state.connectivityCheckInProgress = false;
        });
    };

    $scope.saveSettings = function () {
      var settings = angular.copy($scope.settings);
      if ($scope.formValues.LDAPSettings.AdminAutoPopulate && $scope.formValues.selectedAdminGroups && $scope.formValues.selectedAdminGroups.length > 0) {
        settings.LDAPSettings.AdminGroups = _.map($scope.formValues.selectedAdminGroups, (team) => team.Name);
      } else {
        settings.LDAPSettings.AdminGroups = [];
      }

      if ($scope.formValues.selectedAdminGroups && $scope.formValues.selectedAdminGroups.length === 0) {
        settings.LDAPSettings.AdminAutoPopulate = false;
      }

      var TLSCAFile = $scope.formValues.TLSCACert !== settings.LDAPSettings.TLSConfig.TLSCACert ? $scope.formValues.TLSCACert : null;

      if ($scope.formValues.LDAPSettings.AnonymousMode) {
        settings.LDAPSettings['ReaderDN'] = '';
        settings.LDAPSettings['Password'] = '';
      }

      var uploadRequired = ($scope.formValues.LDAPSettings.TLSConfig.TLS || $scope.formValues.LDAPSettings.StartTLS) && !$scope.formValues.LDAPSettings.TLSConfig.TLSSkipVerify;
      $scope.state.uploadInProgress = uploadRequired;

      $scope.state.actionInProgress = true;
      $q.when(!uploadRequired || FileUploadService.uploadLDAPTLSFiles(TLSCAFile, null, null))
        .then(function success() {
          addLDAPDefaultPort(settings, $scope.formValues.LDAPSettings.TLSConfig.TLS);
          return SettingsService.update(settings);
        })
        .then(function success() {
          Notifications.success('Authentication settings updated');
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to update authentication settings');
        })
        .finally(function final() {
          $scope.state.uploadInProgress = false;
          $scope.state.actionInProgress = false;
        });
    };

    // Add default port if :port is not defined in URL
    function addLDAPDefaultPort(settings, tlsEnabled) {
      if (settings.LDAPSettings.URL.indexOf(':') === -1) {
        settings.LDAPSettings.URL += tlsEnabled ? ':636' : ':389';
      }
    }

    async function prelodaAdminGroup() {
      if ($scope.settings.LDAPSettings.AdminAutoPopulate && $scope.settings.LDAPSettings.AdminGroups && $scope.settings.LDAPSettings.AdminGroups.length > 0) {
        const settings = {
          ...$scope.settings.LDAPSettings,
          AdminGroupSearchSettings: $scope.settings.LDAPSettings.AdminGroupSearchSettings.map((search) => ({
            ...search,
            GroupFilter: search.GroupFilter || '(objectClass=groupOfNames)',
          })),
        };

        $scope.groups = await LDAPService.adminGroups(settings);
        $scope.state.enableAssignAdminGroup = $scope.groups && $scope.groups.length > 0;
      }
    }
    function initView() {
      $q.all({
        settings: SettingsService.settings(),
        teams: TeamService.teams(),
      })
        .then(function success(data) {
          var settings = data.settings;
          $scope.teams = data.teams;
          $scope.settings = settings;
          $scope.formValues.LDAPSettings = settings.LDAPSettings;
          $scope.OAuthSettings = settings.OAuthSettings;
          $scope.formValues.TLSCACert = settings.LDAPSettings.TLSConfig.TLSCACert;
          prelodaAdminGroup();
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to retrieve application settings');
        });
    }

    initView();
  },
]);

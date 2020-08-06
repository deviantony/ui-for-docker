import _ from 'lodash-es';

async function initAuthentication(authManager, Authentication, $rootScope, $state) {
  authManager.checkAuthOnRefresh();
  // The unauthenticated event is broadcasted by the jwtInterceptor when
  // hitting a 401. We're using this instead of the usual combination of
  // authManager.redirectWhenUnauthenticated() + unauthenticatedRedirector
  // to have more controls on which URL should trigger the unauthenticated state.
  $rootScope.$on('unauthenticated', function (event, data) {
    if (!_.includes(data.config.url, '/v2/') && !_.includes(data.config.url, '/api/v4/')) {
      $state.go('portainer.logout', { error: 'Your session has expired' });
    }
  });

  await Authentication.init();
}

// function initAnalytics(Analytics, $rootScope) {
//   Analytics.offline(false);
//   Analytics.registerScriptTags();
//   Analytics.registerTrackers();
//   $rootScope.$on('$stateChangeSuccess', function (event, toState) {
//     Analytics.trackPage(toState.url);
//     Analytics.pageView();
//   });
// }

angular.module('portainer.app', ['portainer.oauth']).config([
  '$stateRegistryProvider',
  function ($stateRegistryProvider) {
    'use strict';

    var root = {
      name: 'root',
      abstract: true,
      resolve: {
        initStateManager: [
          'StateManager',
          'Authentication',
          'Notifications',
          'Analytics',
          'authManager',
          '$rootScope',
          '$state',
          '$async',
          '$q',
          (StateManager, Authentication, Notifications, Analytics, authManager, $rootScope, $state, $async, $q) => {
            const deferred = $q.defer();
            const appState = StateManager.getState();
            if (!appState.loading) {
              deferred.resolve();
            } else {
              StateManager.initialize()
                .then(function success() {
                  // if (state.application.analytics) {
                  //   initAnalytics(Analytics, $rootScope);
                  // }
                  return $async(initAuthentication, authManager, Authentication, $rootScope, $state);
                })
                .then(() => deferred.resolve())
                .catch(function error(err) {
                  Notifications.error('Failure', err, 'Unable to retrieve application settings');
                  deferred.reject(err);
                });
            }
            return deferred.promise;
          },
        ],
      },
      views: {
        'sidebar@': {
          templateUrl: './views/sidebar/sidebar.html',
          controller: 'SidebarController',
        },
      },
    };

    var endpointRoot = {
      name: 'endpoint',
      url: '/:endpointId',
      parent: 'root',
      abstract: true,
      resolve: {
        endpoint: /* @ngInject */ function endpoint($async, $state, $transition$, EndpointService, Notifications) {
          return $async(async () => {
            try {
              const endpointId = +$transition$.params().endpointId;

              const endpoint = await EndpointService.endpoint(endpointId);
              if ((endpoint.Type === 4 || endpoint.Type === 7) && !endpoint.EdgeID) {
                $state.go('portainer.endpoints.endpoint', { id: endpoint.Id });
                return;
              }

              return endpoint;
            } catch (e) {
              Notifications.error('Failed loading endpoint', e);
              $state.go('portainer.home', {}, { reload: true });
              return;
            }
          });
        },
      },
    };

    var portainer = {
      name: 'portainer',
      parent: 'root',
      abstract: true,
    };

    var about = {
      name: 'portainer.about',
      url: '/about',
      views: {
        'content@': {
          templateUrl: './views/about/about.html',
        },
      },
    };

    var account = {
      name: 'portainer.account',
      url: '/account',
      views: {
        'content@': {
          templateUrl: './views/account/account.html',
          controller: 'AccountController',
        },
      },
    };

    var authentication = {
      name: 'portainer.auth',
      url: '/auth',
      params: {
        reload: false,
      },
      views: {
        'content@': {
          templateUrl: './views/auth/auth.html',
          controller: 'AuthenticationController',
          controllerAs: 'ctrl',
        },
        'sidebar@': {},
      },
    };
    const logout = {
      name: 'portainer.logout',
      url: '/logout',
      params: {
        error: '',
        performApiLogout: false,
      },
      views: {
        'content@': {
          templateUrl: './views/logout/logout.html',
          controller: 'LogoutController',
          controllerAs: 'ctrl',
        },
        'sidebar@': {},
      },
    };

    var endpoints = {
      name: 'portainer.endpoints',
      url: '/endpoints',
      views: {
        'content@': {
          templateUrl: './views/endpoints/endpoints.html',
          controller: 'EndpointsController',
        },
      },
    };

    var endpoint = {
      name: 'portainer.endpoints.endpoint',
      url: '/:id',
      views: {
        'content@': {
          templateUrl: './views/endpoints/edit/endpoint.html',
          controller: 'EndpointController',
        },
      },
    };

    const endpointKubernetesConfiguration = {
      name: 'portainer.endpoints.endpoint.kubernetesConfig',
      url: '/configure',
      views: {
        'content@': {
          templateUrl: '../kubernetes/views/configure/configure.html',
          controller: 'KubernetesConfigureController',
          controllerAs: 'ctrl',
        },
      },
    };

    var endpointCreation = {
      name: 'portainer.endpoints.new',
      url: '/new',
      views: {
        'content@': {
          templateUrl: './views/endpoints/create/createendpoint.html',
          controller: 'CreateEndpointController',
        },
      },
    };

    var endpointAccess = {
      name: 'portainer.endpoints.endpoint.access',
      url: '/access',
      views: {
        'content@': {
          templateUrl: './views/endpoints/access/endpointAccess.html',
          controller: 'EndpointAccessController',
          controllerAs: 'ctrl',
        },
      },
    };

    var groups = {
      name: 'portainer.groups',
      url: '/groups',
      views: {
        'content@': {
          templateUrl: './views/groups/groups.html',
          controller: 'GroupsController',
        },
      },
    };

    var group = {
      name: 'portainer.groups.group',
      url: '/:id',
      views: {
        'content@': {
          templateUrl: './views/groups/edit/group.html',
          controller: 'GroupController',
        },
      },
    };

    var groupCreation = {
      name: 'portainer.groups.new',
      url: '/new',
      views: {
        'content@': {
          templateUrl: './views/groups/create/creategroup.html',
          controller: 'CreateGroupController',
        },
      },
    };

    var groupAccess = {
      name: 'portainer.groups.group.access',
      url: '/access',
      views: {
        'content@': {
          templateUrl: './views/groups/access/groupAccess.html',
          controller: 'GroupAccessController',
        },
      },
    };

    var home = {
      name: 'portainer.home',
      url: '/home',
      views: {
        'content@': {
          templateUrl: './views/home/home.html',
          controller: 'HomeController',
        },
      },
    };

    var init = {
      name: 'portainer.init',
      abstract: true,
      url: '/init',
      views: {
        'sidebar@': {},
      },
    };

    var initEndpoint = {
      name: 'portainer.init.endpoint',
      url: '/endpoint',
      views: {
        'content@': {
          templateUrl: './views/init/endpoint/initEndpoint.html',
          controller: 'InitEndpointController',
          controllerAs: 'ctrl',
        },
      },
    };

    var initAdmin = {
      name: 'portainer.init.admin',
      url: '/admin',
      views: {
        'content@': {
          templateUrl: './views/init/admin/initAdmin.html',
          controller: 'InitAdminController',
        },
      },
    };

    var extensions = {
      name: 'portainer.extensions',
      url: '/extensions',
      views: {
        'content@': {
          templateUrl: './views/extensions/extensions.html',
          controller: 'ExtensionsController',
        },
      },
    };

    var extension = {
      name: 'portainer.extensions.extension',
      url: '/extension/:id',
      views: {
        'content@': {
          templateUrl: './views/extensions/inspect/extension.html',
          controller: 'ExtensionController',
        },
      },
    };

    var registries = {
      name: 'portainer.registries',
      url: '/registries',
      views: {
        'content@': {
          templateUrl: './views/registries/registries.html',
          controller: 'RegistriesController',
        },
      },
    };

    var registry = {
      name: 'portainer.registries.registry',
      url: '/:id',
      views: {
        'content@': {
          templateUrl: './views/registries/edit/registry.html',
          controller: 'RegistryController',
        },
      },
    };

    var registryCreation = {
      name: 'portainer.registries.new',
      url: '/new',
      views: {
        'content@': {
          templateUrl: './views/registries/create/createregistry.html',
          controller: 'CreateRegistryController',
        },
      },
    };

    var registryAccess = {
      name: 'portainer.registries.registry.access',
      url: '/access',
      views: {
        'content@': {
          templateUrl: './views/registries/access/registryAccess.html',
          controller: 'RegistryAccessController',
        },
      },
    };

    var settings = {
      name: 'portainer.settings',
      url: '/settings',
      views: {
        'content@': {
          templateUrl: './views/settings/settings.html',
          controller: 'SettingsController',
        },
      },
    };

    var settingsAuthentication = {
      name: 'portainer.settings.authentication',
      url: '/auth',
      views: {
        'content@': {
          templateUrl: './views/settings/authentication/settingsAuthentication.html',
          controller: 'SettingsAuthenticationController',
        },
      },
    };

    var support = {
      name: 'portainer.support',
      url: '/support',
      views: {
        'content@': {
          templateUrl: './views/support/support.html',
          controller: 'SupportController',
        },
      },
      params: {
        product: {},
      },
    };

    var supportProduct = {
      name: 'portainer.support.product',
      url: '/product',
      views: {
        'content@': {
          templateUrl: './views/support/product/product.html',
          controller: 'SupportProductController',
        },
      },
    };

    var tags = {
      name: 'portainer.tags',
      url: '/tags',
      views: {
        'content@': {
          templateUrl: './views/tags/tags.html',
          controller: 'TagsController',
        },
      },
    };

    var users = {
      name: 'portainer.users',
      url: '/users',
      views: {
        'content@': {
          templateUrl: './views/users/users.html',
          controller: 'UsersController',
        },
      },
    };

    var user = {
      name: 'portainer.users.user',
      url: '/:id',
      views: {
        'content@': {
          templateUrl: './views/users/edit/user.html',
          controller: 'UserController',
        },
      },
    };

    var teams = {
      name: 'portainer.teams',
      url: '/teams',
      views: {
        'content@': {
          templateUrl: './views/teams/teams.html',
          controller: 'TeamsController',
        },
      },
    };

    var team = {
      name: 'portainer.teams.team',
      url: '/:id',
      views: {
        'content@': {
          templateUrl: './views/teams/edit/team.html',
          controller: 'TeamController',
        },
      },
    };

    $stateRegistryProvider.register(root);
    $stateRegistryProvider.register(endpointRoot);
    $stateRegistryProvider.register(portainer);
    $stateRegistryProvider.register(about);
    $stateRegistryProvider.register(account);
    $stateRegistryProvider.register(authentication);
    $stateRegistryProvider.register(logout);
    $stateRegistryProvider.register(endpoints);
    $stateRegistryProvider.register(endpoint);
    $stateRegistryProvider.register(endpointAccess);
    $stateRegistryProvider.register(endpointCreation);
    $stateRegistryProvider.register(endpointKubernetesConfiguration);
    $stateRegistryProvider.register(groups);
    $stateRegistryProvider.register(group);
    $stateRegistryProvider.register(groupAccess);
    $stateRegistryProvider.register(groupCreation);
    $stateRegistryProvider.register(home);
    $stateRegistryProvider.register(init);
    $stateRegistryProvider.register(initEndpoint);
    $stateRegistryProvider.register(initAdmin);
    $stateRegistryProvider.register(extensions);
    $stateRegistryProvider.register(extension);
    $stateRegistryProvider.register(registries);
    $stateRegistryProvider.register(registry);
    $stateRegistryProvider.register(registryAccess);
    $stateRegistryProvider.register(registryCreation);
    $stateRegistryProvider.register(settings);
    $stateRegistryProvider.register(settingsAuthentication);
    $stateRegistryProvider.register(support);
    $stateRegistryProvider.register(supportProduct);
    $stateRegistryProvider.register(tags);
    $stateRegistryProvider.register(users);
    $stateRegistryProvider.register(user);
    $stateRegistryProvider.register(teams);
    $stateRegistryProvider.register(team);
  },
]);

angular.module('portainer.kubernetes', ['portainer.app'])
.config(['$stateRegistryProvider', function ($stateRegistryProvider) {
  'use strict';

  const kubernetes = {
    name: 'kubernetes',
    url: '/kubernetes',
    parent: 'root',
    abstract: true,
    resolve: {
      endpointID: ['EndpointProvider', '$state',
        function (EndpointProvider, $state) {
          const id = EndpointProvider.endpointID();
          if (!id) {
            return $state.go('portainer.home');
          }
        }
      ]
    }
  };

  const dashboard = {
    name: 'kubernetes.dashboard',
    url: '/dashboard',
    views: {
      'content@': {
        templateUrl: './views/dashboard/dashboard.html',
        controller: 'KubernetesDashboardController'
      }
    }
  };

  const namespaces = {
    name: 'kubernetes.namespaces',
    url: '/namespaces',
    views: {
      'content@': {
        templateUrl: './views/namespaces/namespaces.html',
        controller: 'KubernetesNamespacesController',
        controllerAs: 'ctrl'
      }
    }
  }

  const services = {
    name: 'kubernetes.services',
    url: '/services',
    views: {
      'content@': {
        templateUrl: './views/services/services.html',
        controller: 'KubernetesServicesController',
        controllerAs: 'ctrl'
      }
    }
  };

  const containers = {
    name: 'kubernetes.containers',
    url: '/containers',
    views: {
      'content@': {
        templateUrl: './views/containers/containers.html',
        controller: 'KubernetesContainersController',
        controllerAs: 'ctrl'
      }
    }
  };

  const configs = {
    name: 'kubernetes.configs',
    url: '/configs',
    views: {
      'content@': {
        templateUrl: './views/configs/configs.html',
        controller: 'KubernetesConfigsController',
        controllerAs: 'ctrl'
      }
    }
  };

  const secrets = {
    name: 'kubernetes.secrets',
    url: '/secrets',
    views: {
      'content@': {
        templateUrl: './views/secrets/secrets.html',
        controller: 'KubernetesSecretsController',
        controllerAs: 'ctrl'
      }
    }
  };

  const nodes = {
    name: 'kubernetes.nodes',
    url: '/nodes',
    views: {
      'content@': {
        templateUrl: './views/nodes/nodes.html',
        controller: 'KubernetesNodesController',
        controllerAs: 'ctrl'
      }
    }
  };

  $stateRegistryProvider.register(kubernetes);
  $stateRegistryProvider.register(dashboard);
  $stateRegistryProvider.register(namespaces);
  $stateRegistryProvider.register(services);
  $stateRegistryProvider.register(containers);
  $stateRegistryProvider.register(configs);
  $stateRegistryProvider.register(secrets);
  $stateRegistryProvider.register(nodes);
}]);

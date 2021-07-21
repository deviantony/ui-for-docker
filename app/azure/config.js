/* @ngInject */
export default function config($stateRegistryProvider) {
  const azure = {
    name: 'azure',
    url: '/azure',
    parent: 'endpoint',
    abstract: true,
    onEnter: /* @ngInject */ function onEnter($async, $state, endpoint, EndpointProvider, Notifications, StateManager) {
      return $async(async () => {
        if (endpoint.Type !== 3) {
          $state.go('portainer.home');
          return;
        }
        try {
          EndpointProvider.setEndpointID(endpoint.Id);
          EndpointProvider.setEndpointPublicURL(endpoint.PublicURL);
          EndpointProvider.setOfflineModeFromStatus(endpoint.Status);
          await StateManager.updateEndpointState(endpoint, []);
        } catch (e) {
          Notifications.error('Failed loading endpoint', e);
          $state.go('portainer.home', {}, { reload: true });
        }
      });
    },
  };

  const containerInstances = {
    name: 'azure.containerinstances',
    url: '/containerinstances',
    views: {
      'content@': {
        component: 'containerInstancesView',
      },
    },
  };

  const containerInstance = {
    name: 'azure.containerinstances.container',
    url: '/:id',
    views: {
      'content@': {
        component: 'containerInstanceDetailsView',
      },
    },
  };

  const containerInstanceCreation = {
    name: 'azure.containerinstances.new',
    url: '/new/',
    views: {
      'content@': {
        component: 'createContainerInstanceView',
      },
    },
  };

  const dashboard = {
    name: 'azure.dashboard',
    url: '/dashboard',
    views: {
      'content@': {
        component: 'azureDashboardView',
      },
    },
  };

  $stateRegistryProvider.register(azure);
  $stateRegistryProvider.register(containerInstances);
  $stateRegistryProvider.register(containerInstance);
  $stateRegistryProvider.register(containerInstanceCreation);
  $stateRegistryProvider.register(dashboard);
}

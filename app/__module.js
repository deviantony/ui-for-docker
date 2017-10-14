angular.module('portainer', [
  'ui.bootstrap',
  'ui.router',
  'isteven-multi-select',
  'ngCookies',
  'ngSanitize',
  'ngFileUpload',
  'angularUtils.directives.dirPagination',
  'LocalStorageModule',
  'angular-jwt',
  'angular-google-analytics',
  'portainer.templates',
  'portainer.filters',
  'portainer.rest',
  'portainer.helpers',
  'portainer.services',
  'auth',
  'dashboard',
  'container',
  'containerConsole',
  'containerLogs',
  'containerStats',
  'serviceLogs',
  'containers',
  'createContainer',
  'createNetwork',
  'createRegistry',
  'createSecret',
  'createService',
  'createVolume',
  'createStack',
  'engine',
  'endpoint',
  'endpointAccess',
  'endpoints',
  'events',
  'image',
  'images',
  'initAdmin',
  'initEndpoint',
  'main',
  'network',
  'networks',
  'node',
  'registries',
  'registry',
  'registryAccess',
  'secrets',
  'secret',
  'service',
  'services',
  'settings',
  'settingsAuthentication',
  'sidebar',
  'unregisteredstackv2',
  'stackv2',
  'stackv3',
  'stacks',
  'swarm',
  'swarmVisualizer',
  'task',
  'team',
  'teams',
  'templates',
  'user',
  'users',
  'userSettings',
  'volume',
  'volumes',
  'rzModule']);

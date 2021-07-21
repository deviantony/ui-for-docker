angular.module('portainer.app').factory('Stack', [
  '$resource',
  'EndpointProvider',
  'API_ENDPOINT_STACKS',
  function StackFactory($resource, EndpointProvider, API_ENDPOINT_STACKS) {
    'use strict';
    return $resource(
      API_ENDPOINT_STACKS + '/:id/:action',
      {},
      {
        get: { method: 'GET', params: { id: '@id' } },
        query: { method: 'GET', isArray: true },
        create: { method: 'POST', ignoreLoadingBar: true },
        update: { method: 'PUT', params: { id: '@id' }, ignoreLoadingBar: true },
        associate: { method: 'PUT', params: { id: '@id', swarmId: '@swarmId', endpointId: '@endpointId', orphanedRunning: '@orphanedRunning', action: 'associate' } },
        remove: { method: 'DELETE', params: { id: '@id', external: '@external', endpointId: '@endpointId' } },
        getStackFile: { method: 'GET', params: { id: '@id', action: 'file' } },
        migrate: { method: 'POST', params: { id: '@id', action: 'migrate', endpointId: '@endpointId' }, ignoreLoadingBar: true },
        start: { method: 'POST', params: { id: '@id', action: 'start' } },
        stop: { method: 'POST', params: { id: '@id', action: 'stop' } },
        updateGit: { method: 'PUT', params: { id: '@id', action: 'git/redeploy' } },
        updateGitStackSettings: { method: 'POST', params: { id: '@id', action: 'git' }, ignoreLoadingBar: true },
      }
    );
  },
]);

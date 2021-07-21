angular.module('portainer.kubernetes').component('kubernetesResourcePoolsView', {
  templateUrl: './resourcePools.html',
  controller: 'KubernetesResourcePoolsController',
  controllerAs: 'ctrl',
  bindings: {
    endpoint: '<',
  },
});

angular.module('portainer.kubernetes').component('kubernetesNodeView', {
  templateUrl: './node.html',
  controller: 'KubernetesNodeController',
  controllerAs: 'ctrl',
  bindings: {
    endpoint: '<',
    $transition$: '<',
  },
});

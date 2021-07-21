angular.module('portainer.kubernetes').component('kubernetesClusterView', {
  templateUrl: './cluster.html',
  controller: 'KubernetesClusterController',
  controllerAs: 'ctrl',
  bindings: {
    endpoint: '<',
  },
});

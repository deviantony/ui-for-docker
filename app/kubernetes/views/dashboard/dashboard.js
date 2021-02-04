angular.module('portainer.kubernetes').component('kubernetesDashboardView', {
  templateUrl: './dashboard.html',
  controller: 'KubernetesDashboardController',
  controllerAs: 'ctrl',
  bindings: {
    endpoint: '<',
  },
});

angular.module('portainer.kubernetes').component('kubernetesResourcePoolsDatatable', {
  templateUrl: './resourcePoolsDatatable.html',
  controller: 'KubernetesResourcePoolsDatatableController',
  bindings: {
    titleText: '@',
    titleIcon: '@',
    dataset: '<',
    tableKey: '@',
    orderBy: '@',
    reverseOrder: '<',
    removeAction: '<',
    refreshCallback: '<',
    endpoint: '<',
  },
});

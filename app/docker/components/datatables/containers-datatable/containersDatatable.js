angular.module('portainer.docker').component('containersDatatable', {
  templateUrl: './containersDatatable.html',
  controller: 'ContainersDatatableController',
  bindings: {
    titleText: '@',
    titleIcon: '@',
    dataset: '<',
    tableKey: '@',
    orderBy: '@',
    reverseOrder: '<',
    showHostColumn: '<',
    showAddAction: '<',
    offlineMode: '<',
    refreshCallback: '<',
    notAutoFocus: '<',
  },
});

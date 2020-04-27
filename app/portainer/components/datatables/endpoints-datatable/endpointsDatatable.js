angular.module('portainer.app').component('endpointsDatatable', {
  templateUrl: './endpointsDatatable.html',
  controller: 'EndpointsDatatableController',
  bindings: {
    titleText: '@',
    titleIcon: '@',
    tableKey: '@',
    orderBy: '@',
    reverseOrder: '<',
    endpointManagement: '<',
    accessManagement: '<',
    removeAction: '<',
    retrievePage: '<',
  },
});

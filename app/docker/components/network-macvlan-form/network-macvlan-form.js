angular.module('portainer.docker').component('networkMacvlanForm', {
  templateUrl: './networkMacvlanForm.html',
  controller: 'NetworkMacvlanFormController',
  bindings: {
    data: '=',
    applicationState: '<',
    driverName: '@',
  },
});

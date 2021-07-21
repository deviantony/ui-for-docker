import angular from 'angular';

angular.module('portainer.edge').component('edgeGroupsSelector', {
  templateUrl: './edgeGroupsSelector.html',
  bindings: {
    model: '<',
    items: '<',
    onChange: '<',
  },
});

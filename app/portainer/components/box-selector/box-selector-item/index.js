import angular from 'angular';

angular.module('portainer.app').component('boxSelectorItem', {
  templateUrl: './box-selector-item.html',
  bindings: {
    radioName: '@',
    isChecked: '<',
    option: '<',
    onChange: '<',
    disabled: '<',
    tooltip: '<',
  },
});

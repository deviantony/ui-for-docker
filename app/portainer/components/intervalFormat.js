import parse from 'parse-duration';
angular.module('portainer.app').directive('intervalFormat', function () {
  return {
    restrict: 'A',
    require: 'ngModel',
    link: function ($scope, $element, $attrs, ngModel) {
      ngModel.$validators.invalidIntervalFormat = function (modelValue) {
        try {
          return modelValue && modelValue.toUpperCase().match(/^P?(?!$)(\d+Y)?(\d+M)?(\d+W)?(\d+D)?(T?(?=\d+[HMS])(\d+H)?(\d+M)?(\d+S)?)?$/gm) !== null;
        } catch (error) {
          return false;
        }
      };

      ngModel.$validators.minimumInterval = function (modelValue) {
        try {
          return modelValue && parse(modelValue, 'minute') >= 1;
        } catch (error) {
          return false;
        }
      };
    },
  };
});

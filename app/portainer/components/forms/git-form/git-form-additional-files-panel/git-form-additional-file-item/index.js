import angular from 'angular';
import controller from './git-form-additional-file-item.controller.js';

export const gitFormAdditionalFileItem = {
  templateUrl: './git-form-additional-file-item.html',
  controller,

  bindings: {
    variable: '<',
    index: '<',

    onChange: '<',
    onRemove: '<',
  },
};

angular.module('portainer.app').component('gitFormAdditionalFileItem', gitFormAdditionalFileItem);

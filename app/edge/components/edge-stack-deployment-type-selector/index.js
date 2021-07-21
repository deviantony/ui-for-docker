import angular from 'angular';
import controller from './edge-stack-deployment-type-selector.controller.js';

export const edgeStackDeploymentTypeSelector = {
  templateUrl: './edge-stack-deployment-type-selector.html',
  controller,

  bindings: {
    value: '<',
    onChange: '<',
    hasDockerEndpoint: '<',
  },
};

angular.module('portainer.edge').component('edgeStackDeploymentTypeSelector', edgeStackDeploymentTypeSelector);

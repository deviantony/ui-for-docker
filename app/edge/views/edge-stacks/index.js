import angular from 'angular';

import createModule from './createEdgeStackView';

export default angular.module('portainer.edge.stacks', [createModule]).name;

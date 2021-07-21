import angular from 'angular';

import config from './config';

import componentsModule from './components';
import restModule from './rest';
import servicesModule from './services';
import viewsModule from './views';

export default angular.module('portainer.azure', [componentsModule, restModule, servicesModule, viewsModule]).config(config).name;

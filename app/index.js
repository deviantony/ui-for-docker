import './assets/css';
import '@babel/polyfill';

import angular from 'angular';

import './matomo-setup';
import './assets/js/angulartics-matomo';

import './agent';
import './azure/_module';
import './docker/__module';
import './edge/__module';
import './portainer/__module';

import { onStartupAngular } from './app';
import { configApp } from './config';

angular
  .module('portainer', [
    'ui.bootstrap',
    'ui.router',
    'ui.select',
    'isteven-multi-select',
    'ngSanitize',
    'ngFileUpload',
    'ngMessages',
    'ngResource',
    'angularUtils.directives.dirPagination',
    'LocalStorageModule',
    'angular-jwt',
    'angular-json-tree',
    'angular-loading-bar',
    'angular-clipboard',
    'ngFileSaver',
    'luegg.directives',
    'portainer.app',
    'portainer.agent',
    'portainer.azure',
    'portainer.docker',
    'portainer.kubernetes',
    'portainer.edge',
    'portainer.integrations',
    'rzModule',
    'moment-picker',
    'angulartics',
    'angulartics.matomo',
  ])
  .run(onStartupAngular)
  .config(configApp);

if (require) {
  var req = require.context('./', true, /^(.*\.(js$))[^.]*$/im);
  req.keys().forEach(function (key) {
    req(key);
  });
}

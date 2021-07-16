import angular from 'angular';
import controller from './kubectl-shell.controller';
import './kubectl-shell.css';

angular.module('portainer.kubernetes').component('kubectlShell', {
  templateUrl: './kubectl-shell.html',
  controller,
});

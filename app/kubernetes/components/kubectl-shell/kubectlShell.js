import './kubectlShell.css';

angular.module('portainer.kubernetes').component('kubectlShell', {
  templateUrl: './kubectlShell.html',
  controller: 'KubectlShellController',
  controllerAs: 'ctrl',
});

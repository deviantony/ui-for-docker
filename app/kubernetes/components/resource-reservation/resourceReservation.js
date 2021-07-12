angular.module('portainer.kubernetes').component('kubernetesResourceReservation', {
  templateUrl: './resourceReservation.html',
  controller: 'KubernetesResourceReservationController',
  bindings: {
    description: '@',
    cpuReservation: '<',
    cpuUsage: '<',
    cpuLimit: '<',
    memoryReservation: '<',
    memoryUsage: '<',
    memoryLimit: '<',
    displayUsage: '<',
  },
});

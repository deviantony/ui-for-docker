import angular from 'angular';

class KubernetesResourceReservationController {
  usageValues() {
    if (this.cpuLimit) {
      this.cpuReservationPercent = Math.round((this.cpuReservation / this.cpuLimit) * 100);
    }
    if (this.memoryLimit) {
      this.memoryReservationPercent = Math.round((this.memoryReservation / this.memoryLimit) * 100);
    }

    if (this.displayUsage && this.cpuLimit && this.memoryLimit) {
      this.cpuUsagePercent = Math.round((this.cpuUsage / this.cpuLimit) * 100);
      this.memoryUsagePercent = Math.round((this.memoryUsage / this.memoryLimit) * 100);
    }
  }

  $onInit() {
    this.usageValues();
  }

  $onChanges() {
    this.usageValues();
  }
}

export default KubernetesResourceReservationController;
angular.module('portainer.kubernetes').controller('KubernetesResourceReservationController', KubernetesResourceReservationController);

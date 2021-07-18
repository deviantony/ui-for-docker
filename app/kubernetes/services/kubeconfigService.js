import angular from 'angular';

class KubernetesConfigService {
  /* @ngInject */
  constructor(KubernetesConfig, FileSaver) {
    this.KubernetesConfig = KubernetesConfig;
    this.FileSaver = FileSaver;
  }

  async downloadConfig() {
    const response = await this.KubernetesConfig.get();
    return this.FileSaver.saveAs(response.data, 'config');
  }
}

export default KubernetesConfigService;
angular.module('portainer.kubernetes').service('KubernetesConfigService', KubernetesConfigService);

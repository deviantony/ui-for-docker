import angular from 'angular';

class KubernetesConfigService {
  /* @ngInject */
  constructor(KubernetesConfig, FileSaver) {
    this.KubernetesConfig = KubernetesConfig;
    this.FileSaver = FileSaver;
  }

  async downloadConfig() {
    try {
      const response = await this.KubernetesConfig.get();
      return this.FileSaver.saveAs(response.data, 'config.yaml');
    } catch (err) {
      throw err;
    }
  }
}

export default KubernetesConfigService;
angular.module('portainer.kubernetes').service('KubernetesConfigService', KubernetesConfigService);

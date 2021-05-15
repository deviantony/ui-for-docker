import angular from 'angular';
import _ from 'lodash-es';
import KubernetesConfigurationConverter from 'Kubernetes/converters/configuration';
import KubernetesConfigMapConverter from 'Kubernetes/converters/configMap';
import KubernetesSecretConverter from 'Kubernetes/converters/secret';
import { KubernetesConfigurationTypes } from 'Kubernetes/models/configuration/models';
import KubernetesCommonHelper from 'Kubernetes/helpers/commonHelper';

class KubernetesConfigurationService {
  /* @ngInject */
  constructor($async, Authentication, KubernetesNamespaceService, KubernetesConfigMapService, KubernetesSecretService) {
    this.$async = $async;
    this.Authentication = Authentication;
    this.KubernetesNamespaceService = KubernetesNamespaceService;
    this.KubernetesConfigMapService = KubernetesConfigMapService;
    this.KubernetesSecretService = KubernetesSecretService;

    this.getAsync = this.getAsync.bind(this);
    this.getAllAsync = this.getAllAsync.bind(this);
    this.createAsync = this.createAsync.bind(this);
    this.updateAsync = this.updateAsync.bind(this);
    this.deleteAsync = this.deleteAsync.bind(this);
  }

  /**
   * GET
   */
  async getAsync(namespace, name) {
    const [configMap, secret] = await Promise.allSettled([this.KubernetesConfigMapService.get(namespace, name), this.KubernetesSecretService.get(namespace, name)]);
    let configuration;
    if (secret.status === 'fulfilled') {
      configuration = KubernetesConfigurationConverter.secretToConfiguration(secret.value);
      return configuration;
    }
    configuration = KubernetesConfigurationConverter.configMapToConfiguration(configMap.value);
    return configuration;
  }

  async getAllAsync(namespace) {
    const namespaces = namespace ? [namespace] : _.map(await this.KubernetesNamespaceService.get(), 'Name');
    const res = await Promise.all(
      _.map(namespaces, async (ns) => {
        const [configMaps, secrets] = await Promise.all([this.KubernetesConfigMapService.get(ns), this.KubernetesSecretService.get(ns)]);
        const secretsConfigurations = _.map(secrets, (secret) => KubernetesConfigurationConverter.secretToConfiguration(secret));
        const configMapsConfigurations = _.map(configMaps, (configMap) => KubernetesConfigurationConverter.configMapToConfiguration(configMap));
        return _.concat(configMapsConfigurations, secretsConfigurations);
      })
    );
    return _.flatten(res);
  }

  get(namespace, name) {
    if (name) {
      return this.$async(this.getAsync, namespace, name);
    }
    return this.$async(this.getAllAsync, namespace);
  }

  /**
   * CREATE
   */
  async createAsync(formValues) {
    formValues.ConfigurationOwner = KubernetesCommonHelper.ownerToLabel(formValues.ConfigurationOwner);

    if (formValues.Type === KubernetesConfigurationTypes.CONFIGMAP) {
      const configMap = KubernetesConfigMapConverter.configurationFormValuesToConfigMap(formValues);
      await this.KubernetesConfigMapService.create(configMap);
    } else {
      const secret = KubernetesSecretConverter.configurationFormValuesToSecret(formValues);
      await this.KubernetesSecretService.create(secret);
    }
  }

  create(formValues) {
    return this.$async(this.createAsync, formValues);
  }

  /**
   * UPDATE
   */
  async updateAsync(formValues, configuration) {
    if (formValues.Type === KubernetesConfigurationTypes.CONFIGMAP) {
      const configMap = KubernetesConfigMapConverter.configurationFormValuesToConfigMap(formValues);
      configMap.ConfigurationOwner = configuration.ConfigurationOwner;
      await this.KubernetesConfigMapService.update(configMap);
    } else {
      const secret = KubernetesSecretConverter.configurationFormValuesToSecret(formValues);
      secret.ConfigurationOwner = configuration.ConfigurationOwner;
      await this.KubernetesSecretService.update(secret);
    }
  }

  update(formValues, configuration) {
    return this.$async(this.updateAsync, formValues, configuration);
  }

  /**
   * DELETE
   */
  async deleteAsync(config) {
    if (config.Type === KubernetesConfigurationTypes.CONFIGMAP) {
      await this.KubernetesConfigMapService.delete(config);
    } else {
      await this.KubernetesSecretService.delete(config);
    }
  }

  delete(config) {
    return this.$async(this.deleteAsync, config);
  }
}

export default KubernetesConfigurationService;
angular.module('portainer.kubernetes').service('KubernetesConfigurationService', KubernetesConfigurationService);

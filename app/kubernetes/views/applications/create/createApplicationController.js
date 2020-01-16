import angular from 'angular';
import _ from 'lodash-es';
import filesizeParser from 'filesize-parser';
import {
  KubernetesApplicationDeploymentTypes,
  KubernetesApplicationEnvironmentVariableFormValue,
  KubernetesApplicationFormValues,
  KubernetesApplicationPersistedFolderFormValue,
  KubernetesApplicationPublishedPortFormValue,
  KubernetesApplicationPublishingTypes
} from 'Kubernetes/models/application';
import {KubernetesPortainerQuotaSuffix} from 'Kubernetes/models/resourceQuota';
import {KubernetesLimitRangeDefaults} from 'Kubernetes/models/limitRange';


function megaBytesValue(mem) {
  return Math.floor(mem / 1000 / 1000);
}
class KubernetesCreateApplicationController {
  /* @ngInject */
  constructor($async, $state, Notifications, EndpointProvider, KubernetesResourcePoolService, KubernetesApplicationService, KubernetesStackService, KubernetesNodeService) {
    this.$async = $async;
    this.$state = $state;
    this.Notifications = Notifications;
    this.EndpointProvider = EndpointProvider;
    this.KubernetesResourcePoolService = KubernetesResourcePoolService;
    this.KubernetesApplicationService = KubernetesApplicationService;
    this.KubernetesStackService = KubernetesStackService;
    this.KubernetesNodeService = KubernetesNodeService;

    this.onInit = this.onInit.bind(this);
    this.deployApplicationAsync = this.deployApplicationAsync.bind(this);
  }

  addEnvironmentVariable() {
    this.formValues.EnvironmentVariables.push(new KubernetesApplicationEnvironmentVariableFormValue());
  }

  removeEnvironmentVariable(index) {
    this.formValues.EnvironmentVariables.splice(index, 1);
  }

  hasEnvironmentVariables() {
    return this.formValues.EnvironmentVariables.length > 0;
  }

  addPersistedFolder() {
    let storageClass = '';
    if (this.storageClasses.length === 1) {
      storageClass = this.storageClasses[0];
    }

    this.formValues.PersistedFolders.push(new KubernetesApplicationPersistedFolderFormValue(storageClass));
  }

  removePersistedFolder(index) {
    this.formValues.PersistedFolders.splice(index, 1);
  }

  addPublishedPort() {
    this.formValues.PublishedPorts.push(new KubernetesApplicationPublishedPortFormValue());
  }

  removePublishedPort(index) {
    this.formValues.PublishedPorts.splice(index, 1);
  }

  storageClassAvailable() {
    return this.storageClasses && this.storageClasses.length > 0;
  }

  hasMultipleStorageClassesAvailable() {
    return this.storageClasses && this.storageClasses.length > 1;
  }

  publishViaLoadBalancerEnabled() {
    return this.state.useLoadBalancer;
  }

  async deployApplicationAsync() {
    this.state.actionInProgress = true;
    try {
      await this.KubernetesApplicationService.create(this.formValues);
      this.Notifications.success('Application successfully deployed', this.formValues.Name);
      this.$state.go('kubernetes.applications');
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to create application');
    } finally {
      this.state.actionInProgress = false;
    }
  }

  deployApplication() {
    return this.$async(this.deployApplicationAsync);
  }

  onSelectResourcePool() {
    // TODO: review @LP
    // slider UI (max memory/CPU) not updated after the execution of this function

    // const quota = _.find(this.formValues.ResourcePool.Quotas, (item) => item.Name === KubernetesPortainerQuotaSuffix + this.formValues.ResourcePool.Namespace.Name);
    // if (quota) {
    //   this.state.maxMemory = megaBytesValue(quota.MemoryLimit);
    //   this.state.maxCpu = quota.CpuLimit;
    //
    //   this.formValues.MemoryLimit = this.KubernetesLimitRangeDefaults.MemoryLimit;
    //   this.formValues.CpuLimit = this.KubernetesLimitRangeDefaults.CpuLimit;
    // } else {
    //   this.formValues.CpuLimit = 0;
    //   this.formValues.MemoryLimit = 0;
    //   _.forEach(this.nodes, (item) => {
    //     this.state.maxMemory += filesizeParser(item.Memory);
    //     this.state.maxCpu += item.CPU;
    //   });
    //   this.state.maxMemory = megaBytesValue(this.state.maxMemory);
    // }
  }

  async onInit() {
    try {
      this.formValues = new KubernetesApplicationFormValues();

      this.state = {
        actionInProgress: false,
        useLoadBalancer: false,
        maxCpu: 0,
        maxMemory: 0
      };

      this.ApplicationDeploymentTypes = KubernetesApplicationDeploymentTypes;
      this.ApplicationPublishingTypes = KubernetesApplicationPublishingTypes;
      this.KubernetesLimitRangeDefaults = KubernetesLimitRangeDefaults;

      const [resourcePools, stacks, nodes] = await Promise.all([
        this.KubernetesResourcePoolService.resourcePools(),
        this.KubernetesStackService.stacks(),
        this.KubernetesNodeService.nodes()
      ]);

      this.stacks = stacks;
      this.nodes = nodes;
      this.resourcePools = resourcePools;
      this.formValues.ResourcePool = this.resourcePools[0];
      const quota = _.find(this.formValues.ResourcePool.Quotas, (item) => item.Name === KubernetesPortainerQuotaSuffix + this.formValues.ResourcePool.Namespace.Name);
      if (quota) {
        this.state.maxMemory = quota.MemoryLimit;
        this.state.maxCpu = quota.CpuLimit;
      } else {
        _.forEach(nodes, (item) => {
          this.state.maxMemory += filesizeParser(item.Memory);
          this.state.maxCpu += item.CPU;
        });
        this.state.maxMemory = megaBytesValue(this.state.maxMemory);
      }

      const endpoint = this.EndpointProvider.currentEndpoint();
      this.storageClasses = endpoint.Kubernetes.Configuration.StorageClasses;
      this.state.useLoadBalancer = endpoint.Kubernetes.Configuration.UseLoadBalancer;
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to load view data');
    }
  }

  $onInit() {
    return this.$async(this.onInit);
  }
}

export default KubernetesCreateApplicationController;
angular.module('portainer.kubernetes').controller('KubernetesCreateApplicationController', KubernetesCreateApplicationController);

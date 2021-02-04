import angular from 'angular';
import _ from 'lodash-es';

import KubernetesConfigurationHelper from 'Kubernetes/helpers/configurationHelper';
import { getEndpointTags } from '@/portainer/helpers/tagHelper';

class KubernetesDashboardController {
  /* @ngInject */
  constructor(
    $async,
    Notifications,
    GroupService,
    KubernetesResourcePoolService,
    KubernetesApplicationService,
    KubernetesConfigurationService,
    KubernetesVolumeService,
    KubernetesNamespaceHelper,
    Authentication,
    TagService
  ) {
    this.$async = $async;
    this.Notifications = Notifications;
    this.GroupService = GroupService;
    this.KubernetesResourcePoolService = KubernetesResourcePoolService;
    this.KubernetesApplicationService = KubernetesApplicationService;
    this.KubernetesConfigurationService = KubernetesConfigurationService;
    this.KubernetesVolumeService = KubernetesVolumeService;
    this.KubernetesNamespaceHelper = KubernetesNamespaceHelper;
    this.Authentication = Authentication;
    this.TagService = TagService;

    this.onInit = this.onInit.bind(this);
    this.getAll = this.getAll.bind(this);
    this.getAllAsync = this.getAllAsync.bind(this);
  }

  async getAllAsync() {
    const isAdmin = this.Authentication.isAdmin();

    try {
      const [endpointGroup, pools, applications, configurations, volumes, tags] = await Promise.all([
        this.GroupService.group(this.endpoint.GroupId),
        this.KubernetesResourcePoolService.get(),
        this.KubernetesApplicationService.get(),
        this.KubernetesConfigurationService.get(),
        this.KubernetesVolumeService.get(),
        this.TagService.tags(),
      ]);
      this.applications = applications;
      this.volumes = volumes;

      const endpointTags = getEndpointTags(this.endpoint, endpointGroup, tags);
      this.endpointTags = _.map(endpointTags, 'Name').join(', ') || '-';

      if (!isAdmin) {
        this.pools = _.filter(pools, (pool) => {
          return !this.KubernetesNamespaceHelper.isSystemNamespace(pool.Namespace.Name);
        });

        this.configurations = _.filter(configurations, (config) => {
          return !KubernetesConfigurationHelper.isSystemToken(config);
        });
      } else {
        this.pools = pools;
        this.configurations = configurations;
      }
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to load dashboard data');
    }
  }

  getAll() {
    return this.$async(this.getAllAsync);
  }

  async onInit() {
    this.state = {
      viewReady: false,
    };

    await this.getAll();

    this.state.viewReady = true;
  }

  $onInit() {
    return this.$async(this.onInit);
  }
}

export default KubernetesDashboardController;
angular.module('portainer.kubernetes').controller('KubernetesDashboardController', KubernetesDashboardController);

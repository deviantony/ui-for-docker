import angular from 'angular';
import _ from 'lodash-es';
import filesizeParser from 'filesize-parser';
import { KubernetesResourceQuotaDefaults } from 'Kubernetes/models/resource-quota/models';
import KubernetesResourceReservationHelper from 'Kubernetes/helpers/resourceReservationHelper';
import { KubernetesResourceReservation } from 'Kubernetes/models/resource-reservation/models';
import KubernetesEventHelper from 'Kubernetes/helpers/eventHelper';
import {
  KubernetesResourcePoolFormValues,
  KubernetesResourcePoolIngressClassAnnotationFormValue,
  KubernetesResourcePoolIngressClassHostFormValue,
} from 'Kubernetes/models/resource-pool/formValues';
import { KubernetesIngressConverter } from 'Kubernetes/ingress/converter';
import { KubernetesFormValidationReferences } from 'Kubernetes/models/application/formValues';
import KubernetesFormValidationHelper from 'Kubernetes/helpers/formValidationHelper';
import { KubernetesIngressClassTypes } from 'Kubernetes/ingress/constants';
import KubernetesResourceQuotaConverter from 'Kubernetes/converters/resourceQuota';

class KubernetesResourcePoolController {
  /* #region  CONSTRUCTOR */
  /* @ngInject */
  constructor(
    $async,
    $state,
    Authentication,
    Notifications,
    LocalStorage,
    EndpointService,
    ModalService,
    KubernetesNodeService,
    KubernetesMetricsService,
    KubernetesResourceQuotaService,
    KubernetesResourcePoolService,
    KubernetesEventService,
    KubernetesPodService,
    KubernetesApplicationService,
    KubernetesNamespaceHelper,
    KubernetesIngressService,
    KubernetesVolumeService
  ) {
    Object.assign(this, {
      $async,
      $state,
      Authentication,
      Notifications,
      LocalStorage,
      EndpointService,
      ModalService,
      KubernetesNodeService,
      KubernetesMetricsService,
      KubernetesResourceQuotaService,
      KubernetesResourcePoolService,
      KubernetesEventService,
      KubernetesPodService,
      KubernetesApplicationService,
      KubernetesNamespaceHelper,
      KubernetesIngressService,
      KubernetesVolumeService,
    });

    this.IngressClassTypes = KubernetesIngressClassTypes;
    this.ResourceQuotaDefaults = KubernetesResourceQuotaDefaults;

    this.updateResourcePoolAsync = this.updateResourcePoolAsync.bind(this);
    this.getEvents = this.getEvents.bind(this);
  }
  /* #endregion */

  /* #region  ANNOTATIONS MANAGEMENT */
  addAnnotation(ingressClass) {
    ingressClass.Annotations.push(new KubernetesResourcePoolIngressClassAnnotationFormValue());
  }

  removeAnnotation(ingressClass, index) {
    ingressClass.Annotations.splice(index, 1);
    this.onChangeIngressHostname();
  }
  /* #endregion */

  /* #region  INGRESS MANAGEMENT */
  onChangeIngressHostname() {
    const state = this.state.duplicates.ingressHosts;
    const otherIngresses = _.without(this.allIngresses, ...this.ingresses);
    const allHosts = _.flatMap(otherIngresses, 'Hosts');

    const hosts = _.flatMap(this.formValues.IngressClasses, 'Hosts');
    const hostsWithoutRemoved = _.filter(hosts, { NeedsDeletion: false });
    const hostnames = _.map(hostsWithoutRemoved, 'Host');
    const formDuplicates = KubernetesFormValidationHelper.getDuplicates(hostnames);
    _.forEach(hostnames, (host, idx) => {
      if (host !== undefined && _.includes(allHosts, host)) {
        formDuplicates[idx] = host;
      }
    });
    const duplicatedHostnames = Object.values(formDuplicates);
    state.hasRefs = false;
    _.forEach(this.formValues.IngressClasses, (ic) => {
      _.forEach(ic.Hosts, (hostFV) => {
        if (_.includes(duplicatedHostnames, hostFV.Host) && hostFV.NeedsDeletion === false) {
          hostFV.Duplicate = true;
          state.hasRefs = true;
        } else {
          hostFV.Duplicate = false;
        }
      });
    });
  }

  addHostname(ingressClass) {
    ingressClass.Hosts.push(new KubernetesResourcePoolIngressClassHostFormValue());
  }

  removeHostname(ingressClass, index) {
    if (!ingressClass.Hosts[index].IsNew) {
      ingressClass.Hosts[index].NeedsDeletion = true;
    } else {
      ingressClass.Hosts.splice(index, 1);
    }
    this.onChangeIngressHostname();
  }

  restoreHostname(host) {
    if (!host.IsNew) {
      host.NeedsDeletion = false;
    }
  }
  /* #endregion*/

  selectTab(index) {
    this.LocalStorage.storeActiveTab('resourcePool', index);
  }

  isUpdateButtonDisabled() {
    return this.state.actionInProgress || (this.formValues.HasQuota && !this.isQuotaValid()) || this.state.duplicates.ingressHosts.hasRefs;
  }

  isQuotaValid() {
    if (
      this.state.sliderMaxCpu < this.formValues.CpuLimit ||
      this.state.sliderMaxMemory < this.formValues.MemoryLimit ||
      (this.formValues.CpuLimit === 0 && this.formValues.MemoryLimit === 0)
    ) {
      return false;
    }
    return true;
  }

  checkDefaults() {
    if (this.formValues.CpuLimit < KubernetesResourceQuotaDefaults.CpuLimit) {
      this.formValues.CpuLimit = KubernetesResourceQuotaDefaults.CpuLimit;
    }
    if (this.formValues.MemoryLimit < KubernetesResourceReservationHelper.megaBytesValue(KubernetesResourceQuotaDefaults.MemoryLimit)) {
      this.formValues.MemoryLimit = KubernetesResourceReservationHelper.megaBytesValue(KubernetesResourceQuotaDefaults.MemoryLimit);
    }
  }

  showEditor() {
    this.state.showEditorTab = true;
    this.selectTab(2);
  }

  hasResourceQuotaBeenReduced() {
    if (this.formValues.HasQuota && this.oldQuota) {
      const cpuLimit = this.formValues.CpuLimit;
      const memoryLimit = KubernetesResourceReservationHelper.bytesValue(this.formValues.MemoryLimit);
      if (cpuLimit < this.oldQuota.CpuLimit || memoryLimit < this.oldQuota.MemoryLimit) {
        return true;
      }
    }
    return false;
  }

  /* #region  UPDATE NAMESPACE */
  async updateResourcePoolAsync() {
    this.state.actionInProgress = true;
    try {
      this.checkDefaults();
      await this.KubernetesResourcePoolService.patch(this.savedFormValues, this.formValues);
      this.Notifications.success('Namespace successfully updated', this.pool.Namespace.Name);
      this.$state.reload();
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to create namespace');
    } finally {
      this.state.actionInProgress = false;
    }
  }

  updateResourcePool() {
    const willBeDeleted = _.filter(this.formValues.IngressClasses, { WasSelected: true, Selected: false });
    const warnings = {
      quota: this.hasResourceQuotaBeenReduced(),
      ingress: willBeDeleted.length !== 0,
    };

    if (warnings.quota || warnings.ingress) {
      const messages = {
        quota:
          'Reducing the quota assigned to an "in-use" namespace may have unintended consequences, including preventing running applications from functioning correctly and potentially even blocking them from running at all.',
        ingress: 'Deactivating ingresses may cause applications to be unaccessible. All ingress configurations from affected applications will be removed.',
      };
      const displayedMessage = `${warnings.quota ? messages.quota : ''}${warnings.quota && warnings.ingress ? '<br/><br/>' : ''}
      ${warnings.ingress ? messages.ingress : ''}<br/><br/>Do you wish to continue?`;
      this.ModalService.confirmUpdate(displayedMessage, (confirmed) => {
        if (confirmed) {
          return this.$async(this.updateResourcePoolAsync);
        }
      });
    } else {
      return this.$async(this.updateResourcePoolAsync);
    }
  }
  /* #endregion */

  hasEventWarnings() {
    return this.state.eventWarningCount;
  }

  /* #region  GET EVENTS */
  getEvents() {
    return this.$async(async () => {
      try {
        this.state.eventsLoading = true;
        this.events = await this.KubernetesEventService.get(this.pool.Namespace.Name);
        this.state.eventWarningCount = KubernetesEventHelper.warningCount(this.events);
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to retrieve namespace related events');
      } finally {
        this.state.eventsLoading = false;
      }
    });
  }
  /* #endregion */

  /* #region  GET APPLICATIONS */
  getApplications() {
    return this.$async(async () => {
      try {
        this.state.applicationsLoading = true;
        this.applications = await this.KubernetesApplicationService.get(this.pool.Namespace.Name);
        this.applications = _.map(this.applications, (app) => {
          const resourceReservation = KubernetesResourceReservationHelper.computeResourceReservation(app.Pods);
          app.CPU = resourceReservation.CPU;
          app.Memory = resourceReservation.Memory;
          return app;
        });

        await this.getResourceUsage(this.pool.Namespace.Name);
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to retrieve applications.');
      } finally {
        this.state.applicationsLoading = false;
      }
    });
  }
  /* #endregion */

  /* #region  GET INGRESSES */
  getIngresses() {
    return this.$async(async () => {
      this.state.ingressesLoading = true;
      try {
        const namespace = this.pool.Namespace.Name;
        this.allIngresses = await this.KubernetesIngressService.get();
        this.ingresses = _.filter(this.allIngresses, { Namespace: namespace });
        _.forEach(this.ingresses, (ing) => {
          ing.Namespace = namespace;
          _.forEach(ing.Paths, (path) => {
            const application = _.find(this.applications, { ServiceName: path.ServiceName });
            path.ApplicationName = application && application.Name ? application.Name : '-';
          });
        });
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to retrieve ingresses.');
      } finally {
        this.state.ingressesLoading = false;
      }
    });
  }
  /* #endregion */

  /* #region  GET REGISTRIES */
  getRegistries() {
    return this.$async(async () => {
      try {
        const namespace = this.$state.params.id;

        if (this.isAdmin) {
          this.registries = await this.EndpointService.registries(this.endpoint.Id);
          this.registries.forEach((reg) => {
            if (reg.RegistryAccesses && reg.RegistryAccesses[this.endpoint.Id] && reg.RegistryAccesses[this.endpoint.Id].Namespaces.includes(namespace)) {
              reg.Checked = true;
              this.formValues.Registries.push(reg);
            }
          });

          return;
        }

        const registries = await this.EndpointService.registries(this.endpoint.Id, namespace);
        this.selectedRegistries = registries.map((r) => r.Name).join(', ');
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to retrieve registries');
      }
    });
  }
  /* #endregion */

  async getResourceUsage(namespace) {
    try {
      const namespaceMetrics = await this.KubernetesMetricsService.getPods(namespace);
      // extract resource usage of all containers within each pod of the namespace
      const containerResourceUsageList = namespaceMetrics.items.flatMap((i) => i.containers.map((c) => c.usage));
      const namespaceResourceUsage = containerResourceUsageList.reduce((total, u) => {
        total.CPU += KubernetesResourceReservationHelper.parseCPU(u.cpu);
        total.Memory += KubernetesResourceReservationHelper.megaBytesValue(u.memory);
        return total;
      }, new KubernetesResourceReservation());
      this.state.resourceUsage = namespaceResourceUsage;
    } catch (err) {
      this.Notifications.error('Failure', 'Unable to retrieve namespace resource usage', err);
    }
  }

  /* #region  ON INIT */
  $onInit() {
    return this.$async(async () => {
      try {
        this.isAdmin = this.Authentication.isAdmin();

        this.state = {
          actionInProgress: false,
          sliderMaxMemory: 0,
          sliderMaxCpu: 0,
          cpuUsage: 0,
          memoryUsage: 0,
          resourceReservation: { CPU: 0, Memory: 0 },
          activeTab: 0,
          currentName: this.$state.$current.name,
          showEditorTab: false,
          eventsLoading: true,
          applicationsLoading: true,
          ingressesLoading: true,
          viewReady: false,
          eventWarningCount: 0,
          canUseIngress: this.endpoint.Kubernetes.Configuration.IngressClasses.length,
          useServerMetrics: this.endpoint.Kubernetes.Configuration.UseServerMetrics,
          duplicates: {
            ingressHosts: new KubernetesFormValidationReferences(),
          },
        };

        this.state.activeTab = this.LocalStorage.getActiveTab('resourcePool');

        const name = this.$state.params.id;

        const [nodes, pools] = await Promise.all([this.KubernetesNodeService.get(), this.KubernetesResourcePoolService.get()]);

        this.pool = _.find(pools, { Namespace: { Name: name } });
        this.formValues = new KubernetesResourcePoolFormValues(KubernetesResourceQuotaDefaults);
        this.formValues.Name = this.pool.Namespace.Name;
        this.formValues.EndpointId = this.endpoint.Id;

        _.forEach(nodes, (item) => {
          this.state.sliderMaxMemory += filesizeParser(item.Memory);
          this.state.sliderMaxCpu += item.CPU;
        });
        this.state.sliderMaxMemory = KubernetesResourceReservationHelper.megaBytesValue(this.state.sliderMaxMemory);

        const quota = this.pool.Quota;
        if (quota) {
          this.oldQuota = angular.copy(quota);
          this.formValues = KubernetesResourceQuotaConverter.quotaToResourcePoolFormValues(quota);
          this.formValues.EndpointId = this.endpoint.Id;

          this.state.resourceReservation.CPU = quota.CpuLimitUsed;
          this.state.resourceReservation.Memory = KubernetesResourceReservationHelper.megaBytesValue(quota.MemoryLimitUsed);
        }

        this.isEditable = !this.KubernetesNamespaceHelper.isSystemNamespace(this.pool.Namespace.Name);
        if (this.pool.Namespace.Name === 'default') {
          this.isEditable = false;
        }

        await this.getEvents();
        await this.getApplications();

        if (this.state.canUseIngress) {
          await this.getIngresses();
          const ingressClasses = this.endpoint.Kubernetes.Configuration.IngressClasses;
          this.formValues.IngressClasses = KubernetesIngressConverter.ingressClassesToFormValues(ingressClasses, this.ingresses);
          _.forEach(this.formValues.IngressClasses, (ic) => {
            if (ic.Hosts.length === 0) {
              ic.Hosts.push(new KubernetesResourcePoolIngressClassHostFormValue());
            }
          });
        }

        await this.getRegistries();

        this.savedFormValues = angular.copy(this.formValues);
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to load view data');
      } finally {
        this.state.viewReady = true;
      }
    });
  }

  /* #endregion */

  $onDestroy() {
    if (this.state.currentName !== this.$state.$current.name) {
      this.LocalStorage.storeActiveTab('resourcePool', 0);
    }
  }
}

export default KubernetesResourcePoolController;
angular.module('portainer.kubernetes').controller('KubernetesResourcePoolController', KubernetesResourcePoolController);

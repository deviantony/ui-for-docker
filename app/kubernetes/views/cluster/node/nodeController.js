import angular from 'angular';
import _ from 'lodash-es';
import KubernetesResourceReservationHelper from 'Kubernetes/helpers/resourceReservationHelper';
import { KubernetesResourceReservation } from 'Kubernetes/models/resource-reservation/models';
import KubernetesEventHelper from 'Kubernetes/helpers/eventHelper';
import KubernetesNodeConverter from 'Kubernetes/node/converter';
import { KubernetesNodeLabelFormValues, KubernetesNodeTaintFormValues } from 'Kubernetes/node/formValues';
import { KubernetesNodeTaintEffects, KubernetesNodeAvailabilities } from 'Kubernetes/node/models';
import KubernetesFormValidationHelper from 'Kubernetes/helpers/formValidationHelper';
import { KubernetesNodeHelper } from 'Kubernetes/node/helper';

class KubernetesNodeController {
  /* @ngInject */
  constructor(
    $async,
    $state,
    Notifications,
    LocalStorage,
    ModalService,
    KubernetesNodeService,
    KubernetesEventService,
    KubernetesPodService,
    KubernetesApplicationService,
    KubernetesEndpointService,
    KubernetesMetricsService,
    Authentication
  ) {
    this.$async = $async;
    this.$state = $state;
    this.Notifications = Notifications;
    this.LocalStorage = LocalStorage;
    this.ModalService = ModalService;
    this.KubernetesNodeService = KubernetesNodeService;
    this.KubernetesEventService = KubernetesEventService;
    this.KubernetesPodService = KubernetesPodService;
    this.KubernetesApplicationService = KubernetesApplicationService;
    this.KubernetesEndpointService = KubernetesEndpointService;
    this.KubernetesMetricsService = KubernetesMetricsService;
    this.Authentication = Authentication;

    this.onInit = this.onInit.bind(this);
    this.getNodesAsync = this.getNodesAsync.bind(this);
    this.getEvents = this.getEvents.bind(this);
    this.getEventsAsync = this.getEventsAsync.bind(this);
    this.getApplicationsAsync = this.getApplicationsAsync.bind(this);
    this.getEndpointsAsync = this.getEndpointsAsync.bind(this);
    this.updateNodeAsync = this.updateNodeAsync.bind(this);
    this.drainNodeAsync = this.drainNodeAsync.bind(this);
    this.getNodeUsageAsync = this.getNodeUsageAsync.bind(this);
  }

  selectTab(index) {
    this.LocalStorage.storeActiveTab('node', index);
  }

  /* #region taint */

  onChangeTaintKey(index) {
    this.state.duplicateTaintKeys = KubernetesFormValidationHelper.getDuplicates(
      _.map(this.formValues.Taints, (taint) => {
        if (taint.NeedsDeletion) {
          return undefined;
        }
        return taint.Key;
      })
    );
    this.state.hasDuplicateTaintKeys = Object.keys(this.state.duplicateTaintKeys).length > 0;
    this.onChangeTaint(index);
  }

  onChangeTaint(index) {
    if (this.formValues.Taints[index]) {
      this.formValues.Taints[index].IsChanged = true;
    }
  }

  addTaint() {
    const taint = new KubernetesNodeTaintFormValues();
    taint.IsNew = true;
    taint.Effect = KubernetesNodeTaintEffects.NOSCHEDULE;
    this.formValues.Taints.push(taint);
  }

  removeTaint(index) {
    const taint = this.formValues.Taints[index];
    if (taint.IsNew) {
      this.formValues.Taints.splice(index, 1);
    } else {
      taint.NeedsDeletion = true;
    }
    this.onChangeTaintKey();
  }

  restoreTaint(index) {
    this.formValues.Taints[index].NeedsDeletion = false;
    this.onChangeTaintKey();
  }

  computeTaintsWarning() {
    return _.filter(this.formValues.Taints, (taint) => {
      return taint.Effect === KubernetesNodeTaintEffects.NOEXECUTE && (taint.IsNew || taint.IsChanged);
    }).length;
  }

  /* #endregion */

  /* #region label */

  onChangeLabelKey(index) {
    this.state.duplicateLabelKeys = KubernetesFormValidationHelper.getDuplicates(
      _.map(this.formValues.Labels, (label) => {
        if (label.NeedsDeletion) {
          return undefined;
        }
        return label.Key;
      })
    );
    this.state.hasDuplicateLabelKeys = Object.keys(this.state.duplicateLabelKeys).length > 0;
    this.onChangeLabel(index);
  }

  onChangeLabel(index) {
    if (this.formValues.Labels[index]) {
      this.formValues.Labels[index].IsChanged = true;
    }
  }

  addLabel() {
    const label = new KubernetesNodeLabelFormValues();
    label.IsNew = true;
    this.formValues.Labels.push(label);
  }

  removeLabel(index) {
    const label = this.formValues.Labels[index];
    if (label.IsNew) {
      this.formValues.Labels.splice(index, 1);
    } else {
      label.NeedsDeletion = true;
    }
    this.onChangeLabelKey();
  }

  restoreLabel(index) {
    this.formValues.Labels[index].NeedsDeletion = false;
    this.onChangeLabelKey();
  }

  isSystemLabel(index) {
    return KubernetesNodeHelper.isSystemLabel(this.formValues.Labels[index]);
  }

  computeLabelsWarning() {
    return _.filter(this.formValues.Labels, (label) => {
      return label.IsUsed && (label.NeedsDeletion || label.IsChanged);
    }).length;
  }

  /* #endregion */

  /* #region cordon */

  computeCordonWarning() {
    return this.formValues.Availability === this.availabilities.PAUSE;
  }

  /* #endregion */

  /* #region drain */

  computeDrainWarning() {
    return this.formValues.Availability === this.availabilities.DRAIN;
  }

  async drainNodeAsync() {
    const pods = _.flatten(_.map(this.applications, (app) => app.Pods));
    let actionCount = pods.length;
    for (const pod of pods) {
      try {
        await this.KubernetesPodService.eviction(pod);
        this.Notifications.success('Pod successfully evicted', pod.Name);
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to evict pod');
        this.formValues.Availability = this.availabilities.PAUSE;
        await this.KubernetesNodeService.patch(this.node, this.formValues);
      } finally {
        --actionCount;
        if (actionCount === 0) {
          this.formValues.Availability = this.availabilities.PAUSE;
          await this.KubernetesNodeService.patch(this.node, this.formValues);
        }
      }
    }
  }

  drainNode() {
    return this.$async(this.drainNodeAsync);
  }

  /* #endregion */

  /* #region actions */

  isNoChangesMade() {
    const newNode = KubernetesNodeConverter.formValuesToNode(this.node, this.formValues);
    const payload = KubernetesNodeConverter.patchPayload(this.node, newNode);
    return !payload.length;
  }

  isDrainError() {
    return (this.state.isDrainOperation || this.state.isContainPortainer) && this.formValues.Availability === this.availabilities.DRAIN;
  }

  isFormValid() {
    return !this.state.hasDuplicateTaintKeys && !this.state.hasDuplicateLabelKeys && !this.isNoChangesMade() && !this.isDrainError();
  }

  resetFormValues() {
    this.formValues = KubernetesNodeConverter.nodeToFormValues(this.node);
  }

  /* #endregion */

  async getEndpointsAsync() {
    try {
      const endpoints = await this.KubernetesEndpointService.get();
      this.endpoint = _.find(endpoints, { Name: 'kubernetes' });
      if (this.endpoint && this.endpoint.Subsets) {
        _.forEach(this.endpoint.Subsets, (subset) => {
          return _.forEach(subset.Ips, (ip) => {
            if (ip === this.node.IPAddress) {
              this.node.Api = true;
              this.node.Port = subset.Port;
              return false;
            }
          });
        });
      }
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve endpoints');
    }
  }

  getEndpoints() {
    return this.$async(this.getEndpointsAsync);
  }

  async updateNodeAsync() {
    try {
      this.node = await this.KubernetesNodeService.patch(this.node, this.formValues);
      if (this.formValues.Availability === 'Drain') {
        await this.drainNode();
      }
      this.Notifications.success('Node updated successfully');
      this.$state.reload();
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to update node');
    }
  }

  updateNode() {
    const taintsWarning = this.computeTaintsWarning();
    const labelsWarning = this.computeLabelsWarning();
    const cordonWarning = this.computeCordonWarning();
    const drainWarning = this.computeDrainWarning();

    if (taintsWarning && !labelsWarning) {
      this.ModalService.confirmUpdate(
        'Changes to taints will immediately deschedule applications running on this node without the corresponding tolerations. Do you wish to continue?',
        (confirmed) => {
          if (confirmed) {
            return this.$async(this.updateNodeAsync);
          }
        }
      );
    } else if (!taintsWarning && labelsWarning) {
      this.ModalService.confirmUpdate(
        'Removing or changing a label that is used might prevent applications from being scheduled on this node in the future. Do you wish to continue?',
        (confirmed) => {
          if (confirmed) {
            return this.$async(this.updateNodeAsync);
          }
        }
      );
    } else if (taintsWarning && labelsWarning) {
      this.ModalService.confirmUpdate(
        'Changes to taints will immediately deschedule applications running on this node without the corresponding tolerations.<br/></br/>Removing or changing a label that is used might prevent applications from scheduling on this node in the future.\n\nDo you wish to continue?',
        (confirmed) => {
          if (confirmed) {
            return this.$async(this.updateNodeAsync);
          }
        }
      );
    } else if (cordonWarning) {
      this.ModalService.confirmUpdate(
        'Marking this node as unschedulable will effectively cordon the node and prevent any new workload from being scheduled on that node. Are you sure?',
        (confirmed) => {
          if (confirmed) {
            return this.$async(this.updateNodeAsync);
          }
        }
      );
    } else if (drainWarning) {
      this.ModalService.confirmUpdate(
        'Draining this node will cause all workloads to be evicted from that node. This might lead to some service interruption. Are you sure?',
        (confirmed) => {
          if (confirmed) {
            return this.$async(this.updateNodeAsync);
          }
        }
      );
    } else {
      return this.$async(this.updateNodeAsync);
    }
  }

  async getNodesAsync() {
    try {
      this.state.dataLoading = true;
      const nodeName = this.$transition$.params().name;
      this.nodes = await this.KubernetesNodeService.get();
      this.node = _.find(this.nodes, { Name: nodeName });
      this.state.isDrainOperation = _.find(this.nodes, { Availability: this.availabilities.DRAIN });
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve node');
    } finally {
      this.state.dataLoading = false;
    }
  }

  getNodes() {
    return this.$async(this.getNodesAsync);
  }

  async getNodeUsageAsync() {
    try {
      const nodeName = this.$transition$.params().name;
      const node = await this.KubernetesMetricsService.getNode(nodeName);
      this.resourceUsage = new KubernetesResourceReservation();
      this.resourceUsage.CPU = KubernetesResourceReservationHelper.parseCPU(node.usage.cpu);
      this.resourceUsage.Memory = KubernetesResourceReservationHelper.megaBytesValue(node.usage.memory);
    } catch (err) {
      this.Notifications.error('Failure', 'Unable to retrieve node resource usage', err);
    }
  }

  getNodeUsage() {
    return this.$async(this.getNodeUsageAsync);
  }

  hasEventWarnings() {
    return this.state.eventWarningCount;
  }

  async getEventsAsync() {
    try {
      this.state.eventsLoading = true;
      this.events = await this.KubernetesEventService.get();
      this.events = _.filter(this.events.items, (item) => item.involvedObject.kind === 'Node');
      this.state.eventWarningCount = KubernetesEventHelper.warningCount(this.events);
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve node events');
    } finally {
      this.state.eventsLoading = false;
    }
  }

  getEvents() {
    return this.$async(this.getEventsAsync);
  }

  showEditor() {
    this.state.showEditorTab = true;
    this.selectTab(2);
  }

  async getApplicationsAsync() {
    try {
      this.state.applicationsLoading = true;
      this.applications = await this.KubernetesApplicationService.get();

      this.resourceReservation = new KubernetesResourceReservation();
      this.applications = _.map(this.applications, (app) => {
        app.Pods = _.filter(app.Pods, (pod) => pod.Node === this.node.Name);
        return app;
      });
      this.applications = _.filter(this.applications, (app) => app.Pods.length !== 0);
      this.applications = _.map(this.applications, (app) => {
        const resourceReservation = KubernetesResourceReservationHelper.computeResourceReservation(app.Pods);
        app.CPU = resourceReservation.CPU;
        app.Memory = resourceReservation.Memory;
        this.resourceReservation.CPU += resourceReservation.CPU;
        this.resourceReservation.Memory += resourceReservation.Memory;
        return app;
      });
      this.resourceReservation.Memory = KubernetesResourceReservationHelper.megaBytesValue(this.resourceReservation.Memory);
      this.memoryLimit = KubernetesResourceReservationHelper.megaBytesValue(this.node.Memory);
      this.state.isContainPortainer = _.find(this.applications, { ApplicationName: 'portainer' });

      if (this.state.isAdmin) {
        await this.getNodeUsage();
      }
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve applications');
    } finally {
      this.state.applicationsLoading = false;
    }
  }

  getApplications() {
    return this.$async(this.getApplicationsAsync);
  }

  async onInit() {
    this.state = {
      isAdmin: this.Authentication.isAdmin(),
      activeTab: 0,
      currentName: this.$state.$current.name,
      dataLoading: true,
      eventsLoading: true,
      applicationsLoading: true,
      showEditorTab: false,
      viewReady: false,
      eventWarningCount: 0,
      duplicateTaintKeys: [],
      hasDuplicateTaintKeys: false,
      duplicateLabelKeys: [],
      hasDuplicateLabelKeys: false,
      isDrainOperation: false,
      isContainPortainer: false,
      useServerMetrics: false,
    };

    this.availabilities = KubernetesNodeAvailabilities;

    this.state.useServerMetrics = this.endpoint.Kubernetes.Configuration.UseServerMetrics;

    this.state.activeTab = this.LocalStorage.getActiveTab('node');

    await this.getNodes();
    await this.getEvents();
    await this.getApplications();
    await this.getEndpoints();

    this.availableEffects = _.values(KubernetesNodeTaintEffects);
    this.formValues = KubernetesNodeConverter.nodeToFormValues(this.node);
    this.formValues.Labels = KubernetesNodeHelper.computeUsedLabels(this.applications, this.formValues.Labels);
    this.formValues.Labels = KubernetesNodeHelper.reorderLabels(this.formValues.Labels);

    this.state.viewReady = true;
  }

  $onInit() {
    return this.$async(this.onInit);
  }

  $onDestroy() {
    if (this.state.currentName !== this.$state.$current.name) {
      this.LocalStorage.storeActiveTab('node', 0);
    }
  }
}

export default KubernetesNodeController;
angular.module('portainer.kubernetes').controller('KubernetesNodeController', KubernetesNodeController);

import _ from 'lodash-es';
import angular from 'angular';
import { KubernetesStorageClass, KubernetesStorageClassAccessPolicies } from 'Kubernetes/models/storage-class/models';
import { KubernetesFormValidationReferences } from 'Kubernetes/models/application/formValues';
import { KubernetesIngressClass } from 'Kubernetes/ingress/models';
import KubernetesFormValidationHelper from 'Kubernetes/helpers/formValidationHelper';
import { KubernetesIngressClassTypes } from 'Kubernetes/ingress/constants';

class KubernetesConfigureController {
  /* #region  CONSTRUCTOR */

  /* @ngInject */
  constructor(
    $async,
    $state,
    Notifications,
    KubernetesStorageService,
    EndpointService,
    EndpointProvider,
    ModalService,
    KubernetesNamespaceHelper,
    KubernetesResourcePoolService,
    KubernetesIngressService,
    KubernetesMetricsService
  ) {
    this.$async = $async;
    this.$state = $state;
    this.Notifications = Notifications;
    this.KubernetesStorageService = KubernetesStorageService;
    this.EndpointService = EndpointService;
    this.EndpointProvider = EndpointProvider;
    this.ModalService = ModalService;
    this.KubernetesNamespaceHelper = KubernetesNamespaceHelper;
    this.KubernetesResourcePoolService = KubernetesResourcePoolService;
    this.KubernetesIngressService = KubernetesIngressService;
    this.KubernetesMetricsService = KubernetesMetricsService;

    this.IngressClassTypes = KubernetesIngressClassTypes;

    this.onInit = this.onInit.bind(this);
    this.configureAsync = this.configureAsync.bind(this);
  }
  /* #endregion */

  /* #region  STORAGE CLASSES UI MANAGEMENT */
  storageClassAvailable() {
    return this.StorageClasses && this.StorageClasses.length > 0;
  }

  hasValidStorageConfiguration() {
    let valid = true;
    _.forEach(this.StorageClasses, (item) => {
      if (item.selected && item.AccessModes.length === 0) {
        valid = false;
      }
    });
    return valid;
  }
  /* #endregion */

  /* #region  INGRESS CLASSES UI MANAGEMENT */
  addIngressClass() {
    this.formValues.IngressClasses.push(new KubernetesIngressClass());
    this.onChangeIngressClass();
  }

  restoreIngressClass(index) {
    this.formValues.IngressClasses[index].NeedsDeletion = false;
    this.onChangeIngressClass();
  }

  removeIngressClass(index) {
    if (!this.formValues.IngressClasses[index].IsNew) {
      this.formValues.IngressClasses[index].NeedsDeletion = true;
    } else {
      this.formValues.IngressClasses.splice(index, 1);
    }
    this.onChangeIngressClass();
  }

  onChangeIngressClass() {
    const state = this.state.duplicates.ingressClasses;
    const source = _.map(this.formValues.IngressClasses, (ic) => (ic.NeedsDeletion ? undefined : ic.Name));
    const duplicates = KubernetesFormValidationHelper.getDuplicates(source);
    state.refs = duplicates;
    state.hasRefs = Object.keys(duplicates).length > 0;
  }

  onChangeIngressClassName(index) {
    const fv = this.formValues.IngressClasses[index];
    if (_.includes(fv.Name, KubernetesIngressClassTypes.NGINX)) {
      fv.Type = KubernetesIngressClassTypes.NGINX;
    } else if (_.includes(fv.Name, KubernetesIngressClassTypes.TRAEFIK)) {
      fv.Type = KubernetesIngressClassTypes.TRAEFIK;
    }
    this.onChangeIngressClass();
  }

  hasTraefikIngress() {
    return _.find(this.formValues.IngressClasses, { Type: this.IngressClassTypes.TRAEFIK });
  }
  /* #endregion */

  /* #region  CONFIGURE */
  assignFormValuesToEndpoint(endpoint, storageClasses, ingressClasses) {
    endpoint.Kubernetes.Configuration.StorageClasses = storageClasses;
    endpoint.Kubernetes.Configuration.UseLoadBalancer = this.formValues.UseLoadBalancer;
    endpoint.Kubernetes.Configuration.UseServerMetrics = this.formValues.UseServerMetrics;
    endpoint.Kubernetes.Configuration.IngressClasses = ingressClasses;
    endpoint.Kubernetes.Configuration.RestrictDefaultNamespace = this.formValues.RestrictDefaultNamespace;
  }

  transformFormValues() {
    const storageClasses = _.map(this.StorageClasses, (item) => {
      if (item.selected) {
        const res = new KubernetesStorageClass();
        res.Name = item.Name;
        res.AccessModes = _.map(item.AccessModes, 'Name');
        res.Provisioner = item.Provisioner;
        res.AllowVolumeExpansion = item.AllowVolumeExpansion;
        return res;
      }
    });
    _.pull(storageClasses, undefined);

    const ingressClasses = _.without(
      _.map(this.formValues.IngressClasses, (ic) => (ic.NeedsDeletion ? undefined : ic)),
      undefined
    );
    _.pull(ingressClasses, undefined);

    return [storageClasses, ingressClasses];
  }

  async removeIngressesAcrossNamespaces() {
    const ingressesToDel = _.filter(this.formValues.IngressClasses, { NeedsDeletion: true });

    if (!ingressesToDel.length) {
      return;
    }

    const promises = [];
    const oldEndpointID = this.EndpointProvider.endpointID();
    this.EndpointProvider.setEndpointID(this.endpoint.Id);

    try {
      const allResourcePools = await this.KubernetesResourcePoolService.get();
      const resourcePools = _.filter(
        allResourcePools,
        (resourcePool) =>
          !this.KubernetesNamespaceHelper.isSystemNamespace(resourcePool.Namespace.Name) && !this.KubernetesNamespaceHelper.isDefaultNamespace(resourcePool.Namespace.Name)
      );

      ingressesToDel.forEach((ingress) => {
        resourcePools.forEach((resourcePool) => {
          promises.push(this.KubernetesIngressService.delete(resourcePool.Namespace.Name, ingress.Name));
        });
      });
    } finally {
      this.EndpointProvider.setEndpointID(oldEndpointID);
    }

    const responses = await Promise.allSettled(promises);
    responses.forEach((respons) => {
      if (respons.status == 'rejected' && respons.reason.err.status != 404) {
        throw respons.reason;
      }
    });
  }

  enableMetricsServer() {
    if (this.formValues.UseServerMetrics) {
      this.state.metrics.userClick = true;
      this.state.metrics.pending = true;
      this.KubernetesMetricsService.capabilities(this.endpoint.Id)
        .then(() => {
          this.state.metrics.isServerRunning = true;
          this.state.metrics.pending = false;
          this.formValues.UseServerMetrics = true;
        })
        .catch(() => {
          this.state.metrics.isServerRunning = false;
          this.state.metrics.pending = false;
          this.formValues.UseServerMetrics = false;
        });
    } else {
      this.state.metrics.userClick = false;
      this.formValues.UseServerMetrics = false;
    }
  }

  async configureAsync() {
    try {
      this.state.actionInProgress = true;
      const [storageClasses, ingressClasses] = this.transformFormValues();

      await this.removeIngressesAcrossNamespaces();

      this.assignFormValuesToEndpoint(this.endpoint, storageClasses, ingressClasses);
      await this.EndpointService.updateEndpoint(this.endpoint.Id, this.endpoint);

      const storagePromises = _.map(storageClasses, (storageClass) => {
        const oldStorageClass = _.find(this.oldStorageClasses, { Name: storageClass.Name });
        if (oldStorageClass) {
          return this.KubernetesStorageService.patch(this.state.endpointId, oldStorageClass, storageClass);
        }
      });
      await Promise.all(storagePromises);

      const endpoints = this.EndpointProvider.endpoints();
      const modifiedEndpoint = _.find(endpoints, (item) => item.Id === this.endpoint.Id);
      if (modifiedEndpoint) {
        this.assignFormValuesToEndpoint(modifiedEndpoint, storageClasses, ingressClasses);
        this.EndpointProvider.setEndpoints(endpoints);
      }
      this.Notifications.success('Configuration successfully applied');
      this.$state.go('portainer.home');
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to apply configuration');
    } finally {
      this.state.actionInProgress = false;
    }
  }

  configure() {
    const toDel = _.filter(this.formValues.IngressClasses, { NeedsDeletion: true });
    if (toDel.length) {
      this.ModalService.confirmUpdate(
        `Removing ingress controllers may cause applications to be unaccessible. All ingress configurations from affected applications will be removed.<br/><br/>Do you wish to continue?`,
        (confirmed) => {
          if (confirmed) {
            return this.$async(this.configureAsync);
          }
        }
      );
    } else {
      return this.$async(this.configureAsync);
    }
  }
  /* #endregion */

  /* #region  ON INIT */
  async onInit() {
    this.state = {
      actionInProgress: false,
      displayConfigureClassPanel: {},
      viewReady: false,
      endpointId: this.$state.params.id,
      duplicates: {
        ingressClasses: new KubernetesFormValidationReferences(),
      },
      metrics: {
        pending: false,
        isServerRunning: false,
        userClick: false,
      },
    };

    this.formValues = {
      UseLoadBalancer: false,
      UseServerMetrics: false,
      IngressClasses: [],
      RestrictDefaultNamespace: false,
    };

    try {
      [this.StorageClasses, this.endpoint] = await Promise.all([this.KubernetesStorageService.get(this.state.endpointId), this.EndpointService.endpoint(this.state.endpointId)]);
      _.forEach(this.StorageClasses, (item) => {
        item.availableAccessModes = new KubernetesStorageClassAccessPolicies();
        const storage = _.find(this.endpoint.Kubernetes.Configuration.StorageClasses, (sc) => sc.Name === item.Name);
        if (storage) {
          item.selected = true;
          _.forEach(storage.AccessModes, (access) => {
            const mode = _.find(item.availableAccessModes, { Name: access });
            if (mode) {
              mode.selected = true;
            }
          });
        }
      });

      this.oldStorageClasses = angular.copy(this.StorageClasses);

      this.formValues.UseLoadBalancer = this.endpoint.Kubernetes.Configuration.UseLoadBalancer;
      this.formValues.UseServerMetrics = this.endpoint.Kubernetes.Configuration.UseServerMetrics;
      this.formValues.RestrictDefaultNamespace = this.endpoint.Kubernetes.Configuration.RestrictDefaultNamespace;
      this.formValues.IngressClasses = _.map(this.endpoint.Kubernetes.Configuration.IngressClasses, (ic) => {
        ic.IsNew = false;
        ic.NeedsDeletion = false;
        return ic;
      });
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve endpoint configuration');
    } finally {
      this.state.viewReady = true;
    }
  }

  $onInit() {
    return this.$async(this.onInit);
  }
  /* #endregion */
}

export default KubernetesConfigureController;
angular.module('portainer.kubernetes').controller('KubernetesConfigureController', KubernetesConfigureController);

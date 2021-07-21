angular.module('portainer.docker').controller('KubernetesResourcePoolsDatatableController', [
  '$scope',
  '$controller',
  'Authentication',
  'KubernetesNamespaceHelper',
  'DatatableService',
  function ($scope, $controller, Authentication, KubernetesNamespaceHelper, DatatableService) {
    angular.extend(this, $controller('GenericDatatableController', { $scope: $scope }));

    var ctrl = this;

    this.settings = Object.assign(this.settings, {
      showSystem: false,
    });

    this.onSettingsShowSystemChange = function () {
      DatatableService.setDataTableSettings(this.tableKey, this.settings);
    };

    this.canManageAccess = function (item) {
      if (!this.endpoint.Kubernetes.Configuration.RestrictDefaultNamespace) {
        return item.Namespace.Name !== 'default' && !this.isSystemNamespace(item);
      } else {
        return !this.isSystemNamespace(item);
      }
    };

    this.disableRemove = function (item) {
      return KubernetesNamespaceHelper.isSystemNamespace(item.Namespace.Name) || item.Namespace.Name === 'default';
    };

    this.isSystemNamespace = function (item) {
      return KubernetesNamespaceHelper.isSystemNamespace(item.Namespace.Name);
    };

    this.isDisplayed = function (item) {
      return !ctrl.isSystemNamespace(item) || (ctrl.settings.showSystem && ctrl.isAdmin);
    };

    /**
     * Do not allow system namespaces to be selected
     */
    this.allowSelection = function (item) {
      return !this.disableRemove(item);
    };

    this.$onInit = function () {
      this.isAdmin = Authentication.isAdmin();
      this.setDefaults();
      this.prepareTableFromDataset();

      this.state.orderBy = this.orderBy;
      var storedOrder = DatatableService.getDataTableOrder(this.tableKey);
      if (storedOrder !== null) {
        this.state.reverseOrder = storedOrder.reverse;
        this.state.orderBy = storedOrder.orderBy;
      }

      var textFilter = DatatableService.getDataTableTextFilters(this.tableKey);
      if (textFilter !== null) {
        this.state.textFilter = textFilter;
        this.onTextFilterChange();
      }

      var storedFilters = DatatableService.getDataTableFilters(this.tableKey);
      if (storedFilters !== null) {
        this.filters = storedFilters;
      }
      if (this.filters && this.filters.state) {
        this.filters.state.open = false;
      }

      var storedSettings = DatatableService.getDataTableSettings(this.tableKey);
      if (storedSettings !== null) {
        this.settings = storedSettings;
        this.settings.open = false;
      }
      this.onSettingsRepeaterChange();
    };
  },
]);

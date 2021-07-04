import _ from 'lodash-es';
import './datatable.css';
import { ResourceControlOwnership as RCO } from 'Portainer/models/resourceControl/resourceControlOwnership';

function isBetween(value, a, b) {
  return (value >= a && value <= b) || (value >= b && value <= a);
}

// TODO: review - refactor to use a class that can be extended
angular.module('portainer.app').controller('GenericDatatableController', [
  '$interval',
  'PaginationService',
  'DatatableService',
  'PAGINATION_MAX_ITEMS',
  function ($interval, PaginationService, DatatableService, PAGINATION_MAX_ITEMS) {
    this.RCO = RCO;

    this.state = {
      selectAll: false,
      orderBy: this.orderBy,
      paginatedItemLimit: PAGINATION_MAX_ITEMS,
      displayTextFilter: false,
      get selectedItemCount() {
        return this.selectedItems.length || 0;
      },
      selectedItems: [],
    };

    this.settings = {
      open: false,
      repeater: {
        autoRefresh: false,
        refreshRate: '30',
      },
    };

    this.resetSelectionState = function () {
      this.state.selectAll = false;
      this.state.selectedItems = [];
      _.map(this.state.filteredDataSet, (item) => (item.Checked = false));
    };

    this.onTextFilterChange = function () {
      DatatableService.setDataTableTextFilters(this.tableKey, this.state.textFilter);
    };

    this.changeOrderBy = function changeOrderBy(orderField) {
      this.state.reverseOrder = this.state.orderBy === orderField ? !this.state.reverseOrder : false;
      this.state.orderBy = orderField;
      DatatableService.setDataTableOrder(this.tableKey, orderField, this.state.reverseOrder);
    };

    this.selectItem = function (item, event) {
      // Handle range select using shift
      if (event && event.originalEvent.shiftKey && this.state.firstClickedItem) {
        const firstItemIndex = this.state.filteredDataSet.indexOf(this.state.firstClickedItem);
        const lastItemIndex = this.state.filteredDataSet.indexOf(item);
        const itemsInRange = _.filter(this.state.filteredDataSet, (item, index) => {
          return isBetween(index, firstItemIndex, lastItemIndex);
        });
        const value = this.state.firstClickedItem.Checked;

        _.forEach(itemsInRange, (i) => {
          if (!this.allowSelection(i)) {
            return;
          }
          i.Checked = value;
        });
        this.state.firstClickedItem = item;
      } else if (event) {
        item.Checked = !item.Checked;
        this.state.firstClickedItem = item;
      }
      this.state.selectedItems = _.uniq(_.concat(this.state.selectedItems, this.state.filteredDataSet)).filter((i) => i.Checked);
      if (event && this.state.selectAll && this.state.selectedItems.length !== this.state.filteredDataSet.length) {
        this.state.selectAll = false;
      }
      this.onSelectionChanged();
    };

    this.selectAll = function () {
      this.state.firstClickedItem = null;
      for (var i = 0; i < this.state.filteredDataSet.length; i++) {
        var item = this.state.filteredDataSet[i];
        if (this.allowSelection(item) && item.Checked !== this.state.selectAll) {
          item.Checked = this.state.selectAll;
          this.selectItem(item);
        }
      }
      this.onSelectionChanged();
    };

    /**
     * Override this method to allow/deny selection
     */
    this.allowSelection = function (/*item*/) {
      return true;
    };

    /**
     * Override this method to prepare data table
     */
    this.prepareTableFromDataset = function () {
      return;
    };

    /**
     * Override this method to execute code after selection changed on datatable
     */
    this.onSelectionChanged = function () {
      return;
    };

    this.changePaginationLimit = function () {
      PaginationService.setPaginationLimit(this.tableKey, this.state.paginatedItemLimit);
    };

    this.setDefaults = function () {
      this.showTextFilter = this.showTextFilter ? this.showTextFilter : false;
      this.state.reverseOrder = this.reverseOrder ? this.reverseOrder : false;
      this.state.paginatedItemLimit = PaginationService.getPaginationLimit(this.tableKey);
    };

    /**
     * Duplicate this function when extending GenericDatatableController
     * Extending-controller's bindings are not accessible there
     * For more details see the following comments
     * https://github.com/portainer/portainer/pull/2877#issuecomment-503333425
     * https://github.com/portainer/portainer/pull/2877#issuecomment-503537249
     */
    this.$onInit = function $onInit() {
      this.$onInitGeneric();
    };

    this.$onInitGeneric = function $onInitGeneric() {
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

      var storedColumnVisibility = DatatableService.getColumnVisibilitySettings(this.tableKey);
      if (storedColumnVisibility !== null) {
        this.columnVisibility = storedColumnVisibility;
      }
    };

    /**
     * REPEATER SECTION
     */
    this.repeater = undefined;

    this.$onDestroy = function () {
      this.stopRepeater();
    };

    this.stopRepeater = function () {
      if (angular.isDefined(this.repeater)) {
        $interval.cancel(this.repeater);
        this.repeater = undefined;
      }
    };

    this.startRepeater = function () {
      this.repeater = $interval(async () => {
        await this.refreshCallback();
        this.onDataRefresh();
      }, this.settings.repeater.refreshRate * 1000);
    };

    this.onSettingsRepeaterChange = function () {
      this.stopRepeater();
      if (angular.isDefined(this.refreshCallback) && this.settings.repeater.autoRefresh) {
        this.startRepeater();
        $('#refreshRateChange').show();
        $('#refreshRateChange').fadeOut(1500);
      }
      DatatableService.setDataTableSettings(this.tableKey, this.settings);
    };

    /**
     * Override this method to execute code after calling the refresh callback
     */
    this.onDataRefresh = function () {
      return;
    };

    /**
     * !REPEATER SECTION
     */
  },
]);

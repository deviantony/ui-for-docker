import angular from 'angular';
import _ from 'lodash-es';

class AssoicatedEndpointsSelectorController {
  /* @ngInject */
  constructor($async, EndpointService) {
    this.$async = $async;
    this.EndpointService = EndpointService;

    this.state = {
      available: {
        limit: '10',
        filter: '',
        pageNumber: 1,
        totalCount: 0,
      },
      associated: {
        limit: '10',
        filter: '',
        pageNumber: 1,
        totalCount: 0,
      },
    };

    this.endpoints = {
      associated: [],
      available: null,
    };

    this.getEndpoints = this.getEndpoints.bind(this);
    this.getEndpointsAsync = this.getEndpointsAsync.bind(this);
    this.getAssociatedEndpoints = this.getAssociatedEndpoints.bind(this);
    this.getAssociatedEndpointsAsync = this.getAssociatedEndpointsAsync.bind(this);
    this.associateEndpoint = this.associateEndpoint.bind(this);
    this.dissociateEndpoint = this.dissociateEndpoint.bind(this);
  }

  $onInit() {
    this.loadData();
  }

  $onChanges({ endpointIds }) {
    if (endpointIds && endpointIds.currentValue) {
      this.loadData();
    }
  }

  loadData() {
    this.getAssociatedEndpoints();
    this.getEndpoints();
  }

  getEndpoints() {
    return this.$async(this.getEndpointsAsync);
  }

  async getEndpointsAsync() {
    const { start, search, limit } = this.getPaginationData('available');
    const query = { search, types: [4, 7] };

    const response = await this.EndpointService.endpoints(start, limit, query);

    const endpoints = _.filter(response.value, (endpoint) => !_.includes(this.endpointIds, endpoint.Id));
    this.setTableData('available', endpoints, response.totalCount);
    this.noEndpoints = this.state.available.totalCount === 0;
  }

  getAssociatedEndpoints() {
    return this.$async(this.getAssociatedEndpointsAsync);
  }

  async getAssociatedEndpointsAsync() {
    let response = { value: [], totalCount: 0 };
    if (this.endpointIds.length > 0) {
      const { start, search, limit } = this.getPaginationData('associated');
      const query = { search, types: [4, 7], endpointIds: this.endpointIds };

      response = await this.EndpointService.endpoints(start, limit, query);
    }

    this.setTableData('associated', response.value, response.totalCount);
  }

  associateEndpoint(endpoint) {
    this.onAssociate(endpoint);
  }

  dissociateEndpoint(endpoint) {
    this.onDissociate(endpoint);
  }

  getPaginationData(tableType) {
    const { pageNumber, limit, search } = this.state[tableType];
    const start = (pageNumber - 1) * limit + 1;

    return { start, search, limit };
  }

  setTableData(tableType, endpoints, totalCount) {
    this.endpoints[tableType] = endpoints;
    this.state[tableType].totalCount = parseInt(totalCount, 10);
  }
}

angular.module('portainer.app').controller('AssoicatedEndpointsSelectorController', AssoicatedEndpointsSelectorController);
export default AssoicatedEndpointsSelectorController;

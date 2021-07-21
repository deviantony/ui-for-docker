import _ from 'lodash-es';

export class ContainerInstanceProviderViewModel {
  constructor(data) {
    this.Id = data.id;
    this.Namespace = data.namespace;

    const containerGroupType = _.find(data.resourceTypes, { resourceType: 'containerGroups' });
    this.Locations = containerGroupType.locations;
  }
}

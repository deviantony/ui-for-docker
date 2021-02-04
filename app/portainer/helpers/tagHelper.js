import _ from 'lodash';

export default class PortainerEndpointTagHelper {
  static idsToTagNames(tags, ids) {
    const filteredTags = _.filter(tags, (tag) => _.includes(ids, tag.Id));
    const tagNames = _.map(filteredTags, 'Name');
    return tagNames;
  }
}

export function getEndpointTags(endpoint, endpointGroup, tags = []) {
  const union = _.union(endpoint.TagIds, endpointGroup.TagIds);
  const endpointTags = _.map(union, (id) => tags.find((t) => t.Id === id));
  return _.compact(endpointTags);
}

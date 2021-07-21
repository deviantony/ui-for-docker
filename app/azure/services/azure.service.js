/* @ngInject */
export function AzureService(SubscriptionService, ResourceGroupService, ContainerGroupService, ProviderService) {
  return { deleteContainerGroup, createContainerGroup, subscriptions, containerInstanceProvider, resourceGroups, containerGroups, aggregate };

  function deleteContainerGroup(id) {
    return ContainerGroupService.delete(id);
  }

  function createContainerGroup(model, subscriptionId, resourceGroupName) {
    return ContainerGroupService.create(model, subscriptionId, resourceGroupName);
  }

  function subscriptions() {
    return SubscriptionService.subscriptions();
  }

  function containerInstanceProvider(subscriptions) {
    return retrieveResourcesForEachSubscription(subscriptions, ProviderService.containerInstanceProvider);
  }

  function resourceGroups(subscriptions) {
    return retrieveResourcesForEachSubscription(subscriptions, ResourceGroupService.resourceGroups);
  }

  function containerGroups(subscriptions) {
    return retrieveResourcesForEachSubscription(subscriptions, ContainerGroupService.containerGroups);
  }

  function aggregate(resourcesBySubscription) {
    var aggregatedResources = [];
    Object.keys(resourcesBySubscription).forEach(function (key) {
      aggregatedResources = aggregatedResources.concat(resourcesBySubscription[key]);
    });
    return aggregatedResources;
  }

  async function retrieveResourcesForEachSubscription(subscriptions, resourceQuery) {
    const resources = {};

    const resourceQueries = [];
    for (let i = 0; i < subscriptions.length; i++) {
      const subscription = subscriptions[i];
      resourceQueries.push(resourceQuery(subscription.Id));
    }

    try {
      const queriesResults = await Promise.all(resourceQueries);
      for (let i = 0; i < queriesResults.length; i++) {
        const result = queriesResults[i];
        resources[subscriptions[i].Id] = result;
      }
      return resources;
    } catch (err) {
      throw { msg: 'Unable to retrieve resources', err };
    }
  }
}

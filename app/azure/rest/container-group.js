/* @ngInject */
export function ContainerGroup($http, $resource, API_ENDPOINT_ENDPOINTS, EndpointProvider) {
  const base = $resource(
    `${API_ENDPOINT_ENDPOINTS}/:endpointId/azure/subscriptions/:subscriptionId/providers/Microsoft.ContainerInstance/containerGroups`,
    {
      endpointId: EndpointProvider.endpointID,
      'api-version': '2018-04-01',
    },
    {
      query: { method: 'GET', params: { subscriptionId: '@subscriptionId' } },
    }
  );

  const withResourceGroup = $resource(
    `${API_ENDPOINT_ENDPOINTS}/:endpointId/azure/subscriptions/:subscriptionId/resourceGroups/:resourceGroupName/providers/Microsoft.ContainerInstance/containerGroups/:containerGroupName`,
    {
      endpointId: EndpointProvider.endpointID,
      'api-version': '2018-04-01',
    },
    {
      create: {
        method: 'PUT',
        params: {
          subscriptionId: '@subscriptionId',
          resourceGroupName: '@resourceGroupName',
          containerGroupName: '@containerGroupName',
        },
      },
      get: {
        method: 'GET',
      },
    }
  );

  function deleteItem(id, apiVersion = '2018-04-01') {
    // id is of structure: "/subscriptions/b4429e9e-27fe-4347-a5ae-ce56b25d0367/resourceGroups/portainer-demo/providers/Microsoft.ContainerInstance/containerGroups/test"
    const endpointId = EndpointProvider.endpointID();
    const url = `${API_ENDPOINT_ENDPOINTS}/${endpointId}/azure${id}?api-version=${apiVersion}`;
    return $http({
      method: 'DELETE',
      url,
    });
  }

  return { query: base.query, create: withResourceGroup.create, get: withResourceGroup.get, delete: deleteItem };
}

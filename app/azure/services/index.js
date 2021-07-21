import angular from 'angular';

import { AzureService } from './azure.service';
import { ContainerGroupService } from './container-group.service';
import { ProviderService } from './provider.service';
import { ResourceGroupService } from './resource-group.service';
import { SubscriptionService } from './subscription.service';

export default angular
  .module('portainer.azure.services', [])
  .factory('AzureService', AzureService)
  .factory('ContainerGroupService', ContainerGroupService)
  .factory('ProviderService', ProviderService)
  .factory('ResourceGroupService', ResourceGroupService)
  .factory('SubscriptionService', SubscriptionService).name;

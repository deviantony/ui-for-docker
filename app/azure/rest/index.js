import angular from 'angular';

import { ContainerGroup } from './container-group';
import { Provider } from './provider';
import { ResourceGroup } from './resource-group';
import { Subscription } from './subscription';

export default angular
  .module('portainer.azure.rest', [])
  .service('ContainerGroup', ContainerGroup)
  .service('Provider', Provider)
  .service('ResourceGroup', ResourceGroup)
  .service('Subscription', Subscription).name;

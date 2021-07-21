import angular from 'angular';

import { azureEndpointConfig } from './azure-endpoint-config';
import { azureSidebarContent } from './azure-sidebar-content';

export default angular.module('portainer.azure.components', []).component('azureEndpointConfig', azureEndpointConfig).component('azureSidebarContent', azureSidebarContent).name;

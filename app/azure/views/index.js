import angular from 'angular';

import { azureDashboardView } from './azure-dashboard.view';
import { containerInstanceDetailsView } from './container-instance-details.view';
import containerInstancesModule from './container-instances.view';
import { createContainerInstanceView } from './create-container-instance.view';

export default angular
  .module('portainer.azure.views', [containerInstancesModule])
  .component('azureDashboardView', azureDashboardView)
  .component('containerInstanceDetailsView', containerInstanceDetailsView)
  .component('createContainerInstanceView', createContainerInstanceView).name;

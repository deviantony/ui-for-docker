import angular from 'angular';

import { containerInstancesView } from './container-instances.view';
import { containergroupsDatatable } from './containergroups-datatable';

export default angular
  .module('portainer.azure.views.container-instances-view', [])
  .component('containerInstancesView', containerInstancesView)
  .component('containergroupsDatatable', containergroupsDatatable).name;

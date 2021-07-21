import angular from 'angular';

import gitFormModule from './forms/git-form';
import porAccessManagementModule from './accessManagement';
import formComponentsModule from './form-components';

export default angular.module('portainer.app.components', [gitFormModule, porAccessManagementModule, formComponentsModule]).name;

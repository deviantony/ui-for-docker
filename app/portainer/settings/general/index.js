import angular from 'angular';

import { sslCertificate } from './ssl-certificate';

export default angular.module('portainer.settings.general', []).component('sslCertificateSettings', sslCertificate).name;

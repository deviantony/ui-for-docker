angular.module('portainer.app').factory('WebhookHelper', [
  '$location',
  'API_ENDPOINT_WEBHOOKS',
  'API_ENDPOINT_STACKS',
  function WebhookHelperFactory($location, API_ENDPOINT_WEBHOOKS, API_ENDPOINT_STACKS) {
    'use strict';

    var helper = {};
    const protocol = $location.protocol().toLowerCase();
    const port = $location.port();
    const displayPort = (protocol === 'http' && port === 80) || (protocol === 'https' && port === 443) ? '' : ':' + port;

    helper.returnWebhookUrl = function (token) {
      return `${protocol}://${$location.host()}${displayPort}/${API_ENDPOINT_WEBHOOKS}/${token}`;
    };

    helper.returnStackWebhookUrl = function (token) {
      return `${protocol}://${$location.host()}${displayPort}/${API_ENDPOINT_STACKS}/webhooks/${token}`;
    };

    return helper;
  },
]);

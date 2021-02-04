import _ from 'lodash-es';

angular.module('portainer.app').factory('EndpointHelper', [
  function EndpointHelperFactory() {
    'use strict';
    var helper = {};

    function findAssociatedGroup(endpoint, groups) {
      return _.find(groups, function (group) {
        return group.Id === endpoint.GroupId;
      });
    }

    helper.mapGroupToEndpoint = function (endpoints, groups) {
      for (var i = 0; i < endpoints.length; i++) {
        var endpoint = endpoints[i];
        endpoint.Group = findAssociatedGroup(endpoint, groups);
      }
    };

    return helper;
  },
]);

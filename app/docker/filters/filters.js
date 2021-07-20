import _ from 'lodash-es';

function includeString(text, values) {
  return values.some(function (val) {
    return text.indexOf(val) !== -1;
  });
}

function strToHash(str) {
  var hash = 0;
  for (var i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash);
  }
  return hash;
}

function hashToHexColor(hash) {
  var color = '#';
  for (var i = 0; i < 3; ) {
    color += ('00' + ((hash >> (i++ * 8)) & 0xff).toString(16)).slice(-2);
  }
  return color;
}

angular
  .module('portainer.docker')
  .filter('visualizerTask', function () {
    'use strict';
    return function (text) {
      var status = _.toLower(text);
      if (includeString(status, ['new', 'allocated', 'assigned', 'accepted', 'complete', 'preparing'])) {
        return 'info';
      } else if (includeString(status, ['pending'])) {
        return 'warning';
      } else if (includeString(status, ['shutdown', 'failed', 'rejected'])) {
        return 'stopped';
      }
      return 'running';
    };
  })
  .filter('visualizerTaskBorderColor', function () {
    'use strict';
    return function (str) {
      var hash = strToHash(str);
      var color = hashToHexColor(hash);
      return color;
    };
  })
  .filter('taskstatusbadge', function () {
    'use strict';
    return function (text) {
      var status = _.toLower(text);
      var labelStyle = 'default';
      if (includeString(status, ['new', 'allocated', 'assigned', 'accepted', 'preparing', 'ready', 'starting', 'remove'])) {
        labelStyle = 'info';
      } else if (includeString(status, ['pending'])) {
        labelStyle = 'warning';
      } else if (includeString(status, ['shutdown', 'failed', 'rejected', 'orphaned'])) {
        labelStyle = 'danger';
      } else if (includeString(status, ['complete'])) {
        labelStyle = 'primary';
      } else if (includeString(status, ['running'])) {
        labelStyle = 'success';
      }
      return labelStyle;
    };
  })
  .filter('taskhaslogs', function () {
    'use strict';
    return function (state) {
      var validState = ['running', 'complete', 'failed', 'shutdown'];
      if (validState.indexOf(state) > -1) {
        return true;
      }
      return false;
    };
  })
  .filter('containerstatusbadge', function () {
    'use strict';
    return function (text) {
      var status = _.toLower(text);
      if (includeString(status, ['paused', 'starting', 'unhealthy'])) {
        return 'warning';
      } else if (includeString(status, ['created'])) {
        return 'info';
      } else if (includeString(status, ['stopped', 'dead', 'exited'])) {
        return 'danger';
      }
      return 'success';
    };
  })
  .filter('nodestatusbadge', function () {
    'use strict';
    return function (text) {
      if (text === 'down' || text === 'Unhealthy') {
        return 'danger';
      }
      return 'success';
    };
  })
  .filter('dockerNodeAvailabilityBadge', function () {
    'use strict';
    return function (text) {
      if (text === 'pause') {
        return 'warning';
      } else if (text === 'drain') {
        return 'danger';
      }
      return 'success';
    };
  })
  .filter('trimcontainername', function () {
    'use strict';
    return function (name) {
      if (name) {
        return name.indexOf('/') === 0 ? name.replace('/', '') : name;
      }
      return '';
    };
  })
  .filter('getstatetext', function () {
    'use strict';
    return function (state) {
      if (state === undefined) {
        return '';
      }
      if (state.Dead) {
        return 'Dead';
      }
      if (state.Ghost && state.Running) {
        return 'Ghost';
      }
      if (state.Running && state.Paused) {
        return 'Running (Paused)';
      }
      if (state.Running) {
        return 'Running';
      }
      if (state.Status === 'created') {
        return 'Created';
      }
      return 'Stopped';
    };
  })
  .filter('getstatelabel', function () {
    'use strict';
    return function (state) {
      if (state === undefined) {
        return 'label-default';
      }
      if (state.Ghost && state.Running) {
        return 'label-important';
      }
      if (state.Running) {
        return 'label-success';
      }
      return 'label-default';
    };
  })
  .filter('containername', function () {
    'use strict';
    return function (container) {
      var name = container.Names[0];
      return name.substring(1, name.length);
    };
  })
  .filter('swarmversion', function () {
    'use strict';
    return function (text) {
      return _.split(text, '/')[1];
    };
  })
  .filter('swarmhostname', function () {
    'use strict';
    return function (container) {
      return _.split(container.Names[0], '/')[1];
    };
  })
  .filter('repotags', function () {
    'use strict';
    return function (image) {
      if (image.RepoTags && image.RepoTags.length > 0) {
        var tag = image.RepoTags[0];
        if (tag === '<none>:<none>') {
          return [];
        }
        return image.RepoTags;
      }
      return [];
    };
  })
  .filter('command', function () {
    'use strict';
    return function (command) {
      if (command) {
        return command.join(' ');
      }
    };
  })
  .filter('hideshasum', function () {
    'use strict';
    return function (imageName) {
      if (imageName) {
        return imageName.split('@sha')[0];
      }
      return '';
    };
  })
  .filter('availablenodecount', [
    'ConstraintsHelper',
    function (ConstraintsHelper) {
      'use strict';
      return function (nodes, service) {
        var availableNodes = 0;
        for (var i = 0; i < nodes.length; i++) {
          var node = nodes[i];
          if (node.Availability === 'active' && node.Status === 'ready' && ConstraintsHelper.matchesServiceConstraints(service, node)) {
            availableNodes++;
          }
        }
        return availableNodes;
      };
    },
  ])
  .filter('runningtaskscount', function () {
    'use strict';
    return function (tasks) {
      var runningTasks = 0;
      for (var i = 0; i < tasks.length; i++) {
        var task = tasks[i];
        if (task.Status.State === 'running' && task.DesiredState === 'running') {
          runningTasks++;
        }
      }
      return runningTasks;
    };
  })
  .filter('runningcontainers', function () {
    'use strict';
    return function runningContainersFilter(containers) {
      return containers.filter(function (container) {
        return container.State === 'running';
      }).length;
    };
  })
  .filter('stoppedcontainers', function () {
    'use strict';
    return function stoppedContainersFilter(containers) {
      return containers.filter(function (container) {
        return container.State === 'exited';
      }).length;
    };
  })
  .filter('healthycontainers', function () {
    'use strict';
    return function healthyContainersFilter(containers) {
      return containers.filter(function (container) {
        return container.Status === 'healthy';
      }).length;
    };
  })
  .filter('unhealthycontainers', function () {
    'use strict';
    return function unhealthyContainersFilter(containers) {
      return containers.filter(function (container) {
        return container.Status === 'unhealthy';
      }).length;
    };
  })
  .filter('imagestotalsize', function () {
    'use strict';
    return function (images) {
      var totalImageSize = 0;
      for (var i = 0; i < images.length; i++) {
        var item = images[i];
        totalImageSize += item.VirtualSize;
      }
      return totalImageSize;
    };
  })
  .filter('tasknodename', function () {
    'use strict';
    return function (nodeId, nodes) {
      var node = _.find(nodes, { Id: nodeId });
      if (node) {
        return node.Hostname;
      }
      return '';
    };
  })
  .filter('imagelayercommand', function () {
    'use strict';
    return function (createdBy) {
      return createdBy.replace('/bin/sh -c #(nop) ', '').replace('/bin/sh -c ', 'RUN ');
    };
  })
  .filter('trimshasum', function () {
    'use strict';
    return function (imageName) {
      if (!imageName) {
        return;
      }
      if (imageName.indexOf('sha256:') === 0) {
        return imageName.substring(7, 19);
      }
      return _.split(imageName, '@sha256')[0];
    };
  })
  .filter('trimversiontag', function () {
    'use strict';
    return function trimversiontag(fullName) {
      if (!fullName) {
        return fullName;
      }
      var versionIdx = fullName.lastIndexOf(':');
      if (versionIdx < 0) {
        return fullName;
      }
      var hostIdx = fullName.indexOf('/');
      if (hostIdx > versionIdx) {
        return fullName;
      }
      return fullName.substring(0, versionIdx);
    };
  })
  .filter('unique', function () {
    return _.uniqBy;
  });

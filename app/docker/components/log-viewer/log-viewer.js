angular.module('portainer.docker').component('logViewer', {
  templateUrl: './logViewer.html',
  controller: 'LogViewerController',
  bindings: {
    data: '=',
    displayTimestamps: '=',
    containerRunning: '=',
    logCollectionChange: '<',
    sinceTimestamp: '=',
    lineCount: '=',
    resourceName: '<',
  },
});

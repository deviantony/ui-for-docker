import moment from 'moment';
import _ from 'lodash-es';

angular.module('portainer.docker').controller('LogViewerController', [
  'clipboard',
  'Blob',
  'FileSaver',
  function (clipboard, Blob, FileSaver) {
    this.state = {
      availableSinceDatetime: [
        { desc: 'Last day', value: moment().subtract(1, 'days').format() },
        { desc: 'Last 4 hours', value: moment().subtract(4, 'hours').format() },
        { desc: 'Last hour', value: moment().subtract(1, 'hours').format() },
        { desc: 'Last 10 minutes', value: moment().subtract(10, 'minutes').format() },
      ],
      copySupported: clipboard.supported,
      logCollection: true,
      autoScroll: true,
      wrapLines: true,
      search: '',
      filteredLogs: [],
      selectedLines: [],
    };

    this.copy = function () {
      clipboard.copyText(this.state.filteredLogs.join(''));
      $('#refreshRateChange').show();
      $('#refreshRateChange').fadeOut(2000);
    };

    this.copySelection = function () {
      clipboard.copyText(this.state.selectedLines.join(''));
      $('#refreshRateChange').show();
      $('#refreshRateChange').fadeOut(2000);
    };

    this.clearSelection = function () {
      this.state.selectedLines = [];
    };

    this.selectLine = function (line) {
      var idx = this.state.selectedLines.indexOf(line);
      if (idx === -1) {
        this.state.selectedLines.push(line);
      } else {
        this.state.selectedLines.splice(idx, 1);
      }
    };

    this.downloadLogs = function () {
      const data = new Blob([_.reduce(this.state.filteredLogs, (acc, log) => acc + '\n' + log, '')]);
      FileSaver.saveAs(data, this.resourceName + '_logs.txt');
    };
  },
]);

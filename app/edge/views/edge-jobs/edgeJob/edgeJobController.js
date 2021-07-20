import _ from 'lodash-es';

export class EdgeJobController {
  /* @ngInject */
  constructor($async, $q, $state, $window, ModalService, EdgeJobService, EndpointService, FileSaver, GroupService, HostBrowserService, Notifications, TagService) {
    this.state = {
      actionInProgress: false,
      showEditorTab: false,
      isEditorDirty: false,
    };

    this.$async = $async;
    this.$q = $q;
    this.$state = $state;
    this.$window = $window;
    this.ModalService = ModalService;
    this.EdgeJobService = EdgeJobService;
    this.EndpointService = EndpointService;
    this.FileSaver = FileSaver;
    this.GroupService = GroupService;
    this.HostBrowserService = HostBrowserService;
    this.Notifications = Notifications;
    this.TagService = TagService;

    this.update = this.update.bind(this);
    this.updateAsync = this.updateAsync.bind(this);
    this.downloadLogs = this.downloadLogs.bind(this);
    this.downloadLogsAsync = this.downloadLogsAsync.bind(this);
    this.collectLogs = this.collectLogs.bind(this);
    this.collectLogsAsync = this.collectLogsAsync.bind(this);
    this.clearLogs = this.clearLogs.bind(this);
    this.clearLogsAsync = this.clearLogsAsync.bind(this);
    this.refresh = this.refresh.bind(this);
    this.refreshAsync = this.refreshAsync.bind(this);
    this.showEditor = this.showEditor.bind(this);
  }

  update() {
    return this.$async(this.updateAsync);
  }

  async updateAsync() {
    const model = this.edgeJob;
    this.state.actionInProgress = true;

    try {
      await this.EdgeJobService.updateEdgeJob(model);
      this.Notifications.success('Edge job successfully updated');
      this.state.isEditorDirty = false;
      this.$state.go('edge.jobs', {}, { reload: true });
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to update Edge job');
    }

    this.state.actionInProgress = false;
  }

  downloadLogs(endpointId) {
    return this.$async(this.downloadLogsAsync, endpointId);
  }
  async downloadLogsAsync(endpointId) {
    try {
      const data = await this.EdgeJobService.logFile(this.edgeJob.Id, endpointId);
      const downloadData = new Blob([data.FileContent], {
        type: 'text/plain;charset=utf-8',
      });
      const logFileName = `job_${this.edgeJob.Id}_task_${endpointId}.log`;
      this.FileSaver.saveAs(downloadData, logFileName);
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to download file');
    }
  }

  associateEndpointsToResults(results, endpoints) {
    return _.map(results, (result) => {
      const endpoint = _.find(endpoints, (endpoint) => endpoint.Id === result.EndpointId);
      result.Endpoint = endpoint;
      return result;
    });
  }

  collectLogs(endpointId) {
    return this.$async(this.collectLogsAsync, endpointId);
  }

  async collectLogsAsync(endpointId) {
    try {
      await this.EdgeJobService.collectLogs(this.edgeJob.Id, endpointId);
      const result = _.find(this.results, (result) => result.EndpointId === endpointId);
      result.LogsStatus = 2;
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to collect logs');
    }
  }

  clearLogs(endpointId) {
    return this.$async(this.clearLogsAsync, endpointId);
  }
  async clearLogsAsync(endpointId) {
    try {
      await this.EdgeJobService.clearLogs(this.edgeJob.Id, endpointId);
      const result = _.find(this.results, (result) => result.EndpointId === endpointId);
      result.LogsStatus = 1;
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to clear logs');
    }
  }

  refresh() {
    return this.$async(this.refreshAsync);
  }
  async refreshAsync() {
    const { id } = this.$state.params;
    const results = await this.EdgeJobService.jobResults(id);
    if (results.length > 0) {
      const endpointIds = _.map(results, (result) => result.EndpointId);
      const endpoints = await this.EndpointService.endpoints(undefined, undefined, { endpointIds });
      this.results = this.associateEndpointsToResults(results, endpoints.value);
    } else {
      this.results = results;
    }
  }

  showEditor() {
    this.state.showEditorTab = true;
  }

  async uiCanExit() {
    if (this.edgeJob.FileContent !== this.oldFileContent && this.state.isEditorDirty) {
      return this.ModalService.confirmWebEditorDiscard();
    }
  }

  async $onInit() {
    const { id, tab } = this.$state.params;
    this.state.activeTab = tab;
    if (!tab || tab === 0) {
      this.state.showEditorTab = true;
    }

    try {
      const [edgeJob, file, results, groups, tags] = await Promise.all([
        this.EdgeJobService.edgeJob(id),
        this.EdgeJobService.getScriptFile(id),
        this.EdgeJobService.jobResults(id),
        this.GroupService.groups(),
        this.TagService.tags(),
      ]);

      edgeJob.FileContent = file.FileContent;
      this.oldFileContent = edgeJob.FileContent;
      this.edgeJob = edgeJob;
      this.groups = groups;
      this.tags = tags;

      if (results.length > 0) {
        const endpointIds = _.map(results, (result) => result.EndpointId);
        const endpoints = await this.EndpointService.endpoints(undefined, undefined, { endpointIds });
        this.results = this.associateEndpointsToResults(results, endpoints.value);
      } else {
        this.results = results;
      }
    } catch (err) {
      this.Notifications.error('Failure', err, 'Unable to retrieve endpoint list');
    }

    this.$window.onbeforeunload = () => {
      if (this.edgeJob.FileContent !== this.oldFileContent && this.state.isEditorDirty) {
        return '';
      }
    };
  }

  $onDestroy() {
    this.state.isEditorDirty = false;
  }
}

export function ImageViewModel(data) {
  this.Id = data.Id;
  this.Tag = data.Tag;
  this.Repository = data.Repository;
  this.Created = data.Created;
  this.Checked = false;
  this.RepoTags = data.RepoTags;
  if (!this.RepoTags && data.RepoDigests) {
    this.RepoTags = [];
    for (var i = 0; i < data.RepoDigests.length; i++) {
      var digest = data.RepoDigests[i];
      var repository = digest.substring(0, digest.indexOf('@'));
      this.RepoTags.push(repository + ':<none>');
    }
  }

  this.VirtualSize = data.VirtualSize;
  this.ContainerCount = data.ContainerCount;

  if (data.Portainer && data.Portainer.Agent) {
    this.NodeName = data.Portainer.Agent.NodeName || '';
  }
  this.Labels = data.Labels;
}

export function ImageBuildModel(data) {
  this.hasError = false;
  var buildLogs = [];

  for (var i = 0; i < data.length; i++) {
    var line = data[i];

    if (line.stream) {
      line = line.stream.replace(/[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]/g, '');
      buildLogs.push(line);
    }

    if (line.errorDetail) {
      buildLogs.push(line.errorDetail.message);
      this.hasError = true;
    }
  }

  this.buildLogs = buildLogs;
}

import { ResourceControlViewModel } from 'Portainer/models/resourceControl/resourceControl';

export function VolumeViewModel(data) {
  this.Id = data.Name;
  this.CreatedAt = data.CreatedAt;
  this.Driver = data.Driver;
  this.Options = data.Options;
  this.Labels = data.Labels;
  if (this.Labels && this.Labels['com.docker.compose.project']) {
    this.StackName = this.Labels['com.docker.compose.project'];
  } else if (this.Labels && this.Labels['com.docker.stack.namespace']) {
    this.StackName = this.Labels['com.docker.stack.namespace'];
  }
  this.Mountpoint = data.Mountpoint;

  this.ResourceId = data.ResourceID;

  if (data.Portainer) {
    if (data.Portainer.ResourceControl) {
      this.ResourceControl = new ResourceControlViewModel(data.Portainer.ResourceControl);
    }
    if (data.Portainer.Agent) {
      this.NodeName = data.Portainer.Agent.NodeName || '';
    }
  }
}

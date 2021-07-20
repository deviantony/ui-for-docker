import { ResourceControlViewModel } from 'Portainer/models/resourceControl/resourceControl';

export function NetworkViewModel(data) {
  this.Id = data.Id;
  this.Name = data.Name;
  this.Scope = data.Scope;
  this.Driver = data.Driver;
  this.Attachable = data.Attachable;
  this.Internal = data.Internal;
  this.IPAM = data.IPAM;
  this.Containers = data.Containers;
  this.Options = data.Options;
  this.Ingress = data.Ingress;

  this.Labels = data.Labels;
  if (this.Labels && this.Labels['com.docker.compose.project']) {
    this.StackName = this.Labels['com.docker.compose.project'];
  } else if (this.Labels && this.Labels['com.docker.stack.namespace']) {
    this.StackName = this.Labels['com.docker.stack.namespace'];
  }

  if (data.Portainer) {
    if (data.Portainer.ResourceControl) {
      this.ResourceControl = new ResourceControlViewModel(data.Portainer.ResourceControl);
    }
    if (data.Portainer.Agent) {
      this.NodeName = data.Portainer.Agent.NodeName || '';
    }
  }

  this.ConfigFrom = data.ConfigFrom;
  this.ConfigOnly = data.ConfigOnly;
}

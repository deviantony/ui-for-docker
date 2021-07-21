export class EditEdgeStackFormController {
  /* @ngInject */
  constructor() {
    this.state = {
      endpointTypes: [],
    };

    this.fileContents = {
      0: '',
      1: '',
    };

    this.onChangeGroups = this.onChangeGroups.bind(this);
    this.onChangeFileContent = this.onChangeFileContent.bind(this);
    this.onChangeComposeConfig = this.onChangeComposeConfig.bind(this);
    this.onChangeKubeManifest = this.onChangeKubeManifest.bind(this);
    this.hasDockerEndpoint = this.hasDockerEndpoint.bind(this);
    this.hasKubeEndpoint = this.hasKubeEndpoint.bind(this);
    this.onChangeDeploymentType = this.onChangeDeploymentType.bind(this);
    this.removeLineBreaks = this.removeLineBreaks.bind(this);
    this.onChangeFileContent = this.onChangeFileContent.bind(this);
  }

  hasKubeEndpoint() {
    return this.state.endpointTypes.includes(7);
  }

  hasDockerEndpoint() {
    return this.state.endpointTypes.includes(4);
  }

  onChangeGroups(groups) {
    this.model.EdgeGroups = groups;

    this.checkEndpointTypes(groups);
  }

  isFormValid() {
    return this.model.EdgeGroups.length && this.model.StackFileContent && this.validateEndpointsForDeployment();
  }

  checkEndpointTypes(groups) {
    const edgeGroups = groups.map((id) => this.edgeGroups.find((e) => e.Id === id));
    this.state.endpointTypes = edgeGroups.flatMap((group) => group.EndpointTypes);
  }

  removeLineBreaks(value) {
    return value.replace(/(\r\n|\n|\r)/gm, '');
  }

  onChangeFileContent(type, value) {
    const oldValue = this.fileContents[type];
    if (this.removeLineBreaks(oldValue) !== this.removeLineBreaks(value)) {
      this.model.StackFileContent = value;
      this.fileContents[type] = value;
      this.isEditorDirty = true;
    }
  }

  onChangeKubeManifest(value) {
    this.onChangeFileContent(1, value);
  }

  onChangeComposeConfig(value) {
    this.onChangeFileContent(0, value);
  }

  onChangeDeploymentType(deploymentType) {
    this.model.DeploymentType = deploymentType;

    this.model.StackFileContent = this.fileContents[deploymentType];
  }

  validateEndpointsForDeployment() {
    return this.model.DeploymentType == 0 || !this.hasDockerEndpoint();
  }

  $onInit() {
    this.checkEndpointTypes(this.model.EdgeGroups);
  }
}

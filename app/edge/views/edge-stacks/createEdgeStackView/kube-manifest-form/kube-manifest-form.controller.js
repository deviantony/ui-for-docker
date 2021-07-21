class KubeManifestFormController {
  /* @ngInject */
  constructor() {
    this.methodOptions = [
      { id: 'method_editor', icon: 'fa fa-edit', label: 'Web editor', description: 'Use our Web editor', value: 'editor' },
      { id: 'method_upload', icon: 'fa fa-upload', label: 'Upload', description: 'Upload from your computer', value: 'upload' },
      { id: 'method_repository', icon: 'fab fa-github', label: 'Repository', description: 'Use a git repository', value: 'repository' },
    ];

    this.onChangeFileContent = this.onChangeFileContent.bind(this);
    this.onChangeFormValues = this.onChangeFormValues.bind(this);
    this.onChangeFile = this.onChangeFile.bind(this);
  }

  onChangeFormValues(values) {
    this.formValues = values;
  }

  onChangeFileContent(value) {
    this.state.isEditorDirty = true;
    this.formValues.StackFileContent = value;
  }

  onChangeFile(value) {
    this.formValues.StackFile = value;
  }
}

export default KubeManifestFormController;

class DockerComposeFormController {
  /* @ngInject */
  constructor($async, EdgeTemplateService, Notifications) {
    Object.assign(this, { $async, EdgeTemplateService, Notifications });

    this.methodOptions = [
      { id: 'method_editor', icon: 'fa fa-edit', label: 'Web editor', description: 'Use our Web editor', value: 'editor' },
      { id: 'method_upload', icon: 'fa fa-upload', label: 'Upload', description: 'Upload from your computer', value: 'upload' },
      { id: 'method_repository', icon: 'fab fa-github', label: 'Repository', description: 'Use a git repository', value: 'repository' },
      { id: 'method_template', icon: 'fa fa-rocket', label: 'Template', description: 'Use an Edge stack template', value: 'template' },
    ];

    this.selectedTemplate = null;

    this.onChangeFileContent = this.onChangeFileContent.bind(this);
    this.onChangeFile = this.onChangeFile.bind(this);
    this.onChangeTemplate = this.onChangeTemplate.bind(this);
    this.onChangeMethod = this.onChangeMethod.bind(this);
    this.onChangeFormValues = this.onChangeFormValues.bind(this);
  }

  onChangeFormValues(values) {
    this.formValues = values;
  }

  onChangeMethod() {
    this.formValues.StackFileContent = '';
    this.selectedTemplate = null;
  }

  onChangeTemplate(template) {
    return this.$async(async () => {
      this.formValues.StackFileContent = '';
      try {
        const fileContent = await this.EdgeTemplateService.edgeTemplate(template);
        this.formValues.StackFileContent = fileContent;
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to retrieve Template');
      }
    });
  }

  onChangeFileContent(value) {
    this.formValues.StackFileContent = value;
    this.state.isEditorDirty = true;
  }

  onChangeFile(value) {
    this.formValues.StackFile = value;
  }

  async $onInit() {
    return this.$async(async () => {
      try {
        const templates = await this.EdgeTemplateService.edgeTemplates();
        this.templates = templates.map((template) => ({ ...template, label: `${template.title} - ${template.description}` }));
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to retrieve Templates');
      }
    });
  }
}

export default DockerComposeFormController;

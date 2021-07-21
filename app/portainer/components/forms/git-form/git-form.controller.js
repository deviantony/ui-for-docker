export default class GitFormController {
  /* @ngInject */
  constructor() {
    this.onChangeField = this.onChangeField.bind(this);
    this.onChangeURL = this.onChangeField('RepositoryURL');
    this.onChangeRefName = this.onChangeField('RepositoryReferenceName');
    this.onChangeComposePath = this.onChangeField('ComposeFilePathInRepository');
  }

  onChangeField(field) {
    return (value) => {
      this.onChange({
        ...this.model,
        [field]: value,
      });
    };
  }

  $onInit() {
    this.deployMethod = this.deployMethod || 'compose';
  }
}

class GitFormAdditionalFileItemController {
  /* @ngInject */
  constructor() {}

  onChangePath(value) {
    const fieldIsInvalid = typeof value === 'undefined';
    if (fieldIsInvalid) {
      return;
    }
    this.onChange(this.index, { value });
  }

  removeValue() {
    this.onChange(this.index);
  }

  $onInit() {
    this.formName = `variableForm${this.index}`;
  }
}

export default GitFormAdditionalFileItemController;

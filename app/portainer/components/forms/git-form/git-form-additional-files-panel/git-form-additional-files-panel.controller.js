class GitFormAutoUpdateFieldsetController {
  /* @ngInject */
  constructor() {
    this.add = this.add.bind(this);
    this.onChangeVariable = this.onChangeVariable.bind(this);
  }

  add() {
    this.model.AdditionalFiles.push('');
  }

  onChangeVariable(index, variable) {
    if (!variable) {
      this.model.AdditionalFiles.splice(index, 1);
    } else {
      this.model.AdditionalFiles[index] = variable.value;
    }

    this.onChange({
      ...this.model,
      AdditionalFiles: this.model.AdditionalFiles,
    });
  }
}

export default GitFormAutoUpdateFieldsetController;

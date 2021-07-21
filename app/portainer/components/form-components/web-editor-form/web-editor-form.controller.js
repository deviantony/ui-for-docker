class WebEditorFormController {
  /* @ngInject */
  constructor() {
    this.editorUpdate = this.editorUpdate.bind(this);
  }

  editorUpdate(cm) {
    this.onChange(cm.getValue());
  }
}

export default WebEditorFormController;

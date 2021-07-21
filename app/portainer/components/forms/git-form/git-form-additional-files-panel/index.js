import controller from './git-form-additional-files-panel.controller.js';

export const gitFormAdditionalFilesPanel = {
  templateUrl: './git-form-additional-files-panel.html',
  controller,
  bindings: {
    model: '<',
    onChange: '<',
  },
};

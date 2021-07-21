import controller from './git-form-auto-update-fieldset.controller.js';

export const gitFormAutoUpdateFieldset = {
  templateUrl: './git-form-auto-update-fieldset.html',
  controller,
  bindings: {
    model: '<',
    onChange: '<',
  },
};

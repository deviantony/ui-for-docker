import controller from './git-form.controller.js';

export const gitForm = {
  templateUrl: './git-form.html',
  controller,
  bindings: {
    model: '<',
    onChange: '<',
    additionalFile: '<',
    autoUpdate: '<',
    showAuthExplanation: '<',
  },
};

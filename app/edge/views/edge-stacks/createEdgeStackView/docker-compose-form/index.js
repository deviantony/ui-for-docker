import controller from './docker-compose-form.controller.js';

export const edgeStacksDockerComposeForm = {
  templateUrl: './docker-compose-form.html',
  controller,

  bindings: {
    formValues: '=',
    state: '=',
  },
};

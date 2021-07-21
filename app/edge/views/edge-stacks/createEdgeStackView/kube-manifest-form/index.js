import controller from './kube-manifest-form.controller.js';

export const kubeManifestForm = {
  templateUrl: './kube-manifest-form.html',
  controller,

  bindings: {
    formValues: '=',
    state: '=',
  },
};

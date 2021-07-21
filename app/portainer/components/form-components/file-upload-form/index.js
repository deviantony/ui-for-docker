export const fileUploadForm = {
  templateUrl: './file-upload-form.html',

  bindings: {
    file: '<',
    ngRequired: '<',
    onChange: '<',
  },

  transclude: {
    description: '?fileUploadDescription',
  },
};

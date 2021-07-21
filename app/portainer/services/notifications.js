import toastr from 'toastr';

angular.module('portainer.app').factory('Notifications', [
  '$sanitize',
  function NotificationsFactory($sanitize) {
    'use strict';
    var service = {};

    service.success = function (title, text) {
      toastr.success($sanitize(text), $sanitize(title));
    };

    service.warning = function (title, text) {
      toastr.warning($sanitize(text), $sanitize(title), { timeOut: 6000 });
    };

    service.error = function (title, e, fallbackText) {
      var msg = fallbackText;
      if (e.err && e.err.data && e.err.data.details) {
        msg = e.err.data.details;
      } else if (e.err && e.err.data && e.err.data.message) {
        msg = e.err.data.message;
      } else if (e.data && e.data.details) {
        msg = e.data.details;
      } else if (e.data && e.data.message) {
        msg = e.data.message;
      } else if (e.data && e.data.content) {
        msg = e.data.content;
      } else if (e.data && e.data.error) {
        msg = e.data.error;
      } else if (e.message) {
        msg = e.message;
      } else if (e.err && e.err.data && e.err.data.length > 0 && e.err.data[0].message) {
        msg = e.err.data[0].message;
      } else if (e.err && e.err.data && e.err.data.err) {
        msg = e.err.data.err;
      } else if (e.data && e.data.err) {
        msg = e.data.err;
      } else if (e.msg) {
        msg = e.msg;
      }

      // eslint-disable-next-line no-console
      console.error(e);

      if (msg !== 'Invalid JWT token') {
        toastr.error($sanitize(msg), $sanitize(title), { timeOut: 6000 });
      }
    };

    return service;
  },
]);

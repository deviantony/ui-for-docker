import toastr from 'toastr';
import { Terminal } from 'xterm';
import * as fit from 'xterm/lib/addons/fit/fit';

/* @ngInject */
export function configApp($urlRouterProvider, $httpProvider, localStorageServiceProvider, jwtOptionsProvider, $uibTooltipProvider, $compileProvider, cfpLoadingBarProvider) {
  var environment = '@@ENVIRONMENT';
  if (environment === 'production') {
    $compileProvider.debugInfoEnabled(false);
  }

  localStorageServiceProvider.setPrefix('portainer');

  jwtOptionsProvider.config({
    tokenGetter: /* @ngInject */ function tokenGetter(LocalStorage) {
      return LocalStorage.getJWT();
    },
  });
  $httpProvider.interceptors.push('jwtInterceptor');
  $httpProvider.interceptors.push('EndpointStatusInterceptor');
  $httpProvider.defaults.headers.post['Content-Type'] = 'application/json';
  $httpProvider.defaults.headers.put['Content-Type'] = 'application/json';
  $httpProvider.defaults.headers.patch['Content-Type'] = 'application/json';

  $httpProvider.interceptors.push(
    /* @ngInject */ function (HttpRequestHelper) {
      return {
        request(config) {
          if (config.url.indexOf('/docker/') > -1) {
            config.headers['X-PortainerAgent-Target'] = HttpRequestHelper.portainerAgentTargetHeader();
            if (HttpRequestHelper.portainerAgentManagerOperation()) {
              config.headers['X-PortainerAgent-ManagerOperation'] = '1';
            }
          }
          return config;
        },
      };
    }
  );

  toastr.options.timeOut = 3000;

  Terminal.applyAddon(fit);

  $uibTooltipProvider.setTriggers({
    mouseenter: 'mouseleave',
    click: 'click',
    focus: 'blur',
    outsideClick: 'outsideClick',
  });

  cfpLoadingBarProvider.includeSpinner = false;
  cfpLoadingBarProvider.parentSelector = '#loadingbar-placeholder';
  cfpLoadingBarProvider.latencyThreshold = 600;

  $urlRouterProvider.otherwise('/auth');
}

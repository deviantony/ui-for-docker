import $ from 'jquery';

/* @ngInject */
export function onStartupAngular($rootScope, $state, $interval, LocalStorage, EndpointProvider, SystemService, cfpLoadingBar, $transitions, HttpRequestHelper) {
  EndpointProvider.initialize();

  $rootScope.$state = $state;
  $rootScope.defaultTitle = document.title;

  // Workaround to prevent the loading bar from going backward
  // https://github.com/chieffancypants/angular-loading-bar/issues/273
  const originalSet = cfpLoadingBar.set;
  cfpLoadingBar.set = function overrideSet(n) {
    if (n > cfpLoadingBar.status()) {
      originalSet.apply(cfpLoadingBar, arguments);
    }
  };

  $transitions.onBefore({}, () => {
    HttpRequestHelper.resetAgentHeaders();
  });

  $state.defaultErrorHandler(() => {
    // Do not log transitionTo errors
  });

  // Keep-alive Edge endpoints by sending a ping request every minute
  $interval(() => {
    ping(EndpointProvider, SystemService);
  }, 60 * 1000);

  $(document).ajaxSend((event, jqXhr, jqOpts) => {
    const type = jqOpts.type === 'POST' || jqOpts.type === 'PUT' || jqOpts.type === 'PATCH';
    const hasNoContentType = jqOpts.contentType !== 'application/json' && jqOpts.headers && !jqOpts.headers['Content-Type'];
    if (type && hasNoContentType) {
      jqXhr.setRequestHeader('Content-Type', 'application/json');
    }
    jqXhr.setRequestHeader('Authorization', 'Bearer ' + LocalStorage.getJWT());
  });
}

function ping(EndpointProvider, SystemService) {
  const endpoint = EndpointProvider.currentEndpoint();
  if (endpoint !== undefined && endpoint.Type === 4) {
    SystemService.ping(endpoint.Id);
  }
}

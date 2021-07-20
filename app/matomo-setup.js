const _paq = (window._paq = window._paq || []);
/* tracker methods like "setCustomDimension" should be called before "trackPageView" */
_paq.push(['enableLinkTracking']);

var u = 'https://portainer-ce.matomo.cloud/';
_paq.push(['setTrackerUrl', u + 'matomo.php']);
_paq.push(['setSiteId', '1']);
var d = document,
  g = d.createElement('script'),
  s = d.getElementsByTagName('script')[0];
g.type = 'text/javascript';
g.async = true;
g.src = 'https://cdn.matomo.cloud/portainer-ce.matomo.cloud/matomo.js';
s.parentNode.insertBefore(g, s);

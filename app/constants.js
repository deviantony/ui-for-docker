export const API_ENDPOINT_AUTH = 'api/auth';
export const API_ENDPOINT_BACKUP = 'api/backup';
export const API_ENDPOINT_DOCKERHUB = 'api/dockerhub';
export const API_ENDPOINT_CUSTOM_TEMPLATES = 'api/custom_templates';
export const API_ENDPOINT_EDGE_GROUPS = 'api/edge_groups';
export const API_ENDPOINT_EDGE_JOBS = 'api/edge_jobs';
export const API_ENDPOINT_EDGE_STACKS = 'api/edge_stacks';
export const API_ENDPOINT_EDGE_TEMPLATES = 'api/edge_templates';
export const API_ENDPOINT_ENDPOINTS = 'api/endpoints';
export const API_ENDPOINT_ENDPOINT_GROUPS = 'api/endpoint_groups';
export const API_ENDPOINT_MOTD = 'api/motd';
export const API_ENDPOINT_EXTENSIONS = 'api/extensions';
export const API_ENDPOINT_REGISTRIES = 'api/registries';
export const API_ENDPOINT_RESOURCE_CONTROLS = 'api/resource_controls';
export const API_ENDPOINT_SETTINGS = 'api/settings';
export const API_ENDPOINT_STACKS = 'api/stacks';
export const API_ENDPOINT_STATUS = 'api/status';
export const API_ENDPOINT_SUPPORT = 'api/support';
export const API_ENDPOINT_USERS = 'api/users';
export const API_ENDPOINT_TAGS = 'api/tags';
export const API_ENDPOINT_TEAMS = 'api/teams';
export const API_ENDPOINT_TEAM_MEMBERSHIPS = 'api/team_memberships';
export const API_ENDPOINT_TEMPLATES = 'api/templates';
export const API_ENDPOINT_WEBHOOKS = 'api/webhooks';
export const DEFAULT_TEMPLATES_URL = 'https://raw.githubusercontent.com/portainer/templates/master/templates.json';
export const PAGINATION_MAX_ITEMS = 10;
export const APPLICATION_CACHE_VALIDITY = 3600;
export const CONSOLE_COMMANDS_LABEL_PREFIX = 'io.portainer.commands.';
export const PREDEFINED_NETWORKS = ['host', 'bridge', 'none'];
export const KUBERNETES_DEFAULT_NAMESPACE = 'default';
export const KUBERNETES_SYSTEM_NAMESPACES = ['kube-system', 'kube-public', 'kube-node-lease', 'portainer'];
export const PORTAINER_FADEOUT = 1500;

// don't declare new constants, either:
// - if only used in one file or module, declare in that file or module (as a regular js constant)
// - if needed across modules, declare like the above constant and use es6 import for that
angular
  .module('portainer')
  .constant('API_ENDPOINT_AUTH', API_ENDPOINT_AUTH)
  .constant('API_ENDPOINT_BACKUP', API_ENDPOINT_BACKUP)
  .constant('API_ENDPOINT_DOCKERHUB', API_ENDPOINT_DOCKERHUB)
  .constant('API_ENDPOINT_CUSTOM_TEMPLATES', API_ENDPOINT_CUSTOM_TEMPLATES)
  .constant('API_ENDPOINT_EDGE_GROUPS', API_ENDPOINT_EDGE_GROUPS)
  .constant('API_ENDPOINT_EDGE_JOBS', API_ENDPOINT_EDGE_JOBS)
  .constant('API_ENDPOINT_EDGE_STACKS', API_ENDPOINT_EDGE_STACKS)
  .constant('API_ENDPOINT_EDGE_TEMPLATES', API_ENDPOINT_EDGE_TEMPLATES)
  .constant('API_ENDPOINT_ENDPOINTS', API_ENDPOINT_ENDPOINTS)
  .constant('API_ENDPOINT_ENDPOINT_GROUPS', API_ENDPOINT_ENDPOINT_GROUPS)
  .constant('API_ENDPOINT_MOTD', API_ENDPOINT_MOTD)
  .constant('API_ENDPOINT_REGISTRIES', API_ENDPOINT_REGISTRIES)
  .constant('API_ENDPOINT_RESOURCE_CONTROLS', API_ENDPOINT_RESOURCE_CONTROLS)
  .constant('API_ENDPOINT_SETTINGS', API_ENDPOINT_SETTINGS)
  .constant('API_ENDPOINT_STACKS', API_ENDPOINT_STACKS)
  .constant('API_ENDPOINT_STATUS', API_ENDPOINT_STATUS)
  .constant('API_ENDPOINT_SUPPORT', API_ENDPOINT_SUPPORT)
  .constant('API_ENDPOINT_USERS', API_ENDPOINT_USERS)
  .constant('API_ENDPOINT_TAGS', API_ENDPOINT_TAGS)
  .constant('API_ENDPOINT_TEAMS', API_ENDPOINT_TEAMS)
  .constant('API_ENDPOINT_TEAM_MEMBERSHIPS', API_ENDPOINT_TEAM_MEMBERSHIPS)
  .constant('API_ENDPOINT_TEMPLATES', API_ENDPOINT_TEMPLATES)
  .constant('API_ENDPOINT_WEBHOOKS', API_ENDPOINT_WEBHOOKS)
  .constant('DEFAULT_TEMPLATES_URL', DEFAULT_TEMPLATES_URL)
  .constant('PAGINATION_MAX_ITEMS', PAGINATION_MAX_ITEMS)
  .constant('APPLICATION_CACHE_VALIDITY', APPLICATION_CACHE_VALIDITY)
  .constant('CONSOLE_COMMANDS_LABEL_PREFIX', CONSOLE_COMMANDS_LABEL_PREFIX)
  .constant('PREDEFINED_NETWORKS', PREDEFINED_NETWORKS)
  .constant('KUBERNETES_DEFAULT_NAMESPACE', KUBERNETES_DEFAULT_NAMESPACE)
  .constant('KUBERNETES_SYSTEM_NAMESPACES', KUBERNETES_SYSTEM_NAMESPACES);

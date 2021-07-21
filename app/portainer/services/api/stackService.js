import _ from 'lodash-es';
import { StackViewModel, OrphanedStackViewModel } from '../../models/stack';

angular.module('portainer.app').factory('StackService', [
  '$q',
  '$async',
  'Stack',
  'FileUploadService',
  'StackHelper',
  'ServiceService',
  'ContainerService',
  'SwarmService',
  'EndpointProvider',
  function StackServiceFactory($q, $async, Stack, FileUploadService, StackHelper, ServiceService, ContainerService, SwarmService, EndpointProvider) {
    'use strict';
    var service = {
      updateGit,
    };

    service.stack = function (id) {
      var deferred = $q.defer();

      Stack.get({ id: id })
        .$promise.then(function success(data) {
          var stack = new StackViewModel(data);
          deferred.resolve(stack);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve stack details', err: err });
        });

      return deferred.promise;
    };

    service.getStackFile = function (id) {
      var deferred = $q.defer();

      Stack.getStackFile({ id: id })
        .$promise.then(function success(data) {
          deferred.resolve(data.StackFileContent);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve stack content', err: err });
        });

      return deferred.promise;
    };

    service.migrateSwarmStack = function (stack, targetEndpointId, newName) {
      var deferred = $q.defer();

      EndpointProvider.setEndpointID(targetEndpointId);

      SwarmService.swarm()
        .then(function success(data) {
          var swarm = data;
          if (swarm.Id === stack.SwarmId) {
            deferred.reject({ msg: 'Target endpoint is located in the same Swarm cluster as the current endpoint', err: null });
            return;
          }
          return Stack.migrate({ id: stack.Id, endpointId: stack.EndpointId }, { EndpointID: targetEndpointId, SwarmID: swarm.Id, Name: newName }).$promise;
        })
        .then(function success() {
          deferred.resolve();
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to migrate stack', err: err });
        })
        .finally(function final() {
          EndpointProvider.setEndpointID(stack.EndpointId);
        });

      return deferred.promise;
    };

    service.migrateComposeStack = function (stack, targetEndpointId, newName) {
      var deferred = $q.defer();

      EndpointProvider.setEndpointID(targetEndpointId);

      Stack.migrate({ id: stack.Id, endpointId: stack.EndpointId }, { EndpointID: targetEndpointId, Name: newName })
        .$promise.then(function success() {
          deferred.resolve();
        })
        .catch(function error(err) {
          EndpointProvider.setEndpointID(stack.EndpointId);
          deferred.reject({ msg: 'Unable to migrate stack', err: err });
        });

      return deferred.promise;
    };

    service.stacks = function (compose, swarm, endpointId, includeOrphanedStacks = false) {
      var deferred = $q.defer();

      var queries = [];
      if (compose) {
        queries.push(service.composeStacks(endpointId, true, { EndpointID: endpointId, IncludeOrphanedStacks: includeOrphanedStacks }));
      }
      if (swarm) {
        queries.push(service.swarmStacks(endpointId, true, { IncludeOrphanedStacks: includeOrphanedStacks }));
      }

      $q.all(queries)
        .then(function success(data) {
          var stacks = [];
          if (data[0]) {
            stacks = stacks.concat(data[0]);
          }
          if (data[1]) {
            stacks = stacks.concat(data[1]);
          }
          deferred.resolve(stacks);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve stacks', err: err });
        });

      return deferred.promise;
    };

    service.externalSwarmStacks = function () {
      var deferred = $q.defer();

      ServiceService.services()
        .then(function success(services) {
          deferred.resolve(StackHelper.getExternalStacksFromServices(services));
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve external stacks', err: err });
        });

      return deferred.promise;
    };

    service.externalComposeStacks = function () {
      var deferred = $q.defer();

      ContainerService.containers(1)
        .then(function success(containers) {
          deferred.resolve(StackHelper.getExternalStacksFromContainers(containers));
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve external stacks', err: err });
        });

      return deferred.promise;
    };

    service.unionStacks = function (stacks, externalStacks) {
      stacks.forEach((stack) => {
        externalStacks.forEach((externalStack) => {
          if (stack.Orphaned && stack.Name == externalStack.Name) {
            stack.OrphanedRunning = true;
          }
        });
      });
      const result = _.unionWith(stacks, externalStacks, function (a, b) {
        return a.Name === b.Name;
      });

      return result;
    };

    service.composeStacks = function (endpointId, includeExternalStacks, filters) {
      var deferred = $q.defer();

      $q.all({
        stacks: Stack.query({ filters: filters }).$promise,
        externalStacks: includeExternalStacks ? service.externalComposeStacks() : [],
      })
        .then(function success(data) {
          var stacks = data.stacks.map(function (item) {
            if (item.EndpointId == endpointId) {
              return new StackViewModel(item);
            } else {
              return new OrphanedStackViewModel(item);
            }
          });

          var externalStacks = data.externalStacks;
          const result = service.unionStacks(stacks, externalStacks);
          deferred.resolve(result);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve stacks', err: err });
        });

      return deferred.promise;
    };

    service.swarmStacks = function (endpointId, includeExternalStacks, filters = {}) {
      var deferred = $q.defer();

      SwarmService.swarm()
        .then(function success(data) {
          var swarm = data;
          filters = { SwarmID: swarm.Id, ...filters };

          return $q.all({
            stacks: Stack.query({ filters: filters }).$promise,
            externalStacks: includeExternalStacks ? service.externalSwarmStacks() : [],
          });
        })
        .then(function success(data) {
          var stacks = data.stacks.map(function (item) {
            if (item.EndpointId == endpointId) {
              return new StackViewModel(item);
            } else {
              return new OrphanedStackViewModel(item);
            }
          });

          var externalStacks = data.externalStacks;
          const result = service.unionStacks(stacks, externalStacks);
          deferred.resolve(result);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve stacks', err: err });
        });

      return deferred.promise;
    };

    service.remove = function (stack, external, endpointId) {
      var deferred = $q.defer();

      Stack.remove({ id: stack.Id ? stack.Id : stack.Name, external: external, endpointId: endpointId })
        .$promise.then(function success() {
          deferred.resolve();
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to remove the stack', err: err });
        });

      return deferred.promise;
    };

    service.associate = function (stack, endpointId, orphanedRunning) {
      var deferred = $q.defer();

      if (stack.Type == 1) {
        SwarmService.swarm()
          .then(function success(data) {
            const swarm = data;
            return Stack.associate({ id: stack.Id, endpointId: endpointId, swarmId: swarm.Id, orphanedRunning }).$promise;
          })
          .then(function success(data) {
            deferred.resolve(data);
          })
          .catch(function error(err) {
            deferred.reject({ msg: 'Unable to associate the stack', err: err });
          });
      } else {
        Stack.associate({ id: stack.Id, endpointId: endpointId, orphanedRunning })
          .$promise.then(function success(data) {
            deferred.resolve(data);
          })
          .catch(function error(err) {
            deferred.reject({ msg: 'Unable to associate the stack', err: err });
          });
      }

      return deferred.promise;
    };

    service.updateStack = function (stack, stackFile, env, prune) {
      return Stack.update({ endpointId: stack.EndpointId }, { id: stack.Id, StackFileContent: stackFile, Env: env, Prune: prune }).$promise;
    };

    service.createComposeStackFromFileUpload = function (name, stackFile, env, endpointId) {
      return FileUploadService.createComposeStack(name, stackFile, env, endpointId);
    };

    service.createSwarmStackFromFileUpload = function (name, stackFile, env, endpointId) {
      var deferred = $q.defer();

      SwarmService.swarm()
        .then(function success(data) {
          var swarm = data;
          return FileUploadService.createSwarmStack(name, swarm.Id, stackFile, env, endpointId);
        })
        .then(function success(data) {
          deferred.resolve(data.data);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to create the stack', err: err });
        });

      return deferred.promise;
    };
    service.createComposeStackFromFileContent = function (name, stackFileContent, env, endpointId) {
      var payload = {
        Name: name,
        StackFileContent: stackFileContent,
        Env: env,
      };
      return Stack.create({ method: 'string', type: 2, endpointId: endpointId }, payload).$promise;
    };

    service.createSwarmStackFromFileContent = function (name, stackFileContent, env, endpointId) {
      var deferred = $q.defer();

      SwarmService.swarm()
        .then(function success(swarm) {
          var payload = {
            Name: name,
            SwarmID: swarm.Id,
            StackFileContent: stackFileContent,
            Env: env,
          };
          return Stack.create({ method: 'string', type: 1, endpointId: endpointId }, payload).$promise;
        })
        .then(function success(data) {
          deferred.resolve(data);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to create the stack', err: err });
        });

      return deferred.promise;
    };

    service.createComposeStackFromGitRepository = function (name, repositoryOptions, env, endpointId) {
      var payload = {
        Name: name,
        RepositoryURL: repositoryOptions.RepositoryURL,
        RepositoryReferenceName: repositoryOptions.RepositoryReferenceName,
        ComposeFile: repositoryOptions.ComposeFilePathInRepository,
        AdditionalFiles: repositoryOptions.AdditionalFiles,
        RepositoryAuthentication: repositoryOptions.RepositoryAuthentication,
        RepositoryUsername: repositoryOptions.RepositoryUsername,
        RepositoryPassword: repositoryOptions.RepositoryPassword,
        Env: env,
      };

      if (repositoryOptions.AutoUpdate) {
        payload.AutoUpdate = repositoryOptions.AutoUpdate;
      }

      return Stack.create({ method: 'repository', type: 2, endpointId: endpointId }, payload).$promise;
    };

    service.createSwarmStackFromGitRepository = function (name, repositoryOptions, env, endpointId) {
      var deferred = $q.defer();

      SwarmService.swarm()
        .then(function success(data) {
          var swarm = data;
          var payload = {
            Name: name,
            SwarmID: swarm.Id,
            RepositoryURL: repositoryOptions.RepositoryURL,
            RepositoryReferenceName: repositoryOptions.RepositoryReferenceName,
            ComposeFile: repositoryOptions.ComposeFilePathInRepository,
            AdditionalFiles: repositoryOptions.AdditionalFiles,
            RepositoryAuthentication: repositoryOptions.RepositoryAuthentication,
            RepositoryUsername: repositoryOptions.RepositoryUsername,
            RepositoryPassword: repositoryOptions.RepositoryPassword,
            Env: env,
          };

          if (repositoryOptions.AutoUpdate) {
            payload.AutoUpdate = repositoryOptions.AutoUpdate;
          }

          return Stack.create({ method: 'repository', type: 1, endpointId: endpointId }, payload).$promise;
        })
        .then(function success(data) {
          deferred.resolve(data);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to create the stack', err: err });
        });

      return deferred.promise;
    };

    service.duplicateStack = function duplicateStack(name, stackFileContent, env, endpointId, type) {
      var action = type === 1 ? service.createSwarmStackFromFileContent : service.createComposeStackFromFileContent;
      return action(name, stackFileContent, env, endpointId);
    };

    async function kubernetesDeployAsync(endpointId, method, payload) {
      try {
        await Stack.create({ endpointId: endpointId, method: method, type: 3 }, payload).$promise;
      } catch (err) {
        throw { err: err };
      }
    }

    service.kubernetesDeploy = function (endpointId, method, payload) {
      return $async(kubernetesDeployAsync, endpointId, method, payload);
    };

    service.start = start;
    function start(id) {
      return Stack.start({ id }).$promise;
    }

    service.stop = stop;
    function stop(id) {
      return Stack.stop({ id }).$promise;
    }

    function updateGit(id, endpointId, env, prune, gitConfig) {
      return Stack.updateGit(
        { endpointId, id },
        {
          env,
          prune,
          RepositoryReferenceName: gitConfig.RefName,
          RepositoryAuthentication: gitConfig.RepositoryAuthentication,
          RepositoryUsername: gitConfig.RepositoryUsername,
          RepositoryPassword: gitConfig.RepositoryPassword,
        }
      ).$promise;
    }

    service.updateGitStackSettings = function (id, endpointId, env, gitConfig) {
      // prepare auto update
      const autoUpdate = {};

      if (gitConfig.AutoUpdate.RepositoryAutomaticUpdates) {
        if (gitConfig.AutoUpdate.RepositoryMechanism === 'Interval') {
          autoUpdate.Interval = gitConfig.AutoUpdate.RepositoryFetchInterval;
        } else if (gitConfig.AutoUpdate.RepositoryMechanism === 'Webhook') {
          autoUpdate.Webhook = gitConfig.AutoUpdate.RepositoryWebhookURL.split('/').reverse()[0];
        }
      }

      return Stack.updateGitStackSettings(
        { endpointId, id },
        {
          AutoUpdate: autoUpdate,
          Env: env,
          RepositoryReferenceName: gitConfig.RefName,
          RepositoryAuthentication: gitConfig.RepositoryAuthentication,
          RepositoryUsername: gitConfig.RepositoryUsername,
          RepositoryPassword: gitConfig.RepositoryPassword,
        }
      ).$promise;
    };

    return service;
  },
]);

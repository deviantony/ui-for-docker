import _ from 'lodash-es';
import * as simpleDuration from 'simple-duration';
import { PorImageRegistryModel } from 'Docker/models/porImageRegistry';
import { ContainerCapabilities, ContainerCapability } from '../../../models/containerCapabilities';
import { AccessControlFormData } from '../../../../portainer/components/accessControlForm/porAccessControlFormModel';
import { ContainerDetailsViewModel } from '../../../models/container';

angular.module('portainer.docker').controller('CreateContainerController', [
  '$q',
  '$scope',
  '$async',
  '$state',
  '$timeout',
  '$transition$',
  '$filter',
  'Container',
  'ContainerHelper',
  'Image',
  'ImageHelper',
  'Volume',
  'NetworkService',
  'ResourceControlService',
  'Authentication',
  'Notifications',
  'ContainerService',
  'ImageService',
  'FormValidator',
  'ModalService',
  'RegistryService',
  'SystemService',
  'SettingsService',
  'PluginService',
  'HttpRequestHelper',
  function (
    $q,
    $scope,
    $async,
    $state,
    $timeout,
    $transition$,
    $filter,
    Container,
    ContainerHelper,
    Image,
    ImageHelper,
    Volume,
    NetworkService,
    ResourceControlService,
    Authentication,
    Notifications,
    ContainerService,
    ImageService,
    FormValidator,
    ModalService,
    RegistryService,
    SystemService,
    SettingsService,
    PluginService,
    HttpRequestHelper
  ) {
    $scope.create = create;

    $scope.formValues = {
      alwaysPull: true,
      Console: 'none',
      Volumes: [],
      NetworkContainer: null,
      Labels: [],
      ExtraHosts: [],
      MacAddress: '',
      IPv4: '',
      IPv6: '',
      DnsPrimary: '',
      DnsSecondary: '',
      AccessControlData: new AccessControlFormData(),
      CpuLimit: 0,
      MemoryLimit: 0,
      MemoryReservation: 0,
      CmdMode: 'default',
      EntrypointMode: 'default',
      NodeName: null,
      capabilities: [],
      LogDriverName: '',
      LogDriverOpts: [],
      RegistryModel: new PorImageRegistryModel(),
      Healthcheck: {},
    };

    $scope.extraNetworks = {};

    $scope.state = {
      formValidationError: '',
      actionInProgress: false,
      mode: '',
    };

    $scope.refreshSlider = function () {
      $timeout(function () {
        $scope.$broadcast('rzSliderForceRender');
      });
    };

    $scope.onImageNameChange = function () {
      $scope.formValues.CmdMode = 'default';
      $scope.formValues.EntrypointMode = 'default';
    };

    $scope.config = {
      Image: '',
      Env: [],
      Cmd: '',
      MacAddress: '',
      ExposedPorts: {},
      Entrypoint: '',
      HostConfig: {
        RestartPolicy: {
          Name: 'no',
        },
        PortBindings: [],
        PublishAllPorts: false,
        Binds: [],
        AutoRemove: false,
        NetworkMode: 'bridge',
        Privileged: false,
        Init: false,
        Runtime: null,
        ExtraHosts: [],
        Devices: [],
        CapAdd: [],
        CapDrop: [],
      },
      NetworkingConfig: {
        EndpointsConfig: {},
      },
      Labels: {},
      Healthcheck: {},
    };

    $scope.addVolume = function () {
      $scope.formValues.Volumes.push({ name: '', containerPath: '', readOnly: false, type: 'volume' });
    };

    $scope.removeVolume = function (index) {
      $scope.formValues.Volumes.splice(index, 1);
    };

    $scope.addEnvironmentVariable = function () {
      $scope.config.Env.push({ name: '', value: '' });
    };

    $scope.removeEnvironmentVariable = function (index) {
      $scope.config.Env.splice(index, 1);
    };

    $scope.addPortBinding = function () {
      $scope.config.HostConfig.PortBindings.push({ hostPort: '', containerPort: '', protocol: 'tcp' });
    };

    $scope.removePortBinding = function (index) {
      $scope.config.HostConfig.PortBindings.splice(index, 1);
    };

    $scope.addLabel = function () {
      $scope.formValues.Labels.push({ name: '', value: '' });
    };

    $scope.removeLabel = function (index) {
      $scope.formValues.Labels.splice(index, 1);
    };

    $scope.addExtraHost = function () {
      $scope.formValues.ExtraHosts.push({ value: '' });
    };

    $scope.removeExtraHost = function (index) {
      $scope.formValues.ExtraHosts.splice(index, 1);
    };

    $scope.addDevice = function () {
      $scope.config.HostConfig.Devices.push({ pathOnHost: '', pathInContainer: '' });
    };

    $scope.removeDevice = function (index) {
      $scope.config.HostConfig.Devices.splice(index, 1);
    };

    $scope.addLogDriverOpt = function () {
      $scope.formValues.LogDriverOpts.push({ name: '', value: '' });
    };

    $scope.removeLogDriverOpt = function (index) {
      $scope.formValues.LogDriverOpts.splice(index, 1);
    };

    $scope.removeHealthcheck = function () {
      $scope.formValues.Healthcheck = {};
    };

    $scope.fromContainerMultipleNetworks = false;

    function prepareImageConfig(config) {
      const imageConfig = ImageHelper.createImageConfigForContainer($scope.formValues.RegistryModel);
      config.Image = imageConfig.fromImage;
    }

    function preparePortBindings(config) {
      const bindings = ContainerHelper.preparePortBindings(config.HostConfig.PortBindings);
      config.ExposedPorts = {};
      _.forEach(bindings, (_, key) => (config.ExposedPorts[key] = {}));
      config.HostConfig.PortBindings = bindings;
    }

    function prepareConsole(config) {
      var value = $scope.formValues.Console;
      var openStdin = true;
      var tty = true;
      if (value === 'tty') {
        openStdin = false;
      } else if (value === 'interactive') {
        tty = false;
      } else if (value === 'none') {
        openStdin = false;
        tty = false;
      }
      config.OpenStdin = openStdin;
      config.Tty = tty;
    }

    function prepareCmd(config) {
      if (_.isEmpty(config.Cmd) || $scope.formValues.CmdMode == 'default') {
        delete config.Cmd;
      } else {
        config.Cmd = ContainerHelper.commandStringToArray(config.Cmd);
      }
    }

    function prepareEntrypoint(config) {
      if ($scope.formValues.EntrypointMode == 'default' || (_.isEmpty(config.Cmd) && _.isEmpty(config.Entrypoint))) {
        config.Entrypoint = null;
      }
    }

    function prepareEnvironmentVariables(config) {
      var env = [];
      config.Env.forEach(function (v) {
        if (v.name && v.value) {
          env.push(v.name + '=' + v.value);
        }
      });
      config.Env = env;
    }

    function prepareVolumes(config) {
      var binds = [];
      var volumes = {};

      $scope.formValues.Volumes.forEach(function (volume) {
        var name = volume.name;
        var containerPath = volume.containerPath;
        if (name && containerPath) {
          var bind = name + ':' + containerPath;
          volumes[containerPath] = {};
          if (volume.readOnly) {
            bind += ':ro';
          }
          binds.push(bind);
        }
      });
      config.HostConfig.Binds = binds;
      config.Volumes = volumes;
    }

    function prepareNetworkConfig(config) {
      var mode = config.HostConfig.NetworkMode;
      var container = $scope.formValues.NetworkContainer;
      var containerName = container;
      if (container && typeof container === 'object') {
        containerName = $filter('trimcontainername')(container.Names[0]);
      }
      var networkMode = mode;
      if (containerName) {
        networkMode += ':' + containerName;
        config.Hostname = '';
      }
      config.HostConfig.NetworkMode = networkMode;
      config.MacAddress = $scope.formValues.MacAddress;

      config.NetworkingConfig.EndpointsConfig[networkMode] = {
        IPAMConfig: {
          IPv4Address: $scope.formValues.IPv4,
          IPv6Address: $scope.formValues.IPv6,
        },
      };

      if (networkMode && _.get($scope.config.NetworkingConfig.EndpointsConfig[networkMode], 'Aliases')) {
        var aliases = $scope.config.NetworkingConfig.EndpointsConfig[networkMode].Aliases;
        config.NetworkingConfig.EndpointsConfig[networkMode].Aliases = _.filter(aliases, (o) => {
          return !_.startsWith($scope.fromContainer.Id, o);
        });
      }

      var dnsServers = [];
      if ($scope.formValues.DnsPrimary) {
        dnsServers.push($scope.formValues.DnsPrimary);
      }
      if ($scope.formValues.DnsSecondary) {
        dnsServers.push($scope.formValues.DnsSecondary);
      }
      config.HostConfig.Dns = dnsServers;

      config.HostConfig.ExtraHosts = _.map(
        _.filter($scope.formValues.ExtraHosts, (v) => v.value),
        'value'
      );
    }

    function prepareLabels(config) {
      var labels = {};
      $scope.formValues.Labels.forEach(function (label) {
        if (label.name) {
          if (label.value) {
            labels[label.name] = label.value;
          } else {
            labels[label.name] = '';
          }
        }
      });
      config.Labels = labels;
    }

    function prepareDevices(config) {
      var path = [];
      config.HostConfig.Devices.forEach(function (p) {
        if (p.pathOnHost) {
          if (p.pathInContainer === '') {
            p.pathInContainer = p.pathOnHost;
          }
          path.push({ PathOnHost: p.pathOnHost, PathInContainer: p.pathInContainer, CgroupPermissions: 'rwm' });
        }
      });
      config.HostConfig.Devices = path;
    }

    function prepareResources(config) {
      // Memory Limit - Round to 0.125
      if ($scope.formValues.MemoryLimit >= 0) {
        var memoryLimit = (Math.round($scope.formValues.MemoryLimit * 8) / 8).toFixed(3);
        memoryLimit *= 1024 * 1024;
        config.HostConfig.Memory = memoryLimit;
      }

      // Memory Resevation - Round to 0.125
      if ($scope.formValues.MemoryReservation >= 0) {
        var memoryReservation = (Math.round($scope.formValues.MemoryReservation * 8) / 8).toFixed(3);
        memoryReservation *= 1024 * 1024;
        config.HostConfig.MemoryReservation = memoryReservation;
      }

      // CPU Limit
      if ($scope.formValues.CpuLimit >= 0) {
        config.HostConfig.NanoCpus = $scope.formValues.CpuLimit * 1000000000;
      }
    }

    function prepareLogDriver(config) {
      var logOpts = {};
      if ($scope.formValues.LogDriverName) {
        config.HostConfig.LogConfig = { Type: $scope.formValues.LogDriverName };
        if ($scope.formValues.LogDriverName !== 'none') {
          $scope.formValues.LogDriverOpts.forEach(function (opt) {
            if (opt.name) {
              logOpts[opt.name] = opt.value;
            }
          });
          if (Object.keys(logOpts).length !== 0 && logOpts.constructor === Object) {
            config.HostConfig.LogConfig.Config = logOpts;
          }
        }
      }
    }

    function prepareCapabilities(config) {
      var allowed = $scope.formValues.capabilities.filter(function (item) {
        return item.allowed === true;
      });
      var notAllowed = $scope.formValues.capabilities.filter(function (item) {
        return item.allowed === false;
      });

      var getCapName = function (item) {
        return item.capability;
      };
      config.HostConfig.CapAdd = allowed.map(getCapName);
      config.HostConfig.CapDrop = notAllowed.map(getCapName);
    }

    $scope.durationRx = /((\d+(\.\d+)?)(ms|[smhdy]))/gi;

    function prepareHealthcheck(config) {
      config.Healthcheck = {};

      if (!$scope.formValues.Healthcheck.Type) return;

      config.Healthcheck.Test = [$scope.formValues.Healthcheck.Type];
      if ($scope.formValues.Healthcheck.Test) {
        config.Healthcheck.Test.push($scope.formValues.Healthcheck.Test);
      }

      if (!isNaN($scope.formValues.Healthcheck.Retries)) {
        config.Healthcheck.Retries = parseInt($scope.formValues.Healthcheck.Retries, 10);
      }

      // Docker uses nanos
      var duration = $scope.formValues.Healthcheck.StartPeriod.match($scope.durationRx);
      if (duration) {
        var startPeriodS = simpleDuration.parse(duration.join('').toLowerCase());
        config.Healthcheck.StartPeriod = startPeriodS * 1e9;
      }
      duration = $scope.formValues.Healthcheck.Interval.match($scope.durationRx);
      if (duration) {
        var intervalS = simpleDuration.parse(duration.join('').toLowerCase());
        config.Healthcheck.Interval = intervalS * 1e9;
      }
      duration = $scope.formValues.Healthcheck.Timeout.match($scope.durationRx);
      if (duration) {
        var timeoutS = simpleDuration.parse(duration.join('').toLowerCase());
        config.Healthcheck.Timeout = timeoutS * 1e9;
      }
    }

    function prepareConfiguration() {
      var config = angular.copy($scope.config);
      prepareCmd(config);
      prepareEntrypoint(config);
      prepareNetworkConfig(config);
      prepareImageConfig(config);
      preparePortBindings(config);
      prepareConsole(config);
      prepareEnvironmentVariables(config);
      prepareVolumes(config);
      prepareLabels(config);
      prepareDevices(config);
      prepareResources(config);
      prepareLogDriver(config);
      prepareCapabilities(config);
      prepareHealthcheck(config);
      return config;
    }

    function loadFromContainerCmd() {
      if ($scope.config.Cmd) {
        $scope.config.Cmd = ContainerHelper.commandArrayToString($scope.config.Cmd);
        $scope.formValues.CmdMode = 'override';
      }
    }

    function loadFromContainerEntrypoint() {
      if (_.has($scope.config, 'Entrypoint')) {
        if ($scope.config.Entrypoint == null) {
          $scope.config.Entrypoint = '';
        }
        $scope.formValues.EntrypointMode = 'override';
      }
    }

    function loadFromContainerPortBindings() {
      const bindings = ContainerHelper.sortAndCombinePorts($scope.config.HostConfig.PortBindings);
      $scope.config.HostConfig.PortBindings = bindings;
    }

    function loadFromContainerVolumes(d) {
      for (var v in d.Mounts) {
        if ({}.hasOwnProperty.call(d.Mounts, v)) {
          var mount = d.Mounts[v];
          var volume = {
            type: mount.Type,
            name: mount.Name || mount.Source,
            containerPath: mount.Destination,
            readOnly: mount.RW === false,
          };
          $scope.formValues.Volumes.push(volume);
        }
      }
    }

    $scope.resetNetworkConfig = function () {
      $scope.config.NetworkingConfig = {
        EndpointsConfig: {},
      };
    };

    function loadFromContainerNetworkConfig(d) {
      $scope.config.NetworkingConfig = {
        EndpointsConfig: {},
      };
      var networkMode = d.HostConfig.NetworkMode;
      if (networkMode === 'default') {
        $scope.config.HostConfig.NetworkMode = 'bridge';
        if (!_.find($scope.availableNetworks, { Name: 'bridge' })) {
          $scope.config.HostConfig.NetworkMode = 'nat';
        }
      }
      if ($scope.config.HostConfig.NetworkMode.indexOf('container:') === 0) {
        var netContainer = $scope.config.HostConfig.NetworkMode.split(/^container:/)[1];
        $scope.config.HostConfig.NetworkMode = 'container';
        for (var c in $scope.runningContainers) {
          if ($scope.runningContainers[c].Id == netContainer) {
            $scope.formValues.NetworkContainer = $scope.runningContainers[c];
          }
        }
      }
      $scope.fromContainerMultipleNetworks = Object.keys(d.NetworkSettings.Networks).length >= 2;
      if (d.NetworkSettings.Networks[$scope.config.HostConfig.NetworkMode]) {
        if (d.NetworkSettings.Networks[$scope.config.HostConfig.NetworkMode].IPAMConfig) {
          if (d.NetworkSettings.Networks[$scope.config.HostConfig.NetworkMode].IPAMConfig.IPv4Address) {
            $scope.formValues.IPv4 = d.NetworkSettings.Networks[$scope.config.HostConfig.NetworkMode].IPAMConfig.IPv4Address;
          }
          if (d.NetworkSettings.Networks[$scope.config.HostConfig.NetworkMode].IPAMConfig.IPv6Address) {
            $scope.formValues.IPv6 = d.NetworkSettings.Networks[$scope.config.HostConfig.NetworkMode].IPAMConfig.IPv6Address;
          }
        }
      }
      $scope.config.NetworkingConfig.EndpointsConfig[$scope.config.HostConfig.NetworkMode] = d.NetworkSettings.Networks[$scope.config.HostConfig.NetworkMode];
      if (Object.keys(d.NetworkSettings.Networks).length > 1) {
        var firstNetwork = d.NetworkSettings.Networks[Object.keys(d.NetworkSettings.Networks)[0]];
        $scope.config.NetworkingConfig.EndpointsConfig[$scope.config.HostConfig.NetworkMode] = firstNetwork;
        $scope.extraNetworks = angular.copy(d.NetworkSettings.Networks);
        delete $scope.extraNetworks[Object.keys(d.NetworkSettings.Networks)[0]];
      }
      $scope.formValues.MacAddress = d.Config.MacAddress;

      if (d.HostConfig.Dns && d.HostConfig.Dns[0]) {
        $scope.formValues.DnsPrimary = d.HostConfig.Dns[0];
        if (d.HostConfig.Dns[1]) {
          $scope.formValues.DnsSecondary = d.HostConfig.Dns[1];
        }
      }

      // ExtraHosts
      if ($scope.config.HostConfig.ExtraHosts) {
        var extraHosts = $scope.config.HostConfig.ExtraHosts;
        for (var i = 0; i < extraHosts.length; i++) {
          var host = extraHosts[i];
          $scope.formValues.ExtraHosts.push({ value: host });
        }
        $scope.config.HostConfig.ExtraHosts = [];
      }
    }

    function loadFromContainerEnvironmentVariables() {
      var envArr = [];
      for (var e in $scope.config.Env) {
        if ({}.hasOwnProperty.call($scope.config.Env, e)) {
          var arr = $scope.config.Env[e].split(/\=(.*)/);
          envArr.push({ name: arr[0], value: arr[1] });
        }
      }
      $scope.config.Env = envArr;
    }

    function loadFromContainerLabels() {
      for (var l in $scope.config.Labels) {
        if ({}.hasOwnProperty.call($scope.config.Labels, l)) {
          $scope.formValues.Labels.push({ name: l, value: $scope.config.Labels[l] });
        }
      }
    }

    function loadFromContainerConsole() {
      if ($scope.config.OpenStdin && $scope.config.Tty) {
        $scope.formValues.Console = 'both';
      } else if (!$scope.config.OpenStdin && $scope.config.Tty) {
        $scope.formValues.Console = 'tty';
      } else if ($scope.config.OpenStdin && !$scope.config.Tty) {
        $scope.formValues.Console = 'interactive';
      } else if (!$scope.config.OpenStdin && !$scope.config.Tty) {
        $scope.formValues.Console = 'none';
      }
    }

    function loadFromContainerDevices() {
      var path = [];
      for (var dev in $scope.config.HostConfig.Devices) {
        if ({}.hasOwnProperty.call($scope.config.HostConfig.Devices, dev)) {
          var device = $scope.config.HostConfig.Devices[dev];
          path.push({ pathOnHost: device.PathOnHost, pathInContainer: device.PathInContainer });
        }
      }
      $scope.config.HostConfig.Devices = path;
    }

    function loadFromContainerImageConfig() {
      RegistryService.retrievePorRegistryModelFromRepository($scope.config.Image)
        .then((model) => {
          $scope.formValues.RegistryModel = model;
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to retrive registry');
        });
    }

    function loadFromContainerResources(d) {
      if (d.HostConfig.NanoCpus) {
        $scope.formValues.CpuLimit = d.HostConfig.NanoCpus / 1000000000;
      }
      if (d.HostConfig.Memory) {
        $scope.formValues.MemoryLimit = d.HostConfig.Memory / 1024 / 1024;
      }
      if (d.HostConfig.MemoryReservation) {
        $scope.formValues.MemoryReservation = d.HostConfig.MemoryReservation / 1024 / 1024;
      }
    }

    function loadFromContainerCapabilities(d) {
      if (d.HostConfig.CapAdd) {
        d.HostConfig.CapAdd.forEach(function (cap) {
          $scope.formValues.capabilities.push(new ContainerCapability(cap, true));
        });
      }
      if (d.HostConfig.CapDrop) {
        d.HostConfig.CapDrop.forEach(function (cap) {
          $scope.formValues.capabilities.push(new ContainerCapability(cap, false));
        });
      }

      function hasCapability(item) {
        return item.capability === cap.capability;
      }

      var capabilities = new ContainerCapabilities();
      for (var i = 0; i < capabilities.length; i++) {
        var cap = capabilities[i];
        if (!_.find($scope.formValues.capabilities, hasCapability)) {
          $scope.formValues.capabilities.push(cap);
        }
      }

      $scope.formValues.capabilities.sort(function (a, b) {
        return a.capability < b.capability ? -1 : 1;
      });
    }

    // Convert healthcheck to human-readable
    function loadConfigHealthcheck() {
      if ($scope.config.Healthcheck.Test) {
        $scope.formValues.Healthcheck.Type = $scope.config.Healthcheck.Test[0];
        if ($scope.config.Healthcheck.Test.length > 1) {
          $scope.formValues.Healthcheck.Test = $scope.config.Healthcheck.Test[1];
        }

        $scope.formValues.Healthcheck.Retries = $scope.config.Healthcheck.Retries;

        // Docker uses nanos
        if (!isNaN($scope.config.Healthcheck.StartPeriod)) {
          $scope.formValues.Healthcheck.StartPeriod = simpleDuration.stringify($scope.config.Healthcheck.StartPeriod / 1e9, 'ms');
        }
        if (!isNaN($scope.config.Healthcheck.Interval)) {
          $scope.formValues.Healthcheck.Interval = simpleDuration.stringify($scope.config.Healthcheck.Interval / 1e9, 'ms');
        }
        if (!isNaN($scope.config.Healthcheck.Timeout)) {
          $scope.formValues.Healthcheck.Timeout = simpleDuration.stringify($scope.config.Healthcheck.Timeout / 1e9, 'ms');
        }
      }
    }

    function loadFromContainerLogging(config) {
      var logConfig = config.HostConfig.LogConfig;
      $scope.formValues.LogDriverName = logConfig.Type;
      $scope.formValues.LogDriverOpts = _.map(logConfig.Config, function (value, name) {
        return {
          name: name,
          value: value,
        };
      });
    }

    function loadFromContainerSpec() {
      // Get container
      Container.get({ id: $transition$.params().from })
        .$promise.then(function success(d) {
          var fromContainer = new ContainerDetailsViewModel(d);
          if (fromContainer.ResourceControl && fromContainer.ResourceControl.Public) {
            $scope.formValues.AccessControlData.AccessControlEnabled = false;
          }
          $scope.fromContainer = fromContainer;
          $scope.state.mode = 'duplicate';
          $scope.config = ContainerHelper.configFromContainer(fromContainer.Model);
          loadFromContainerCmd(d);
          loadFromContainerEntrypoint(d);
          loadFromContainerLogging(d);
          loadFromContainerPortBindings(d);
          loadFromContainerVolumes(d);
          loadFromContainerNetworkConfig(d);
          loadFromContainerEnvironmentVariables(d);
          loadFromContainerLabels(d);
          loadFromContainerConsole(d);
          loadFromContainerDevices(d);
          loadFromContainerImageConfig(d);
          loadFromContainerResources(d);
          loadFromContainerCapabilities(d);
          loadConfigHealthcheck();
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to retrieve container');
        });
    }

    async function initView() {
      var nodeName = $transition$.params().nodeName;
      $scope.formValues.NodeName = nodeName;
      HttpRequestHelper.setPortainerAgentTargetHeader(nodeName);

      $scope.isAdmin = Authentication.isAdmin();
      $scope.showDeviceMapping = await shouldShowDevices();
      $scope.areContainerCapabilitiesEnabled = await checkIfContainerCapabilitiesEnabled();
      $scope.isAdminOrEndpointAdmin = Authentication.isAdmin();

      Volume.query(
        {},
        function (d) {
          $scope.availableVolumes = d.Volumes.sort((vol1, vol2) => {
            return vol1.Name.localeCompare(vol2.Name);
          });
        },
        function (e) {
          Notifications.error('Failure', e, 'Unable to retrieve volumes');
        }
      );

      var provider = $scope.applicationState.endpoint.mode.provider;
      var apiVersion = $scope.applicationState.endpoint.apiVersion;
      NetworkService.networks(provider === 'DOCKER_STANDALONE' || provider === 'DOCKER_SWARM_MODE', false, provider === 'DOCKER_SWARM_MODE' && apiVersion >= 1.25)
        .then(function success(networks) {
          networks.push({ Name: 'container' });
          $scope.availableNetworks = networks.sort((a, b) => a.Name.localeCompare(b.Name));

          if (_.find(networks, { Name: 'nat' })) {
            $scope.config.HostConfig.NetworkMode = 'nat';
          }
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to retrieve networks');
        });

      Container.query(
        {},
        function (d) {
          var containers = d;
          $scope.runningContainers = containers;
          if ($transition$.params().from) {
            loadFromContainerSpec();
          } else {
            $scope.fromContainer = {};
            $scope.formValues.capabilities = $scope.areContainerCapabilitiesEnabled ? new ContainerCapabilities() : [];
          }
        },
        function (e) {
          Notifications.error('Failure', e, 'Unable to retrieve running containers');
        }
      );

      SystemService.info()
        .then(function success(data) {
          $scope.availableRuntimes = data.Runtimes ? Object.keys(data.Runtimes) : [];
          $scope.state.sliderMaxCpu = 32;
          if (data.NCPU) {
            $scope.state.sliderMaxCpu = data.NCPU;
          }
          $scope.state.sliderMaxMemory = 32768;
          if (data.MemTotal) {
            $scope.state.sliderMaxMemory = Math.floor(data.MemTotal / 1000 / 1000);
          }
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to retrieve engine details');
        });

      SettingsService.publicSettings()
        .then(function success(data) {
          $scope.allowBindMounts = $scope.isAdminOrEndpointAdmin || data.AllowBindMountsForRegularUsers;
          $scope.allowPrivilegedMode = data.AllowPrivilegedModeForRegularUsers;
        })
        .catch(function error(err) {
          Notifications.error('Failure', err, 'Unable to retrieve application settings');
        });

      PluginService.loggingPlugins(apiVersion < 1.25).then(function success(loggingDrivers) {
        $scope.availableLoggingDrivers = loggingDrivers;
      });
    }

    function validateForm(accessControlData, isAdmin) {
      $scope.state.formValidationError = '';
      var error = '';
      error = FormValidator.validateAccessControl(accessControlData, isAdmin);

      if (error) {
        $scope.state.formValidationError = error;
        return false;
      }
      return true;
    }

    function create() {
      var oldContainer = null;
      HttpRequestHelper.setPortainerAgentTargetHeader($scope.formValues.NodeName);
      return findCurrentContainer().then(setOldContainer).then(confirmCreateContainer).then(startCreationProcess).catch(notifyOnError).finally(final);

      function final() {
        $scope.state.actionInProgress = false;
      }

      function setOldContainer(container) {
        oldContainer = container;
        return container;
      }

      function findCurrentContainer() {
        return Container.query({ all: 1, filters: { name: ['^/' + $scope.config.name + '$'] } })
          .$promise.then(function onQuerySuccess(containers) {
            if (!containers.length) {
              return;
            }
            return containers[0];
          })
          .catch(notifyOnError);

        function notifyOnError(err) {
          Notifications.error('Failure', err, 'Unable to retrieve containers');
        }
      }

      function startCreationProcess(confirmed) {
        if (!confirmed) {
          return $q.when();
        }
        if (!validateAccessControl()) {
          return $q.when();
        }
        $scope.state.actionInProgress = true;
        return pullImageIfNeeded()
          .then(stopAndRenameContainer)
          .then(createNewContainer)
          .then(applyResourceControl)
          .then(connectToExtraNetworks)
          .then(removeOldContainer)
          .then(onSuccess)
          .catch(onCreationProcessFail);
      }

      function onCreationProcessFail(error) {
        var deferred = $q.defer();
        removeNewContainer()
          .then(restoreOldContainerName)
          .then(function () {
            deferred.reject(error);
          })
          .catch(function (restoreError) {
            deferred.reject(restoreError);
          });
        return deferred.promise;
      }

      function removeNewContainer() {
        return findCurrentContainer().then(function onContainerLoaded(container) {
          if (container && (!oldContainer || container.Id !== oldContainer.Id)) {
            return ContainerService.remove(container, true);
          }
        });
      }

      function restoreOldContainerName() {
        if (!oldContainer) {
          return;
        }
        return ContainerService.renameContainer(oldContainer.Id, oldContainer.Names[0].substring(1));
      }

      function confirmCreateContainer(container) {
        if (!container) {
          return $q.when(true);
        }

        return showConfirmationModal();

        function showConfirmationModal() {
          var deferred = $q.defer();

          ModalService.confirm({
            title: 'Are you sure ?',
            message: 'A container with the same name already exists. Portainer can automatically remove it and re-create one. Do you want to replace it?',
            buttons: {
              confirm: {
                label: 'Replace',
                className: 'btn-danger',
              },
            },
            callback: function onConfirm(confirmed) {
              deferred.resolve(confirmed);
            },
          });

          return deferred.promise;
        }
      }

      function stopAndRenameContainer() {
        if (!oldContainer) {
          return $q.when();
        }
        return stopContainerIfNeeded(oldContainer).then(renameContainer);
      }

      function stopContainerIfNeeded(oldContainer) {
        if (oldContainer.State !== 'running') {
          return $q.when();
        }
        return ContainerService.stopContainer(oldContainer.Id);
      }

      function renameContainer() {
        return ContainerService.renameContainer(oldContainer.Id, oldContainer.Names[0].substring(1) + '-old');
      }

      function pullImageIfNeeded() {
        return $q.when($scope.formValues.alwaysPull && ImageService.pullImage($scope.formValues.RegistryModel, true));
      }

      function createNewContainer() {
        return $async(async () => {
          const config = prepareConfiguration();
          return await ContainerService.createAndStartContainer(config);
        });
      }

      function applyResourceControl(newContainer) {
        const userId = Authentication.getUserDetails().ID;
        const resourceControl = newContainer.Portainer.ResourceControl;
        const containerId = newContainer.Id;
        const accessControlData = $scope.formValues.AccessControlData;

        return ResourceControlService.applyResourceControl(userId, accessControlData, resourceControl).then(function onApplyResourceControlSuccess() {
          return containerId;
        });
      }

      function connectToExtraNetworks(newContainerId) {
        if (!$scope.extraNetworks) {
          return $q.when();
        }

        var connectionPromises = _.forOwn($scope.extraNetworks, function (network, networkName) {
          if (_.has(network, 'Aliases')) {
            var aliases = _.filter(network.Aliases, (o) => {
              return !_.startsWith($scope.fromContainer.Id, o);
            });
          }
          return NetworkService.connectContainer(networkName, newContainerId, aliases);
        });

        return $q.all(connectionPromises);
      }

      function removeOldContainer() {
        var deferred = $q.defer();

        if (!oldContainer) {
          deferred.resolve();
          return;
        }

        ContainerService.remove(oldContainer, true).then(notifyOnRemoval).catch(notifyOnRemoveError);

        return deferred.promise;

        function notifyOnRemoval() {
          Notifications.success('Container Removed', oldContainer.Id);
          deferred.resolve();
        }

        function notifyOnRemoveError(err) {
          deferred.reject({ msg: 'Unable to remove container', err: err });
        }
      }

      function notifyOnError(err) {
        Notifications.error('Failure', err, 'Unable to create container');
      }

      function validateAccessControl() {
        var accessControlData = $scope.formValues.AccessControlData;
        return validateForm(accessControlData, $scope.isAdmin);
      }

      function onSuccess() {
        Notifications.success('Container successfully created');
        $state.go('docker.containers', {}, { reload: true });
      }
    }

    async function shouldShowDevices() {
      const { allowDeviceMappingForRegularUsers } = $scope.applicationState.application;

      return allowDeviceMappingForRegularUsers || Authentication.isAdmin();
    }

    async function checkIfContainerCapabilitiesEnabled() {
      const { allowContainerCapabilitiesForRegularUsers } = $scope.applicationState.application;

      return allowContainerCapabilitiesForRegularUsers || Authentication.isAdmin();
    }

    initView();
  },
]);

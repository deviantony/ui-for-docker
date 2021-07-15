import _ from 'lodash-es';
import uuidv4 from 'uuid/v4';
import { PortainerEndpointTypes } from 'Portainer/models/endpoint/models';
import { EndpointSecurityFormData } from '../../../components/endpointSecurity/porEndpointSecurityModel';

angular.module('portainer.app').controller('EndpointController', EndpointController);

/* @ngInject */
function EndpointController(
  $async,
  $q,
  $scope,
  $state,
  $transition$,
  $filter,
  clipboard,
  EndpointService,
  GroupService,
  TagService,
  EndpointProvider,
  Notifications,
  Authentication,
  SettingsService
) {
  const DEPLOYMENT_TABS = {
    SWARM: 'swarm',
    STANDALONE: 'standalone',
    KUBERNETES: 'kubernetes',
  };

  const PLATFORM_TYPES = {
    WINDOWS: 'windows',
    LINUX: 'linux',
  };

  $scope.state = {
    uploadInProgress: false,
    actionInProgress: false,
    deploymentTab: DEPLOYMENT_TABS.KUBERNETES,
    platformType: PLATFORM_TYPES.LINUX,
    azureEndpoint: false,
    kubernetesEndpoint: false,
    agentEndpoint: false,
    edgeEndpoint: false,
    allowCreate: Authentication.isAdmin(),
    availableEdgeAgentCheckinOptions: [
      { key: 'Use default interval', value: 0 },
      {
        key: '5 seconds',
        value: 5,
      },
      {
        key: '10 seconds',
        value: 10,
      },
      {
        key: '30 seconds',
        value: 30,
      },
      { key: '5 minutes', value: 300 },
      { key: '1 hour', value: 3600 },
      { key: '1 day', value: 86400 },
    ],
    allowSelfSignedCerts: true,
  };

  $scope.dockerCommands = {
    [DEPLOYMENT_TABS.STANDALONE]: {
      [PLATFORM_TYPES.LINUX]: buildLinuxStandaloneCommand,
      [PLATFORM_TYPES.WINDOWS]: buildWindowsStandaloneCommand,
    },
    [DEPLOYMENT_TABS.SWARM]: {
      [PLATFORM_TYPES.LINUX]: buildLinuxSwarmCommand,
      [PLATFORM_TYPES.WINDOWS]: buildWindowsSwarmCommand,
    },
    [DEPLOYMENT_TABS.KUBERNETES]: {
      [PLATFORM_TYPES.LINUX]: buildKubernetesCommand,
      [PLATFORM_TYPES.WINDOWS]: () => '',
    },
  };

  $scope.formValues = {
    SecurityFormData: new EndpointSecurityFormData(),
  };

  $scope.copyEdgeAgentDeploymentCommand = copyEdgeAgentDeploymentCommand;
  function copyEdgeAgentDeploymentCommand() {
    const command = $scope.dockerCommands[$scope.state.deploymentTab][$scope.state.platformType]($scope.randomEdgeID, $scope.endpoint.EdgeKey, $scope.state.allowSelfSignedCerts);
    clipboard.copyText(command.trim());
    $('#copyNotificationDeploymentCommand').show().fadeOut(2500);
  }

  $scope.copyEdgeAgentKey = function () {
    clipboard.copyText($scope.endpoint.EdgeKey);
    $('#copyNotificationEdgeKey').show().fadeOut(2500);
  };

  $scope.onCreateTag = function onCreateTag(tagName) {
    return $async(onCreateTagAsync, tagName);
  };

  async function onCreateTagAsync(tagName) {
    try {
      const tag = await TagService.createTag(tagName);
      $scope.availableTags = $scope.availableTags.concat(tag);
      $scope.endpoint.TagIds = $scope.endpoint.TagIds.concat(tag.Id);
    } catch (err) {
      Notifications.error('Failue', err, 'Unable to create tag');
    }
  }

  $scope.updateEndpoint = function () {
    var endpoint = $scope.endpoint;
    var securityData = $scope.formValues.SecurityFormData;
    var TLS = securityData.TLS;
    var TLSMode = securityData.TLSMode;
    var TLSSkipVerify = TLS && (TLSMode === 'tls_client_noca' || TLSMode === 'tls_only');
    var TLSSkipClientVerify = TLS && (TLSMode === 'tls_ca' || TLSMode === 'tls_only');

    var payload = {
      Name: endpoint.Name,
      PublicURL: endpoint.PublicURL,
      GroupID: endpoint.GroupId,
      TagIds: endpoint.TagIds,
      EdgeCheckinInterval: endpoint.EdgeCheckinInterval,
      TLS: TLS,
      TLSSkipVerify: TLSSkipVerify,
      TLSSkipClientVerify: TLSSkipClientVerify,
      TLSCACert: TLSSkipVerify || securityData.TLSCACert === endpoint.TLSConfig.TLSCACert ? null : securityData.TLSCACert,
      TLSCert: TLSSkipClientVerify || securityData.TLSCert === endpoint.TLSConfig.TLSCert ? null : securityData.TLSCert,
      TLSKey: TLSSkipClientVerify || securityData.TLSKey === endpoint.TLSConfig.TLSKey ? null : securityData.TLSKey,
      AzureApplicationID: endpoint.AzureCredentials.ApplicationID,
      AzureTenantID: endpoint.AzureCredentials.TenantID,
      AzureAuthenticationKey: endpoint.AzureCredentials.AuthenticationKey,
    };

    if (
      $scope.endpointType !== 'local' &&
      endpoint.Type !== PortainerEndpointTypes.AzureEnvironment &&
      endpoint.Type !== PortainerEndpointTypes.KubernetesLocalEnvironment &&
      endpoint.Type !== PortainerEndpointTypes.AgentOnKubernetesEnvironment
    ) {
      payload.URL = 'tcp://' + endpoint.URL;
    }

    if (endpoint.Type === PortainerEndpointTypes.AgentOnKubernetesEnvironment || endpoint.Type === PortainerEndpointTypes.KubernetesLocalEnvironment) {
      payload.URL = endpoint.URL;
    }

    $scope.state.actionInProgress = true;
    EndpointService.updateEndpoint(endpoint.Id, payload).then(
      function success() {
        Notifications.success('Endpoint updated', $scope.endpoint.Name);
        EndpointProvider.setEndpointPublicURL(endpoint.PublicURL);
        $state.go('portainer.endpoints', {}, { reload: true });
      },
      function error(err) {
        Notifications.error('Failure', err, 'Unable to update endpoint');
        $scope.state.actionInProgress = false;
      },
      function update(evt) {
        if (evt.upload) {
          $scope.state.uploadInProgress = evt.upload;
        }
      }
    );
  };

  function decodeEdgeKey(key) {
    let keyInformation = {};

    if (key === '') {
      return keyInformation;
    }

    let decodedKey = _.split(atob(key), '|');
    keyInformation.instanceURL = decodedKey[0];
    keyInformation.tunnelServerAddr = decodedKey[1];

    return keyInformation;
  }

  function configureState() {
    if (
      $scope.endpoint.Type === PortainerEndpointTypes.KubernetesLocalEnvironment ||
      $scope.endpoint.Type === PortainerEndpointTypes.AgentOnKubernetesEnvironment ||
      $scope.endpoint.Type === PortainerEndpointTypes.EdgeAgentOnKubernetesEnvironment
    ) {
      $scope.state.kubernetesEndpoint = true;
    }
    if ($scope.endpoint.Type === PortainerEndpointTypes.EdgeAgentOnDockerEnvironment || $scope.endpoint.Type === PortainerEndpointTypes.EdgeAgentOnKubernetesEnvironment) {
      $scope.state.edgeEndpoint = true;
    }
    if ($scope.endpoint.Type === PortainerEndpointTypes.AzureEnvironment) {
      $scope.state.azureEndpoint = true;
    }
    if (
      $scope.endpoint.Type === PortainerEndpointTypes.AgentOnDockerEnvironment ||
      $scope.endpoint.Type === PortainerEndpointTypes.EdgeAgentOnDockerEnvironment ||
      $scope.endpoint.Type === PortainerEndpointTypes.AgentOnKubernetesEnvironment ||
      $scope.endpoint.Type === PortainerEndpointTypes.EdgeAgentOnKubernetesEnvironment
    ) {
      $scope.state.agentEndpoint = true;
    }
  }

  async function initView() {
    return $async(async () => {
      try {
        const [endpoint, groups, tags, settings] = await Promise.all([
          EndpointService.endpoint($transition$.params().id),
          GroupService.groups(),
          TagService.tags(),
          SettingsService.settings(),
        ]);

        if (endpoint.URL.indexOf('unix://') === 0 || endpoint.URL.indexOf('npipe://') === 0) {
          $scope.endpointType = 'local';
        } else {
          $scope.endpointType = 'remote';
        }

        endpoint.URL = $filter('stripprotocol')(endpoint.URL);

        if (endpoint.Type === PortainerEndpointTypes.EdgeAgentOnDockerEnvironment || endpoint.Type === PortainerEndpointTypes.EdgeAgentOnKubernetesEnvironment) {
          $scope.edgeKeyDetails = decodeEdgeKey(endpoint.EdgeKey);
          $scope.randomEdgeID = uuidv4();

          $scope.state.availableEdgeAgentCheckinOptions[0].key += ` (${settings.EdgeAgentCheckinInterval} seconds)`;
        }

        $scope.endpoint = endpoint;
        $scope.groups = groups;
        $scope.availableTags = tags;

        configureState();
      } catch (err) {
        Notifications.error('Failure', err, 'Unable to retrieve endpoint details');
      }
    });
  }

  function buildLinuxStandaloneCommand(edgeId, edgeKey, allowSelfSignedCerts) {
    return `
docker run -d \\
    -v /var/run/docker.sock:/var/run/docker.sock \\
    -v /var/lib/docker/volumes:/var/lib/docker/volumes \\
    -v /:/host \\
    -v portainer_agent_data:/data \\
    --restart always \\
    -e EDGE=1 \\
    -e EDGE_ID=${edgeId} \\
    -e EDGE_KEY=${edgeKey} \\
    -e CAP_HOST_MANAGEMENT=1 \\
    -e EDGE_INSECURE_POLL=${allowSelfSignedCerts ? 1 : 0} \\
    --name portainer_edge_agent \\
    portainer/agent:2.4.0`;
  }

  function buildWindowsStandaloneCommand(edgeId, edgeKey, allowSelfSignedCerts) {
    return `
docker run -d \\
  --mount type=npipe,src=\\\\.\\pipe\\docker_engine,dst=\\\\.\\pipe\\docker_engine \\
  --mount type=bind,src=C:\\ProgramData\\docker\\volumes,dst=C:\\ProgramData\\docker\\volumes \\
  --mount type=volume,src=portainer_agent_data,dst=C:\\data \\
  --restart always \\
  -e EDGE=1 \\
  -e EDGE_ID=${edgeId} \\
  -e EDGE_KEY=${edgeKey} \\
  -e CAP_HOST_MANAGEMENT=1 \\
  -e EDGE_INSECURE_POLL=${allowSelfSignedCerts ? 1 : 0} \\
  --name portainer_edge_agent \\
  portainer/agent`;
  }

  function buildLinuxSwarmCommand(edgeId, edgeKey, allowSelfSignedCerts) {
    return `
docker network create \\
  --driver overlay \\
  portainer_agent_network;

docker service create \\
  --name portainer_edge_agent \\
  --network portainer_agent_network \\
  -e AGENT_CLUSTER_ADDR=tasks.portainer_edge_agent \\
  -e EDGE=1 \\
  -e EDGE_ID=${edgeId} \\
  -e EDGE_KEY=${edgeKey} \\
  -e CAP_HOST_MANAGEMENT=1 \\
  -e EDGE_INSECURE_POLL=${allowSelfSignedCerts ? 1 : 0} \\
  --mode global \\
  --constraint 'node.platform.os == linux' \\
  --mount type=bind,src=//var/run/docker.sock,dst=/var/run/docker.sock \\
  --mount type=bind,src=//var/lib/docker/volumes,dst=/var/lib/docker/volumes \\
  --mount type=bind,src=//,dst=/host \\
  --mount type=volume,src=portainer_agent_data,dst=/data \\
  portainer/agent`;
  }

  function buildWindowsSwarmCommand(edgeId, edgeKey, allowSelfSignedCerts) {
    return `
docker network create \\
  --driver overlay \\
  portainer_edge_agent_network && \\
docker service create \\
  --name portainer_edge_agent \\
  --network portainer_edge_agent_network \\
  -e AGENT_CLUSTER_ADDR=tasks.portainer_edge_agent \\
  -e EDGE=1 \\
  -e EDGE_ID=${edgeId} \\
  -e EDGE_KEY=${edgeKey} \\
  -e CAP_HOST_MANAGEMENT=1 \\
  -e EDGE_INSECURE_POLL=${allowSelfSignedCerts ? 1 : 0} \\
  --mode global \\
  --constraint node.platform.os==windows \\
  --mount type=npipe,src=\\\\.\\pipe\\docker_engine,dst=\\\\.\\pipe\\docker_engine \\
  --mount type=bind,src=C:\\ProgramData\\docker\\volumes,dst=C:\\ProgramData\\docker\\volumes \\
  --mount type=volume,src=portainer_agent_data,dst=C:\\data \\
  portainer/agent`;
  }

  function buildKubernetesCommand(edgeId, edgeKey, allowSelfSignedCerts) {
    return `
curl https://downloads.portainer.io/portainer-ce29-edge-agent-setup.sh | bash -s -- ${edgeId} ${edgeKey} ${allowSelfSignedCerts ? '1' : ''}
`;
  }

  initView();
}

angular.module('portainer.docker').controller('SwarmNodeDetailsPanelController', [
  'NodeService',
  'LabelHelper',
  'Notifications',
  '$state',
  function SwarmNodeDetailsPanelController(NodeService, LabelHelper, Notifications, $state) {
    var ctrl = this;
    ctrl.state = {
      managerAddress: '',
      hasChanges: false,
    };
    ctrl.$onChanges = $onChanges;
    ctrl.addLabel = addLabel;
    ctrl.updateNodeLabels = updateNodeLabels;
    ctrl.updateNodeAvailability = updateNodeAvailability;
    ctrl.saveChanges = saveChanges;
    ctrl.cancelChanges = cancelChanges;

    var managerRole = 'manager';

    function $onChanges() {
      if (!ctrl.details) {
        return;
      }
      if (ctrl.details.role === managerRole) {
        ctrl.state.managerAddress = '(' + ctrl.details.managerAddress + ')';
      }
    }

    function addLabel() {
      ctrl.details.nodeLabels.push({
        key: '',
        value: '',
        originalValue: '',
        originalKey: '',
      });
    }

    function updateNodeLabels(labels) {
      ctrl.details.nodeLabels = labels;
      ctrl.state.hasChanges = true;
    }

    function updateNodeAvailability(availability) {
      ctrl.details.availability = availability;
      ctrl.state.hasChanges = true;
    }

    function saveChanges() {
      var originalNode = ctrl.originalNode;
      var config = {
        Name: originalNode.Name,
        Availability: ctrl.details.availability,
        Role: originalNode.Role,
        Labels: LabelHelper.fromKeyValueToLabelHash(ctrl.details.nodeLabels),
        Id: originalNode.Id,
        Version: originalNode.Version,
      };

      NodeService.updateNode(config).then(onUpdateSuccess).catch(notifyOnError);

      function onUpdateSuccess(node) {
        Notifications.success('Node successfully updated', 'Node updated');

        if (node.Availability === 'drain') {
          $state.go('docker.swarm', { reload: true });
        } else {
          $state.go('docker.nodes.node', { id: originalNode.Id }, { reload: true });
        }
      }

      function notifyOnError(error) {
        Notifications.error('Failure', error, 'Failed to update node');
      }
    }

    function cancelChanges() {
      cancelLabelChanges();
      ctrl.details.availability = ctrl.originalNode.Availability;
      ctrl.state.hasChanges = false;
    }

    function cancelLabelChanges() {
      ctrl.details.nodeLabels = ctrl.details.nodeLabels
        .filter(function (label) {
          return label.originalValue || label.originalKey;
        })
        .map(function (label) {
          return Object.assign(label, {
            value: label.originalValue,
            key: label.originalKey,
          });
        });
    }
  },
]);

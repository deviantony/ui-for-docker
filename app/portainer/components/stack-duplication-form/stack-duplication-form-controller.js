angular.module('portainer.app').controller('StackDuplicationFormController', [
  'Notifications',
  function StackDuplicationFormController(Notifications) {
    var ctrl = this;

    ctrl.state = {
      duplicationInProgress: false,
      migrationInProgress: false,
    };

    ctrl.formValues = {
      endpoint: null,
      newName: '',
    };

    ctrl.isFormValidForDuplication = isFormValidForDuplication;
    ctrl.isFormValidForMigration = isFormValidForMigration;
    ctrl.duplicateStack = duplicateStack;
    ctrl.migrateStack = migrateStack;
    ctrl.isMigrationButtonDisabled = isMigrationButtonDisabled;
    ctrl.isEndpointSelected = isEndpointSelected;

    function isFormValidForMigration() {
      return ctrl.formValues.endpoint && ctrl.formValues.endpoint.Id;
    }

    function isFormValidForDuplication() {
      return isFormValidForMigration() && ctrl.formValues.newName && !ctrl.yamlError;
    }

    function duplicateStack() {
      if (!ctrl.formValues.newName) {
        Notifications.error('Failure', null, 'Stack name is required for duplication');
        return;
      }
      ctrl.state.duplicationInProgress = true;
      ctrl
        .onDuplicate({
          endpointId: ctrl.formValues.endpoint.Id,
          name: ctrl.formValues.newName ? ctrl.formValues.newName : undefined,
        })
        .finally(function () {
          ctrl.state.duplicationInProgress = false;
        });
    }

    function migrateStack() {
      ctrl.state.migrationInProgress = true;
      ctrl
        .onMigrate({
          endpointId: ctrl.formValues.endpoint.Id,
          name: ctrl.formValues.newName ? ctrl.formValues.newName : undefined,
        })
        .finally(function () {
          ctrl.state.migrationInProgress = false;
        });
    }

    function isMigrationButtonDisabled() {
      return !ctrl.isFormValidForMigration() || ctrl.state.duplicationInProgress || ctrl.state.migrationInProgress || isTargetEndpointAndCurrentEquals();
    }

    function isTargetEndpointAndCurrentEquals() {
      return ctrl.formValues.endpoint && ctrl.formValues.endpoint.Id === ctrl.currentEndpointId;
    }

    function isEndpointSelected() {
      return ctrl.formValues.endpoint && ctrl.formValues.endpoint.Id;
    }
  },
]);

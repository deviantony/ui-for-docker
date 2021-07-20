angular.module('portainer.docker').controller('BuildImageController', BuildImageController);

function BuildImageController($scope, $async, $window, ModalService, BuildService, Notifications, HttpRequestHelper) {
  $scope.state = {
    BuildType: 'editor',
    actionInProgress: false,
    activeTab: 0,
    isEditorDirty: false,
  };

  $scope.formValues = {
    ImageNames: [{ Name: '' }],
    UploadFile: null,
    DockerFileContent: '',
    URL: '',
    Path: 'Dockerfile',
    NodeName: null,
  };

  $window.onbeforeunload = () => {
    if ($scope.state.BuildType === 'editor' && $scope.formValues.DockerFileContent && $scope.state.isEditorDirty) {
      return '';
    }
  };

  $scope.$on('$destroy', function () {
    $scope.state.isEditorDirty = false;
  });

  $scope.addImageName = function () {
    $scope.formValues.ImageNames.push({ Name: '' });
  };

  $scope.removeImageName = function (index) {
    $scope.formValues.ImageNames.splice(index, 1);
  };

  function buildImageBasedOnBuildType(method, names) {
    var buildType = $scope.state.BuildType;
    var dockerfilePath = $scope.formValues.Path;

    if (buildType === 'upload') {
      var file = $scope.formValues.UploadFile;
      return BuildService.buildImageFromUpload(names, file, dockerfilePath);
    } else if (buildType === 'url') {
      var URL = $scope.formValues.URL;
      return BuildService.buildImageFromURL(names, URL, dockerfilePath);
    } else {
      var dockerfileContent = $scope.formValues.DockerFileContent;
      return BuildService.buildImageFromDockerfileContent(names, dockerfileContent);
    }
  }

  $scope.buildImage = buildImage;

  async function buildImage() {
    return $async(async () => {
      var buildType = $scope.state.BuildType;

      if (buildType === 'editor' && $scope.formValues.DockerFileContent === '') {
        $scope.state.formValidationError = 'Dockerfile content must not be empty';
        return;
      }

      $scope.state.actionInProgress = true;

      var imageNames = $scope.formValues.ImageNames.filter(function filterNull(x) {
        return x.Name;
      }).map(function getNames(x) {
        return x.Name;
      });

      var nodeName = $scope.formValues.NodeName;
      HttpRequestHelper.setPortainerAgentTargetHeader(nodeName);

      try {
        const data = await buildImageBasedOnBuildType(buildType, imageNames);
        $scope.buildLogs = data.buildLogs;
        $scope.state.activeTab = 1;
        if (data.hasError) {
          Notifications.error('An error occurred during build', { msg: 'Please check build logs output' });
        } else {
          Notifications.success('Image successfully built');
          $scope.state.isEditorDirty = false;
        }
      } catch (err) {
        Notifications.error('Failure', err, 'Unable to build image');
      } finally {
        $scope.state.actionInProgress = false;
      }
    });
  }

  $scope.validImageNames = function () {
    for (var i = 0; i < $scope.formValues.ImageNames.length; i++) {
      var item = $scope.formValues.ImageNames[i];
      if (item.Name !== '') {
        return true;
      }
    }
    return false;
  };

  $scope.editorUpdate = function (cm) {
    $scope.formValues.DockerFileContent = cm.getValue();
    $scope.state.isEditorDirty = true;
  };

  this.uiCanExit = async function () {
    if ($scope.state.BuildType === 'editor' && $scope.formValues.DockerFileContent && $scope.state.isEditorDirty) {
      return ModalService.confirmWebEditorDiscard();
    }
  };
}

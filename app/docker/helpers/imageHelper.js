import _ from 'lodash-es';
import { RegistryTypes } from 'Portainer/models/registryTypes';

angular.module('portainer.docker').factory('ImageHelper', ImageHelperFactory);
function ImageHelperFactory() {
  return {
    isValidTag,
    createImageConfigForContainer,
    getImagesNamesForDownload,
    removeDigestFromRepository,
    imageContainsURL,
  };

  function isValidTag(tag) {
    return tag.match(/^(?![\.\-])([a-zA-Z0-9\_\.\-])+$/g);
  }

  function getImagesNamesForDownload(images) {
    var names = images.map(function (image) {
      return image.RepoTags[0] !== '<none>:<none>' ? image.RepoTags[0] : image.Id;
    });
    return {
      names: names,
    };
  }

  /**
   *
   * @param {PorImageRegistryModel} registry
   */
  function createImageConfigForContainer(imageModel) {
    return {
      fromImage: buildImageFullURI(imageModel),
    };
  }

  function imageContainsURL(image) {
    const split = _.split(image, '/');
    const url = split[0];
    if (split.length > 1) {
      return _.includes(url, '.') || _.includes(url, ':');
    }
    return false;
  }

  function removeDigestFromRepository(repository) {
    return repository.split('@sha')[0];
  }
}
/**
 * builds the complete uri for an image based on its registry
 * @param {PorImageRegistryModel} imageModel
 */
export function buildImageFullURI(imageModel) {
  if (!imageModel.UseRegistry) {
    return ensureTag(imageModel.Image);
  }

  const imageName = buildImageFullURIWithRegistry(imageModel);

  return ensureTag(imageName);

  function ensureTag(image, defaultTag = 'latest') {
    return image.includes(':') ? image : `${image}:${defaultTag}`;
  }
}

function buildImageFullURIWithRegistry(imageModel) {
  switch (imageModel.Registry.Type) {
    case RegistryTypes.GITLAB:
      return buildImageURIForGitLab(imageModel);
    case RegistryTypes.QUAY:
      return buildImageURIForQuay(imageModel);
    case RegistryTypes.ANONYMOUS:
      return imageModel.Image;
    default:
      return buildImageURIForOtherRegistry(imageModel);
  }

  function buildImageURIForGitLab(imageModel) {
    const slash = imageModel.Image.startsWith(':') ? '' : '/';
    return `${imageModel.Registry.URL}/${imageModel.Registry.Gitlab.ProjectPath}${slash}${imageModel.Image}`;
  }

  function buildImageURIForQuay(imageModel) {
    const name = imageModel.Registry.Quay.UseOrganisation ? imageModel.Registry.Quay.OrganisationName : imageModel.Registry.Username;
    const url = imageModel.Registry.URL ? imageModel.Registry.URL + '/' : '';
    return `${url}${name}/${imageModel.Image}`;
  }

  function buildImageURIForOtherRegistry(imageModel) {
    const url = imageModel.Registry.URL ? imageModel.Registry.URL + '/' : '';
    return url + imageModel.Image;
  }
}

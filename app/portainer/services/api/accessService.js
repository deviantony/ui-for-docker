import _ from 'lodash-es';
import { UserAccessViewModel } from '../../models/access';
import { TeamAccessViewModel } from '../../models/access';

angular.module('portainer.app').factory('AccessService', [
  '$q',
  '$async',
  'UserService',
  'TeamService',
  function AccessServiceFactory($q, $async, UserService, TeamService) {
    'use strict';
    var service = {};

    function _mapAccessData(accesses, authorizedPolicies, inheritedPolicies) {
      var availableAccesses = [];
      var authorizedAccesses = [];

      for (var i = 0; i < accesses.length; i++) {
        const access = accesses[i];

        const authorized = authorizedPolicies && authorizedPolicies[access.Id];
        const inherited = inheritedPolicies && inheritedPolicies[access.Id];

        if (authorized && inherited) {
          access.Override = true;
          authorizedAccesses.push(access);
        } else if (authorized && !inherited) {
          authorizedAccesses.push(access);
        } else if (!authorized && inherited) {
          access.Inherited = true;
          authorizedAccesses.push(access);
          availableAccesses.push(access);
        } else {
          availableAccesses.push(access);
        }
      }

      return {
        available: availableAccesses,
        authorized: authorizedAccesses,
      };
    }

    function getAccesses(authorizedUserPolicies, authorizedTeamPolicies, inheritedUserPolicies, inheritedTeamPolicies) {
      var deferred = $q.defer();

      $q.all({
        users: UserService.users(false),
        teams: TeamService.teams(),
      })
        .then(function success(data) {
          var userAccesses = data.users.map(function (user) {
            return new UserAccessViewModel(user);
          });
          var teamAccesses = data.teams.map(function (team) {
            return new TeamAccessViewModel(team);
          });

          var userAccessData = _mapAccessData(userAccesses, authorizedUserPolicies, inheritedUserPolicies);
          var teamAccessData = _mapAccessData(teamAccesses, authorizedTeamPolicies, inheritedTeamPolicies);

          var accessData = {
            availableUsersAndTeams: userAccessData.available.concat(teamAccessData.available),
            authorizedUsersAndTeams: userAccessData.authorized.concat(teamAccessData.authorized),
          };

          deferred.resolve(accessData);
        })
        .catch(function error(err) {
          deferred.reject({ msg: 'Unable to retrieve users and teams', err: err });
        });

      return deferred.promise;
    }

    async function accessesAsync(entity, parent) {
      if (!entity) {
        throw { msg: 'Unable to retrieve accesses' };
      }
      if (!entity.UserAccessPolicies) {
        entity.UserAccessPolicies = {};
      }
      if (!entity.TeamAccessPolicies) {
        entity.TeamAccessPolicies = {};
      }
      if (parent && !parent.UserAccessPolicies) {
        parent.UserAccessPolicies = {};
      }
      if (parent && !parent.TeamAccessPolicies) {
        parent.TeamAccessPolicies = {};
      }
      return await getAccesses(entity.UserAccessPolicies, entity.TeamAccessPolicies, parent ? parent.UserAccessPolicies : {}, parent ? parent.TeamAccessPolicies : {});
    }

    function accesses(entity, parent) {
      return $async(accessesAsync, entity, parent);
    }

    service.accesses = accesses;

    service.generateAccessPolicies = function (userAccessPolicies, teamAccessPolicies, selectedUserAccesses, selectedTeamAccesses, selectedRoleId) {
      const newUserPolicies = _.clone(userAccessPolicies);
      const newTeamPolicies = _.clone(teamAccessPolicies);

      _.forEach(selectedUserAccesses, (access) => (newUserPolicies[access.Id] = { RoleId: selectedRoleId ? selectedRoleId : access.Role.Id }));
      _.forEach(selectedTeamAccesses, (access) => (newTeamPolicies[access.Id] = { RoleId: selectedRoleId ? selectedRoleId : access.Role.Id }));

      const accessPolicies = {
        userAccessPolicies: newUserPolicies,
        teamAccessPolicies: newTeamPolicies,
      };

      return accessPolicies;
    };

    return service;
  },
]);

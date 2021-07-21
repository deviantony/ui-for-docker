import uuidv4 from 'uuid/v4';
class StackRedeployGitFormController {
  /* @ngInject */
  constructor($async, $state, StackService, ModalService, Notifications, WebhookHelper, FormHelper) {
    this.$async = $async;
    this.$state = $state;
    this.StackService = StackService;
    this.ModalService = ModalService;
    this.Notifications = Notifications;
    this.WebhookHelper = WebhookHelper;
    this.FormHelper = FormHelper;

    this.state = {
      inProgress: false,
      redeployInProgress: false,
      showConfig: false,
    };

    this.formValues = {
      RefName: '',
      RepositoryAuthentication: false,
      RepositoryUsername: '',
      RepositoryPassword: '',
      Env: [],
      // auto upadte
      AutoUpdate: {
        RepositoryAutomaticUpdates: false,
        RepositoryMechanism: 'Interval',
        RepositoryFetchInterval: '',
        RepositoryWebhookURL: '',
      },
    };

    this.onChange = this.onChange.bind(this);
    this.onChangeRef = this.onChangeRef.bind(this);
    this.handleEnvVarChange = this.handleEnvVarChange.bind(this);
  }

  onChangeRef(value) {
    this.onChange({ RefName: value });
  }

  onChange(values) {
    this.formValues = {
      ...this.formValues,
      ...values,
    };
  }

  async submit() {
    return this.$async(async () => {
      try {
        const confirmed = await this.ModalService.confirmAsync({
          title: 'Are you sure?',
          message: 'Any changes to this stack made locally in Portainer will be overridden by the definition in git and may cause a service interruption. Do you wish to continue',
          buttons: {
            confirm: {
              label: 'Update',
              className: 'btn-warning',
            },
          },
        });
        if (!confirmed) {
          return;
        }

        this.state.redeployInProgress = true;

        await this.StackService.updateGit(this.stack.Id, this.stack.EndpointId, this.FormHelper.removeInvalidEnvVars(this.formValues.Env), false, this.formValues);

        await this.$state.reload();
      } catch (err) {
        this.Notifications.error('Failure', err, 'Failed redeploying stack');
      } finally {
        this.state.redeployInProgress = false;
      }
    });
  }

  async saveGitSettings() {
    return this.$async(async () => {
      try {
        this.state.inProgress = true;
        await this.StackService.updateGitStackSettings(this.stack.Id, this.stack.EndpointId, this.FormHelper.removeInvalidEnvVars(this.formValues.Env), this.formValues);
        this.Notifications.success('Save stack settings successfully');
      } catch (err) {
        this.Notifications.error('Failure', err, 'Unable to save stack settings');
      } finally {
        this.state.inProgress = false;
      }
    });
  }

  isSubmitButtonDisabled() {
    return this.state.inProgress || this.state.redeployInProgress;
  }

  handleEnvVarChange(value) {
    this.formValues.Env = value;
  }

  $onInit() {
    this.formValues.RefName = this.model.ReferenceName;
    this.formValues.Env = this.stack.Env;
    // Init auto update
    if (this.stack.AutoUpdate && (this.stack.AutoUpdate.Interval || this.stack.AutoUpdate.Webhook)) {
      this.formValues.AutoUpdate.RepositoryAutomaticUpdates = true;

      if (this.stack.AutoUpdate.Interval) {
        this.formValues.AutoUpdate.RepositoryMechanism = `Interval`;
        this.formValues.AutoUpdate.RepositoryFetchInterval = this.stack.AutoUpdate.Interval;
      } else if (this.stack.AutoUpdate.Webhook) {
        this.formValues.AutoUpdate.RepositoryMechanism = `Webhook`;
        this.formValues.AutoUpdate.RepositoryWebhookURL = this.WebhookHelper.returnStackWebhookUrl(this.stack.AutoUpdate.Webhook);
      }
    }

    if (!this.formValues.AutoUpdate.RepositoryWebhookURL) {
      this.formValues.AutoUpdate.RepositoryWebhookURL = this.WebhookHelper.returnStackWebhookUrl(uuidv4());
    }

    if (this.stack.GitConfig && this.stack.GitConfig.Authentication) {
      this.formValues.RepositoryUsername = this.stack.GitConfig.Authentication.Username;
      this.formValues.RepositoryPassword = this.stack.GitConfig.Authentication.Password;
      this.formValues.RepositoryAuthentication = this.formValues.RepositoryPassword !== '';
    }
  }
}

export default StackRedeployGitFormController;

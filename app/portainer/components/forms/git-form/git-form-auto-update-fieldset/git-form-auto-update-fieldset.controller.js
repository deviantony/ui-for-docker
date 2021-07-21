class GitFormAutoUpdateFieldsetController {
  /* @ngInject */
  constructor(clipboard) {
    this.onChangeAutoUpdate = this.onChangeField('RepositoryAutomaticUpdates');
    this.onChangeMechanism = this.onChangeField('RepositoryMechanism');
    this.onChangeInterval = this.onChangeField('RepositoryFetchInterval');
    this.clipboard = clipboard;
  }

  copyWebhook() {
    this.clipboard.copyText(this.model.RepositoryWebhookURL);
    $('#copyNotification').show();
    $('#copyNotification').fadeOut(2000);
  }

  onChangeField(field) {
    return (value) => {
      this.onChange({
        ...this.model,
        [field]: value,
      });
    };
  }
}

export default GitFormAutoUpdateFieldsetController;

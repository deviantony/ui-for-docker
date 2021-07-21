export class SubscriptionViewModel {
  constructor(data) {
    this.Id = data.subscriptionId;
    this.Name = data.displayName;
  }
}

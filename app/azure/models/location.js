export class LocationViewModel {
  constructor(data) {
    this.Id = data.id;
    this.SubscriptionId = data.subscriptionId;
    this.DisplayName = data.displayName;
    this.Name = data.name;
  }
}

export class ResourceGroupViewModel {
  constructor(data, subscriptionId) {
    this.Id = data.id;
    this.SubscriptionId = subscriptionId;
    this.Name = data.name;
    this.Location = data.location;
  }
}

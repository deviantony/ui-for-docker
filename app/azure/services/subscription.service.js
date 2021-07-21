import { SubscriptionViewModel } from '../models/subscription';

/* @ngInject */
export function SubscriptionService(Subscription) {
  return {
    subscriptions,
    subscription,
  };

  async function subscriptions() {
    try {
      const results = await Subscription.query({}).$promise;
      return results.value.map((item) => new SubscriptionViewModel(item));
    } catch (err) {
      throw { msg: 'Unable to retrieve subscriptions', err };
    }
  }

  async function subscription(id) {
    const subscription = await Subscription.get({ id }).$promise;
    return new SubscriptionViewModel(subscription);
  }
}

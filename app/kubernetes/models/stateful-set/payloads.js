import { KubernetesCommonMetadataPayload } from "Kubernetes/models/common/payloads";

/**
 * KubernetesStatefulSetCreatePayload Model
 */
 const _KubernetesStatefulSetCreatePayload = Object.freeze({
    metadata: new KubernetesCommonMetadataPayload(),
    spec: {
      replicas: 0,
      serviceName: '',
      selector: {
        matchLabels: {
          app: ''
        }
      },
      volumeClaimTemplates: [],
      template: {
        metadata: {
          labels: {
            app: ''
          }
        },
        spec: {
          containers: [
            {
              name: '',
              image: '',
              env: [],
              resources: {
                limits: {},
                requests: {}
              },
              volumeMounts: []
            }
          ],
          volumes: []
        }
      }
    }
});

export class KubernetesStatefulSetCreatePayload {
  constructor() {
    Object.assign(this, JSON.parse(JSON.stringify(_KubernetesStatefulSetCreatePayload)));
  }
}

/**
 * KubernetesStatefulSetPatchPayload Model
 */
const _KubernetesStatefulSetPatchPayload = Object.freeze({
  metadata: new KubernetesCommonMetadataPayload(),
});

export class KubernetesStatefulSetPatchPayload {
  constructor() {
    Object.assign(this, JSON.parse(JSON.stringify(_KubernetesStatefulSetPatchPayload)));
  }
}

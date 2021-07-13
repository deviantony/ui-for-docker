import _ from 'lodash-es';

/**
 * NodesLimits Model
 */
const _KubernetesNodesLimits = Object.freeze({
  MaxCPU: 0,
  MaxMemory: 0,
  nodesLimits: {},
});

export class KubernetesNodesLimits {
  constructor(nodesLimits) {
    Object.assign(this, JSON.parse(JSON.stringify(_KubernetesNodesLimits)));
    this.nodesLimits = this.convertCPU(nodesLimits);

    this.calculateMaxCPUMemory();
  }

  convertCPU(nodesLimits) {
    _.forEach(nodesLimits, (value) => {
      if (value.CPU) {
        value.CPU /= 1000.0;
      }
    });
    return nodesLimits;
  }

  calculateMaxCPUMemory() {
    const nodesLimitsArray = Object.values(this.nodesLimits);
    this.MaxCPU = _.maxBy(nodesLimitsArray, 'CPU').CPU;
    this.MaxMemory = _.maxBy(nodesLimitsArray, 'Memory').Memory;
  }

  overflowForReplica(cpu, memory, instance) {
    const nodesLimits = _.cloneDeep(this.nodesLimits);

    _.forEach(nodesLimits, (value) => {
      while (instance && cpu <= value.CPU && memory <= value.Memory) {
        value.CPU -= cpu;
        value.Memory -= memory;
        instance--;
      }
    });

    return !!instance;
  }

  overflowForGlobal(cpu, memory) {
    let overflow = false;

    _.forEach(this.nodesLimits, (value) => {
      if (cpu > value.CPU || memory > value.Memory) {
        overflow = true;
      }
    });

    return overflow;
  }

  excludesPods(pods, cpuLimit, memoryLimit) {
    const nodesLimits = this.nodesLimits;

    _.forEach(pods, (value) => {
      const node = value.Node;
      if (node && nodesLimits[node]) {
        nodesLimits[node].CPU += cpuLimit;
        nodesLimits[node].Memory += memoryLimit;
      }
    });

    this.calculateMaxCPUMemory();
  }
}

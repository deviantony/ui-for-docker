package cli

import (
	portainer "github.com/portainer/portainer/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetNodesLimits gets the CPU and Memory limits(unused resources) of all nodes in the current k8s endpoint connection
func (kcl *KubeClient) GetNodesLimits() (portainer.K8sNodesLimits, error) {
	nodesLimits := make(portainer.K8sNodesLimits)

	nodes, err := kcl.cli.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	pods, err := kcl.cli.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, item := range nodes.Items {
		cpu := item.Status.Allocatable.Cpu().MilliValue()
		memory := item.Status.Allocatable.Memory().Value()

		nodesLimits[item.ObjectMeta.Name] = &portainer.K8sNodeLimits{
			CPU:    cpu,
			Memory: memory,
		}
	}

	for _, item := range pods.Items {
		if nodeLimits, ok := nodesLimits[item.Spec.NodeName]; ok {
			for _, container := range item.Spec.Containers {
				nodeLimits.CPU -= container.Resources.Requests.Cpu().MilliValue()
				nodeLimits.Memory -= container.Resources.Requests.Memory().Value()
			}
		}
	}

	return nodesLimits, nil
}

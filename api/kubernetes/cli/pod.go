package cli

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
	portainer "github.com/portainer/portainer/api"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const shellPodImage = "portainer/kubectl-shell"

// CreateUserShellPod will create a kubectl based shell for the specified user by mounting their respective service account.
// The lifecycle of the pod is managed in this function; this entails management of the following pod operations:
// - The shell pod will be scoped to specified service accounts access permissions
// - The shell pod will be automatically removed if it's not ready after specified period of time
// - The shell pod will be automatically removed after a specified max life (prevent zombie pods)
// - The shell pod will be automatically removed if request is cancelled (or client closes websocket connection)
func (kcl *KubeClient) CreateUserShellPod(ctx context.Context, serviceAccountName string) (*portainer.KubernetesShellPod, error) {
	// Schedule the pod for automatic removal
	maxPodKeepAlive := 1 * time.Hour
	maxPodKeepAliveSecondsStr := fmt.Sprintf("%d", int(maxPodKeepAlive.Seconds()))

	podPrefix := userShellPodPrefix(serviceAccountName)

	podSpec := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: podPrefix,
			Namespace:    portainerNamespace,
			Annotations: map[string]string{
				"kubernetes.io/pod.type": "kubectl-shell",
			},
		},
		Spec: v1.PodSpec{
			TerminationGracePeriodSeconds: new(int64),
			ServiceAccountName:            serviceAccountName,
			Containers: []v1.Container{
				{
					Name:    "kubectl-shell-container",
					Image:   shellPodImage,
					Command: []string{"sleep"},
					// Specify sleep time to prevent zombie pods in case portainer process is terminated
					Args:            []string{maxPodKeepAliveSecondsStr},
					ImagePullPolicy: v1.PullIfNotPresent,
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}

	shellPod, err := kcl.cli.CoreV1().Pods(portainerNamespace).Create(podSpec)
	if err != nil {
		return nil, errors.Wrap(err, "error creating shell pod")
	}

	// Wait for pod to reach ready state
	timeoutCtx, cancelFunc := context.WithTimeout(ctx, 20*time.Second)
	defer cancelFunc()
	err = kcl.waitForPodStatus(timeoutCtx, v1.PodRunning, shellPod)
	if err != nil {
		kcl.cli.CoreV1().Pods(portainerNamespace).Delete(shellPod.Name, nil)
		return nil, errors.Wrap(err, "aborting pod creation; error waiting for shell pod ready status")
	}

	if len(shellPod.Spec.Containers) != 1 {
		kcl.cli.CoreV1().Pods(portainerNamespace).Delete(shellPod.Name, nil)
		return nil, fmt.Errorf("incorrect shell pod state, expecting single container to be present")
	}

	podData := &portainer.KubernetesShellPod{
		Namespace:        shellPod.Namespace,
		PodName:          shellPod.Name,
		ContainerName:    shellPod.Spec.Containers[0].Name,
		ShellExecCommand: "env COLUMNS=200 /bin/bash", // env COLUMNS dictates minimum width of the shell
	}

	// Handle pod lifecycle/cleanup - terminate pod after maxPodKeepAlive or upon request (long-lived) cancellation
	go func() {
		select {
		case <-time.After(maxPodKeepAlive):
			log.Println("[DEBUG] [internal,kubernetes/pod] [message: pod removal schedule duration exceeded]")
			kcl.cli.CoreV1().Pods(portainerNamespace).Delete(shellPod.Name, nil)
		case <-ctx.Done():
			err := ctx.Err()
			log.Printf("[DEBUG] [internal,kubernetes/pod] [message: context error: err=%s ]\n", err)
			kcl.cli.CoreV1().Pods(portainerNamespace).Delete(shellPod.Name, nil)
		}
	}()

	return podData, nil
}

// waitForPodStatus will wait until duration d (from now) for a pod to reach defined phase/status.
// The pod status will be polled at specified delay until the pod reaches ready state.
func (kcl *KubeClient) waitForPodStatus(ctx context.Context, phase v1.PodPhase, pod *v1.Pod) error {
	log.Printf("[DEBUG] [internal,kubernetes/pod] [message: waiting for pod ready: pod=%s... ]\n", pod.Name)

	pollDelay := 500 * time.Millisecond
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			pod, err := kcl.cli.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			if pod.Status.Phase == phase {
				return nil
			}

			<-time.After(pollDelay)
		}
	}
}

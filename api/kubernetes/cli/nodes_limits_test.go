package cli

import (
	portainer "github.com/portainer/portainer/api"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	kfake "k8s.io/client-go/kubernetes/fake"
	"reflect"
	"testing"
)

func newNodes() *v1.NodeList {
	return &v1.NodeList{
		Items: []v1.Node{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node-0",
				},
				Status: v1.NodeStatus{
					Allocatable: v1.ResourceList{
						v1.ResourceName(v1.ResourceCPU):    resource.MustParse("2"),
						v1.ResourceName(v1.ResourceMemory): resource.MustParse("4M"),
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-node-1",
				},
				Status: v1.NodeStatus{
					Allocatable: v1.ResourceList{
						v1.ResourceName(v1.ResourceCPU):    resource.MustParse("3"),
						v1.ResourceName(v1.ResourceMemory): resource.MustParse("6M"),
					},
				},
			},
		},
	}
}

func newPods() *v1.PodList {
	return &v1.PodList{
		Items: []v1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-container-0",
					Namespace: "test-namespace-0",
				},
				Spec: v1.PodSpec{
					NodeName: "test-node-0",
					Containers: []v1.Container{
						{
							Name: "test-container-0",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceName(v1.ResourceCPU):    resource.MustParse("1"),
									v1.ResourceName(v1.ResourceMemory): resource.MustParse("2M"),
								},
							},
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-container-1",
					Namespace: "test-namespace-1",
				},
				Spec: v1.PodSpec{
					NodeName: "test-node-1",
					Containers: []v1.Container{
						{
							Name: "test-container-1",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceName(v1.ResourceCPU):    resource.MustParse("2"),
									v1.ResourceName(v1.ResourceMemory): resource.MustParse("3M"),
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestKubeClient_GetNodesLimits(t *testing.T) {
	type fields struct {
		cli kubernetes.Interface
	}

	fieldsInstance := fields{
		cli: kfake.NewSimpleClientset(newNodes(), newPods()),
	}

	tests := []struct {
		name    string
		fields  fields
		want    portainer.K8sNodesLimits
		wantErr bool
	}{
		{
			name:   "2 nodes 2 pods",
			fields: fieldsInstance,
			want: portainer.K8sNodesLimits{
				"test-node-0": &portainer.K8sNodeLimits{
					CPU:    1000,
					Memory: 2000000,
				},
				"test-node-1": &portainer.K8sNodeLimits{
					CPU:    1000,
					Memory: 3000000,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kcl := &KubeClient{
				cli: tt.fields.cli,
			}
			got, err := kcl.GetNodesLimits()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNodesLimits() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNodesLimits() got = %v, want %v", got, tt.want)
			}
		})
	}
}

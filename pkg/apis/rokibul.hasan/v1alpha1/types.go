package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type RokibulHasan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RokibulHasanSpec   `json:"spec"`
	Status RokibulHasanStatus `json:"status,omitempty"`
}
type ContainerSpec struct {
	Image string `json:"image,omitempty"`
	Port  int32  `json:"port,omitempty"`
}

type RokibulHasanSpec struct {
	DeploymentName string        `json:"deploymentName"`
	Replicas       *int32        `json:"replicas"`
	Container      ContainerSpec `json:"container,container"`
}

type RokibulHasanStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type RokibulHasanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []RokibulHasan `json:"items"`
}

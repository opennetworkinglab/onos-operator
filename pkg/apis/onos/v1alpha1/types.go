package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OnosClusterSpec defines the desired state of OnosCluster
type OnosClusterSpec struct {
	Size      int32                       `json:"size,omitempty"`
	Env       []corev1.EnvVar             `json:"env,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	Apps      []string                    `json:"apps,omitempty"`
	Atomix    AtomixClusterSpec           `json:"atomix,omitempty"`
}

// AtomixClusterSpec defines the desired state of the Atomix cluster
type AtomixClusterSpec struct {
	Service string `json:"service,omitempty"`
}

// OnosClusterStatus defines the observed state of OnosCluster
type OnosClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OnosCluster is the Schema for the onosclusters API
// +k8s:openapi-gen=true
type OnosCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OnosClusterSpec   `json:"spec,omitempty"`
	Status OnosClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OnosClusterList contains a list of OnosCluster
type OnosClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OnosCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OnosCluster{}, &OnosClusterList{})
}

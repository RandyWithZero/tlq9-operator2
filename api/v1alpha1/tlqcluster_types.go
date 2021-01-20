/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TLQClusterSpec defines the desired state of TLQCluster
type TLQClusterSpec struct {
	// worker size
	WorkerSize uint8 `json:"workerSize,omitempty"`
	//enable worker backup
	EnableWorkerBackup bool `json:"enableWorkerBackup,omitempty"`
}

// TLQClusterStatus defines the observed state of TLQCluster
type TLQClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TLQCluster is the Schema for the tlqclusters API
type TLQCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TLQClusterSpec   `json:"spec,omitempty"`
	Status TLQClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TLQClusterList contains a list of TLQCluster
type TLQClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TLQCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TLQCluster{}, &TLQClusterList{})
}

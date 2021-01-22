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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TLQMasterSpec defines the desired state of TLQMaster
type TLQMasterSpec struct {
	// UserName
	UserName string `json:"username,omitempty"`
	// Password
	Password string `json:"password,omitempty"`
	// AdvertiseInterval
	AdvertiseInterval uint `json:"advertiseInterval,omitempty"`
	// VRRPPasswd
	VRRPPasswd string `json:"vrrpPassword,omitempty"`
	// master image
	Image string `json:"image,omitempty"`
	// master image pull policy
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty"`
	//env
	Envs []v1.EnvVar `json:"env,omitempty"`
	//volumes
	Volumes []v1.Volume `json:"volumes,omitempty"`
	//VolumeMounts
	VolumeMounts []v1.VolumeMount `json:"volumeMounts,omitempty"`
	//VolumeClaimTemplates
	VolumeClaimTemplates []v1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`
	//port
	Port int32 `json:"port,omitempty"`
	//resource
	Resource v1.ResourceRequirements `json:"resource,omitempty"`
}
type MasterStatus string

var (
	Healthy   MasterStatus = "Healthy"
	UnHealthy MasterStatus = "UnHealthy"
	Pending   MasterStatus = "Pending"
	Fail      MasterStatus = "Fail"
)

// TLQMasterStatus defines the observed state of TLQMaster
type TLQMasterStatus struct {
	Parse MasterStatus `json:"parse,omitempty"`
	// master server address
	Server string `json:"server,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Name",type="string",JSONPath=".metadata.name"
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.parse"
//+kubebuilder:printcolumn:name="MasterAddress",type="string",JSONPath=".status.server",priority=10

// TLQMaster is the Schema for the tlqmasters API
type TLQMaster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TLQMasterSpec   `json:"spec,omitempty"`
	Status TLQMasterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TLQMasterList contains a list of TLQMaster
type TLQMasterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TLQMaster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TLQMaster{}, &TLQMasterList{})
}

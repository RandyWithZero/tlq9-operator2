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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DataPersistentSpec struct {
	//DataDir
	DataDir string `json:"dataDir,omitempty"`
	//DataDir
	DataMountType MountType `json:"dataMountType,omitempty"`
	//HostPath
	HostPath string `json:"hostPath,omitempty"`
	//AccessMode
	AccessMode v1.PersistentVolumeAccessMode `json:"accessMode,omitempty"`
	//Storage
	DataStorage resource.Quantity `json:"dataStorage,omitempty"`
}

type Resources struct {
	//LimitCpu
	LimitCpu resource.Quantity `json:"limitCpu,omitempty"`
	//LimitMemory
	LimitMemory resource.Quantity `json:"limitMemory,omitempty"`
	//RequestCpu
	RequestCpu resource.Quantity `json:"requestCpu,omitempty"`
	//RequestMemory
	RequestMemory resource.Quantity `json:"requestMemory,omitempty"`
}

type Spec struct {
	//Image: image
	Image string `json:"image,omitempty"`
	// master image pull policy
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty"`
	//port
	Port int32 `json:"port,omitempty"`
	//ZoneConfigMapName
	ZoneConfigMapName string `json:"zoneConfigMapName,omitempty"`
	//TopicConfigMapName
	TopicConfigMapName string `json:"topicConfigMapName,omitempty"`
	//DataPersistentSpec
	DataPersistentSpec *DataPersistentSpec `json:"dataPersistentSpec,omitempty"`
	//Resources
	Resources *Resources `json:"resources,omitempty"`
}

func (spec *Spec) DeepCopy() *Spec {
	persistentSpec := spec.DataPersistentSpec
	var dataPersistentSpec *DataPersistentSpec
	if persistentSpec != nil {
		dataPersistentSpec = &DataPersistentSpec{
			DataDir:       persistentSpec.DataDir,
			DataMountType: persistentSpec.DataMountType,
			HostPath:      persistentSpec.HostPath,
			AccessMode:    persistentSpec.AccessMode,
			DataStorage:   persistentSpec.DataStorage.DeepCopy(),
		}
	}
	resourcesSpec := spec.Resources
	var resources *Resources
	if persistentSpec != nil {
		resources = &Resources{
			LimitMemory:   resourcesSpec.LimitMemory.DeepCopy(),
			LimitCpu:      resourcesSpec.LimitCpu.DeepCopy(),
			RequestMemory: resourcesSpec.RequestMemory.DeepCopy(),
			RequestCpu:    resourcesSpec.RequestCpu.DeepCopy(),
		}
	}
	return &Spec{
		Image:              spec.Image,
		Port:               spec.Port,
		ImagePullPolicy:    spec.ImagePullPolicy,
		ZoneConfigMapName:  spec.ZoneConfigMapName,
		TopicConfigMapName: spec.TopicConfigMapName,
		DataPersistentSpec: dataPersistentSpec,
		Resources:          resources,
	}
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TLQMasterSpec defines the desired state of TLQMaster
type TLQMasterSpec struct {
	//UserName
	UserName string `json:"username,omitempty"`
	//Password
	Password string `json:"password,omitempty"`
	//AdvertiseInterval
	AdvertiseInterval uint `json:"advertiseInterval,omitempty"`
	//VRRPPasswd
	VRRPPasswd string `json:"vrrpPassword,omitempty"`
	//spec
	Detail *Spec `json:"detail,omitempty"`
}

// TLQMasterStatus defines the observed state of TLQMaster
type TLQMasterStatus struct {
	Parse Status `json:"parse,omitempty"`
	// master server address
	Server string `json:"server,omitempty"`
}

func (thisStatus *TLQMasterStatus) Equal(thatStatus *TLQMasterStatus) bool {
	var1 := thisStatus.Parse == thatStatus.Parse
	var2 := thisStatus.Server == thatStatus.Server
	return var1 && var2
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.parse"
//+kubebuilder:printcolumn:name="Master-Address",type="string",JSONPath=".status.server"

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

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
	"reflect"
)

type MountType string

var (
	HostPath            MountType = "hostPath"
	VolumeChaimTemplate MountType = "volumeChaimTemplate"
)

type Status string

var (
	Healthy   Status = "Healthy"
	UnHealthy Status = "UnHealthy"
	Pending   Status = "Pending"
	//Fail      Status = "Fail"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TLQClusterSpec defines the desired state of TLQCluster
type TLQClusterSpec struct {
	//worker size
	WorkerSize int `json:"workerSize,omitempty"`
	//enable worker backup
	EnableWorkerBackup bool `json:"enableWorkerBackup,omitempty"`
	//enable master backup
	EnableMasterBackup bool `json:"enableMasterBackup,omitempty"`
	//MasterTemplate
	MasterTemplate TLQMasterSpec `json:"masterTemplate,omitempty"`
	//WorkerTemplate
	WorkerTemplate TLQWorkerSpec `json:"workerTemplate,omitempty"`
}

// TLQClusterStatus defines the observed state of TLQCluster
type TLQClusterStatus struct {
	//TotalWorkerCount
	TotalWorkerCount int `json:"totalWorkerCount"`
	//ReadyWorkerCount
	ReadyWorkerCount int `json:"readyWorkerCount"`
	//ReadyWorkerServer
	ReadyWorkerServer map[string]string `json:"readyWorkerServer,omitempty"`
	//ReadyMasterServer
	ReadyMasterServer string `json:"readyMasterServer,omitempty"`
	//Parse
	Parse Status `json:"parse,omitempty"`
}

func (thisStatus *TLQClusterStatus) Equal(thatStatus *TLQClusterStatus) bool {
	var1 := thisStatus.Parse == thatStatus.Parse
	var2 := thisStatus.TotalWorkerCount == thatStatus.TotalWorkerCount
	var3 := thisStatus.ReadyWorkerCount == thatStatus.ReadyWorkerCount
	var4 := thisStatus.ReadyMasterServer == thatStatus.ReadyMasterServer
	var5 := reflect.DeepEqual(thisStatus.ReadyWorkerServer, thatStatus.ReadyWorkerServer)
	return var1 && var2 && var3 && var4 && var5
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.parse"
//+kubebuilder:printcolumn:name="Worker-Count",type="integer",JSONPath=".status.totalWorkerCount"
//+kubebuilder:printcolumn:name="Ready",type="integer",JSONPath=".status.readyWorkerCount"

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

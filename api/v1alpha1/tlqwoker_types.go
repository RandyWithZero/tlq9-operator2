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

// TLQWorkerSpec defines the desired state of TLQWorker
type TLQWorkerSpec struct {
	//WorkRootDir: worker root dir
	WorkRootDir string `json:"workRootDir,omitempty"`
	//LogLevel: log level
	LogLevel int `json:"logLevel,omitempty"`
	//IsAffinity
	IsAffinity int `json:"isAffinity,omitempty"`
	//RegisterStatus: 0 enable register ; 1 disable register
	RegisterStatus int `json:"registerStatus,omitempty"`
	//RequestServiceNum
	RequestServiceNum int `json:"requestServiceNum,omitempty"`
	//ResponseServiceNum
	ResponseServiceNum int `json:"responseServiceNum,omitempty"`
	//spec
	Detail *Spec `json:"detail,omitempty"`
}

// TLQWorkerStatus defines the observed state of TLQWorker
type TLQWorkerStatus struct {
	//parse
	Parse Status `json:"parse,omitempty"`
	// worker server address
	Server string `json:"server,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.parse"
//+kubebuilder:printcolumn:name="Worker-Address",type="string",JSONPath=".status.server"

// TLQWorker is the Schema for the tlqworkers API
type TLQWorker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TLQWorkerSpec   `json:"spec,omitempty"`
	Status TLQWorkerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TLQWorkerList contains a list of TLQWorker
type TLQWorkerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TLQWorker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TLQWorker{}, &TLQWorkerList{})
}

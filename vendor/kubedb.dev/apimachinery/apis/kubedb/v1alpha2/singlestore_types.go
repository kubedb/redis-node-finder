/*
Copyright 2023.

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

package v1alpha2

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kmapi "kmodules.xyz/client-go/api/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v2"
)

const (
	ResourceCodeSinglestore     = "sdb"
	ResourceKindSinglestore     = "Singlestore"
	ResourceSingularSinglestore = "singlestore"
	ResourcePluralSinglestore   = "singlestores"
)

// Singlestore is the Schema for the singlestores API

// +genclient
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=singlestores,singular=singlestore,shortName=sdb,categories={datastore,kubedb,appscode,all}
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".apiVersion"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Singlestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SinglestoreSpec   `json:"spec,omitempty"`
	Status            SinglestoreStatus `json:"status,omitempty"`
}

// SinglestoreSpec defines the desired state of Singlestore

type SinglestoreSpec struct {
	// AutoOps contains configuration of automatic ops-request-recommendation generation
	// +optional
	AutoOps AutoOpsSpec `json:"autoOps,omitempty"`

	// Version of Singlestore to be deployed.
	// +optional
	Version string `json:"version"`

	// Singlestore topology for node specification
	// +optional
	Topology *SinglestoreTopology `json:"topology,omitempty"`

	// StorageType can be durable (default) or ephemeral
	StorageType StorageType `json:"storageType,omitempty"`

	// Storage to specify how storage shall be used.
	Storage *core.PersistentVolumeClaimSpec `json:"storage,omitempty"`

	// ConfigSecret is an optional field to provide custom configuration file for database (i.e config.properties).
	// If specified, this file will be used as configuration file otherwise default configuration file will be used.
	// +optional
	ConfigSecret *core.LocalObjectReference `json:"configSecret,omitempty"`

	// Init is used to initialize database
	// +optional
	Init *InitSpec `json:"init,omitempty"`

	// Singlestore License secret
	LicenseSecret *SecretReference `json:"licenseSecret"`

	// Database authentication secret
	// +optional
	AuthSecret *SecretReference `json:"authSecret,omitempty"`

	// PodTemplate is an optional configuration for pods used to expose database
	// +optional
	PodTemplate *ofst.PodTemplateSpec `json:"podTemplate,omitempty"`

	// TLS contains tls configurations for client and server.
	// +optional
	TLS *kmapi.TLSConfig `json:"tls,omitempty"`

	// ServiceTemplates is an optional configuration for services used to expose database
	// +optional
	ServiceTemplates []NamedServiceTemplateSpec `json:"serviceTemplates,omitempty"`

	// Indicates that the database is halted and all offshoot Kubernetes resources except PVCs are deleted.
	// +optional
	Halted bool `json:"halted,omitempty"`

	// DeletionPolicy controls the delete operation for database
	// +optional
	DeletionPolicy DeletionPolicy `json:"deletionPolicy,omitempty"`

	// HealthChecker defines attributes of the health checker
	// +optional
	// +kubebuilder:default={periodSeconds: 10, timeoutSeconds: 10, failureThreshold: 1}
	HealthChecker kmapi.HealthCheckSpec `json:"healthChecker"`

	// Monitor is used monitor database instance
	// +optional
	Monitor *mona.AgentSpec `json:"monitor,omitempty"`
}

// SinglestoreTopology defines singlestore topology node specs for aggregators and leaves node
type SinglestoreTopology struct {
	Aggregator *SinglestoreNode `json:"aggregator,omitempty"`
	Leaf       *SinglestoreNode `json:"leaf,omitempty"`
}

type SinglestoreNode struct {
	// Replicas represents number of replica for this specific type of node
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// suffix to append with node name
	// +optional
	Suffix string `json:"suffix,omitempty"`

	// ConfigSecret is an optional field to provide custom configuration file for database (i.e config.properties).
	// If specified, this file will be used as configuration file otherwise default configuration file will be used.
	// +optional
	ConfigSecret *core.LocalObjectReference `json:"configSecret,omitempty"`

	// Storage to specify how storage shall be used.
	// +optional
	Storage *core.PersistentVolumeClaimSpec `json:"storage,omitempty"`

	// PodTemplate is an optional configuration for pods used to expose database
	// +optional
	PodTemplate *ofst.PodTemplateSpec `json:"podTemplate,omitempty"`
}

// SinglestoreStatus defines the observed state of Singlestore
type SinglestoreStatus struct {
	// Specifies the current phase of the database
	// +optional
	Phase DatabasePhase `json:"phase,omitempty"`
	// observedGeneration is the most recent generation observed for this resource. It corresponds to the
	// resource's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Conditions applied to the database, such as approval or denial.
	// +optional
	Conditions []kmapi.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:validation:Enum=server;client;metrics-exporter
type SinglestoreCertificateAlias string

const (
	SinglestoreServerCert SinglestoreCertificateAlias = "server"
	SinglestoreClientCert SinglestoreCertificateAlias = "client"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SinglestoreList contains a list of Singlestore
type SinglestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Singlestore `json:"items"`
}

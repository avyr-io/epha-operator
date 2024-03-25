/*
Copyright 2024.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// TargetResource defines the Kubernetes resource to be annotated.
type TargetResource struct {
	// Kind is the type of Kubernetes resource (e.g., Deployment, Service, etc.)
	Kind string `json:"kind"`

	// Name is the name of the resource.
	Name string `json:"name"`

	// Namespace is the namespace of the resource. Optional if the resource is not namespaced.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// AnnotatedObjectSpec defines the desired state of AnnotatedObject
type AnnotatedObjectSpec struct {
	// Specifies the target resources and the metadata to apply to each.
	Targets []TargetResourceWithMetadata `json:"targets"`
}

type TargetResourceWithMetadata struct {
	// Inherits TargetResource fields (Kind, Name, Namespace)
	TargetResource `json:",inline"`

	// Metadata contains the annotations and potentially other metadata to merge with the target resource.
	Metadata ResourceMetadata `json:"metadata"`
}

type ResourceMetadata struct {
	// Annotations to apply or merge with the target resource.
	Annotations map[string]string `json:"annotations"`
}

// AnnotatedObjectStatus defines the observed state of AnnotatedObject
type AnnotatedObjectStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=ao
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// AnnotatedObject is the Schema for the epha.avyr.io API
type AnnotatedObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AnnotatedObjectSpec   `json:"spec,omitempty"`
	Status AnnotatedObjectStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AnnotatedObjectList contains a list of AnnotatedObject
type AnnotatedObjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AnnotatedObject `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AnnotatedObject{}, &AnnotatedObjectList{})
}

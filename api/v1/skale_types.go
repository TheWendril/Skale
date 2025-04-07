/*
Copyright 2025.

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
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SkaleSpec defines the desired state of Skale.

type MetricSpec struct {
	Type     string               `json:"type"` // Ex: "Resource"
	Resource ResourceMetricSource `json:"resource,omitempty"`
}

type ResourceMetricSource struct {
	Name                     string `json:"name"`                     // Ex: "cpu" ou "memory"
	TargetAverageUtilization *int32 `json:"targetAverageUtilization"` // Ex: 80 (%)
}

type SkaleSpec struct {
	ScaleTargetRef autoscalingv1.CrossVersionObjectReference `json:"scaleTargetRef"`
	MinReplicas    int32                                     `json:"minReplicas"`
	MaxReplicas    int32                                     `json:"maxReplicas"`
	Metrics        []MetricSpec                              `json:"metrics,omitempty"`
}

// SkaleStatus defines the observed state of Skale.
type SkaleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Skale is the Schema for the skales API.
type Skale struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SkaleSpec   `json:"spec,omitempty"`
	Status SkaleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SkaleList contains a list of Skale.
type SkaleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Skale `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Skale{}, &SkaleList{})
}

/*


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

// SampleOperatorSpec defines the desired state of SampleOperator
type SampleOperatorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of SampleOperator. Edit SampleOperator_types.go to remove/update
	Size                int    `json:"size"`
	ServiceInstanceName string `json:"serviceInstanceName"`
}

// SampleOperatorStatus defines the observed state of SampleOperator
type SampleOperatorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	SampleConfigMap string `json:"sampleConfigMap"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// SampleOperator is the Schema for the sampleoperators API
type SampleOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SampleOperatorSpec   `json:"spec,omitempty"`
	Status SampleOperatorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SampleOperatorList contains a list of SampleOperator
type SampleOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SampleOperator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SampleOperator{}, &SampleOperatorList{})
}

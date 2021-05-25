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

// DeveloperSpec defines the desired state of Developer
type DeveloperSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Developer. Edit Developer_types.go to remove/update

	Email      string           `json:"email,omitempty"`
	FirstName  string           `json:"firstName,omitempty"`
	LastName   string           `json:"lastName,omitempty"`
	UserName   string           `json:"userName,omitempty"`
	Attributes []AttributesSpec `json:"attributes,omitempty"`

	/*{
		"email" : "developer_email",
		"firstName" : "first_name",
		"lastName" : "last_name",
		"userName" : "user_name",
		"attributes" : [{
		   "name": "MINT_BILLING_TYPE",
		   "value": "one of PREPAID | POSTPAID"
		}]
	  }*/

}

// DeveloperStatus defines the observed state of Developer
type DeveloperStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	DeveloperId string `json:"developerId,omitempty"`
}

// +kubebuilder:object:root=true

// Developer is the Schema for the developers API
// +kubebuilder:printcolumn:name="DeveloperEmail",type="string",JSONPath=".spec.email",description="Developer Email"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Developer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeveloperSpec   `json:"spec,omitempty"`
	Status DeveloperStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DeveloperList contains a list of Developer
type DeveloperList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Developer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Developer{}, &DeveloperList{})
}

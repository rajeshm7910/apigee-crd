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

// DeveloperAppSpec defines the desired state of DeveloperApp
type DeveloperAppSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of DeveloperApp. Edit DeveloperApp_types.go to remove/update
	Name        string            `json:"name,omitempty"`
	Attributes  []AttributesSpec  `json:"attributes,omitempty"`
	Status      string            `json:"status,omitempty"`
	Scopes      []string          `json:"scopes,omitempty" protobuf:"bytes,3,rep,name=scopes"`
	ApiProducts []string          `json:"apiProducts,omitempty" protobuf:"bytes,3,rep,name=apiProducts"`
	Credentials []CredentialsSpec `json:"credentials,omitempty"`
	CallbackUrl string            `json:"callbackUrl,omitempty"`
	/*
		  ---
			attributes:
			- name: ADMIN_EMAIL
			  value: admin@example.com
			- name: DisplayName
			  value: My App
			- name: Notes
			  value: Notes for developer app
			- name: MINT_BILLING_TYPE
			  value: POSTPAID
			credentials:
			- apiProducts: []
			  attributes: []
			  consumerKey: F91jQrfX6CKhyEheXFBL3gxxxxx
			  consumerSecret: TLbUJFyzOlLxxxx
			  expiresAt: -1
			  scopes: []
			  status: approved
			name: myapp
			scopes: []
			status: approved
	*/

}

type CredentialsSpec struct {
	Attributes     []AttributesSpec `json:"attributes"`
	Status         string           `json:"status,omitempty"`
	ConsumerKey    string           `json:"consumerKey,omitempty"`
	ConsumerSecret string           `json:"consumerSecret,omitempty"`
	ApiProducts    []AppProductSpec `json:"apiProducts,omitempty" protobuf:"bytes,3,rep,name=apiProducts"`
	Scopes         []string         `json:"scopes,omitempty" protobuf:"bytes,3,rep,name=scopes"`
	ExpiresAt      int              `json:"expiresAt,omitempty"`
}

type AppProductSpec struct {
	AppProduct string `json:"apiproduct,omitempty"`
	Status     string `json:"status,omitempty"`
}

// DeveloperAppStatus defines the observed state of DeveloperApp
type DeveloperAppStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ConsumerKey    string `json:"consumerKey,omitempty"`
	ConsumerSecret string `json:"consumerSecret,omitempty"`
}

// +kubebuilder:object:root=true

// DeveloperApp is the Schema for the developerapps API
// +kubebuilder:printcolumn:name="AppName",type="string",JSONPath=".spec.name",description="App Name"
// +kubebuilder:printcolumn:name="ConsumerKey",type="string",JSONPath=".status.consumerKey",description="Consumer Key"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type DeveloperApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeveloperAppSpec   `json:"spec,omitempty"`
	Status DeveloperAppStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DeveloperAppList contains a list of DeveloperApp
type DeveloperAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DeveloperApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DeveloperApp{}, &DeveloperAppList{})
}

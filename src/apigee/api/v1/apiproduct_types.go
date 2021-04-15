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

// ApiProductSpec defines the desired state of ApiProduct
type ApiProductSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ApiProduct. Edit ApiProduct_types.go to remove/update
	//Foo string `json:"foo,omitempty"`

	Name          string           `json:"name,omitempty"`
	ApprovalType  string           `json:"approvalType,omitempty"`
	Description   string           `json:"description,omitempty"`
	DisplayName   string           `json:"displayName,omitempty"`
	ApiResources  []string         `json:"apiResources,omitempty" protobuf:"bytes,3,rep,name=apiResources"`
	Attributes    []AttributesSpec `json:"attributes"`
	Environments  []string         `json:"environments,omitempty" protobuf:"bytes,3,rep,name=environments"`
	Proxies       []string         `json:"proxies,omitempty" protobuf:"bytes,3,rep,name=proxies"`
	Scopes        []string         `json:"scopes,omitempty" protobuf:"bytes,3,rep,name=scopes"`
	Quota         string           `json:"quota,omitempty"`
	QuotaInterval string           `json:"quotaInterval,omitempty"`
	QuotaTimeUnit string           `json:"quotaTimeUnit,omitempty"`
}

type AttributesSpec struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

/*
type ApiResourcesSpec struct {
	Path string `json:"path,omitempty"`
}

type EnvironmentsSpec struct {
	Name string `json:"name,omitempty"`
}

type ProxiesSpec struct {
	Proxy string `json:"proxy,omitempty"`
}

type ScopesSpec struct {
	Proxy string `json:"scope,omitempty"`
}
*/

// ApiProductStatus defines the observed state of ApiProduct
type ApiProductStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// ApiProduct is the Schema for the apiproducts API
type ApiProduct struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApiProductSpec   `json:"spec,omitempty"`
	Status ApiProductStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApiProductList contains a list of ApiProduct
type ApiProductList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApiProduct `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ApiProduct{}, &ApiProductList{})
}

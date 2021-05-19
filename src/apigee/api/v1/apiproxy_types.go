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

type ApiProxySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ApiProxy. Edit ApiProxy_types.go to remove/update
	Name        string `json:"name,omitempty"`
	ZipUrl      string `json:"zipurl,omitempty"`
	OpenApiSpec string `json:"openapispec,omitempty"`
}

// ApiProxyStatus defines the observed state of ApiProxy
type ApiProxyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	DeploymentState string `json:"deploymentState,omitempty"`
	Revision        int    `json:"revision,omitempty"`
}

// +kubebuilder:object:root=true

// ApiProxy is the Schema for the apiproxies API
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="DEPLOYMENT",type="string",JSONPath=".status.deploymentState",description="Deployment Status"
// +kubebuilder:printcolumn:name="REVISION",type="integer",JSONPath=".status.revision",description="Revision"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

type ApiProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApiProxySpec   `json:"spec,omitempty"`
	Status ApiProxyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApiProxyList contains a list of ApiProxy
type ApiProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApiProxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ApiProxy{}, &ApiProxyList{})
}

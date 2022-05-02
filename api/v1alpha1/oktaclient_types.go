/*
Copyright 2022.

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

// OktaClientSpec defines the desired state of OktaClient
type OktaClientSpec struct {

	// +kubebuilder:validation:MinLength=1
	Name string `json:"name,omitempty"`

	// +kubebuilder:validation:MinLength=1
	ClientUri string `json:"clientUri,omitempty"`

	// +kubebuilder:validation:MinItems=1
	RedirectUris []string `json:"redirectUris,omitempty"`

	// +kubebuilder:validation:MinItems=1
	PostLogoutRedirectUris []string `json:"postLogoutRedirectUris,omitempty"`

	// +kubebuilder:validation:MinItems=1
	TrustedOrigins []string `json:"trustedOrigins,omitempty"`

	// +kubebuilder:validation:MaxLength=30
	// +kubebuilder:validation:MinLength=1
	GroupId string `json:"groupId,omitempty"`
}

// OktaClientStatus defines the observed state of OktaClient
type OktaClientStatus struct {

	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OktaClient is the Schema for the oktaclients API
type OktaClient struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OktaClientSpec   `json:"spec,omitempty"`
	Status OktaClientStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OktaClientList contains a list of OktaClient
type OktaClientList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OktaClient `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OktaClient{}, &OktaClientList{})
}

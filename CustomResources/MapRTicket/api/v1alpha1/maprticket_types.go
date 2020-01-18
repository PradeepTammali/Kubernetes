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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MapRTicketSpec defines the desired state of MapRTicket
type MapRTicketSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of MapRTicket. Edit MapRTicket_types.go to remove/update
	// Foo string `json:"foo,omitempty"`

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Required

	// UserID of the user who is trying to create ticket.
	// UserID is GLobalUnixID we get when we register in IDM.
	UserID int64 `json:"userID,omitempty"`

	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required

	// UserName of the user who is trying to create ticket
	// UserName will be the signum id of the user
	UserName string `json:"userName,omitempty"`

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Required

	// GroupID of the user who is trying to create ticket
	// GroupID will be the IDM group id of the user which he belongs to in NSC.
	GroupID int64 `json:"groupID,omitempty"`

	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required

	// GroupName of the user who is trying to create ticket
	// GroupName will be the IDM group name of the user which he belongs to in NSC.
	GroupName string `json:"groupName,omitempty"`

	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required

	// Password of the user who is trying to create ticket
	Password string `json:"password,omitempty"`

	// +kubebuilder:validation:Required

	// This flag tells the controller whether to create secret or not with mapr ticket encoded in base64.
	// MaprTicket Resource does not delete the secret upon deletion of MapRTicket resource.
	// When set to true, creates secret with same name as Resource name.
	CreateSecret bool `json:"createSecret,omitempty"`
}

/*
We define a custom type to hold our  MapRTicket status.  It's actually
just a string under the hood, but the type gives extra documentation,
and allows us to attach validation on the type instead of the field,
making the validation more easily reusable.
*/

// TicketPhase describes MapRTicket status.
// Only one of the following TicketPhase may be specified.
// If none of the following TicketPhase is specified, the default one
// is Creating.
// +kubebuilder:validation:Enum=Creating;Completed;Failed
type TicketPhase string

const (
	// Creating is when the process of ticket creation is started.
	Creating TicketPhase = "Creating"

	Completed TicketPhase = "Completed"

	Failed TicketPhase = "Failed"
)

// MapRTicketStatus defines the observed state of MapRTicket
type MapRTicketStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Ticket validity which is been generated. Ticket will be expired on the date which is been shown here.
	// +optional
	TicketExpiryDate *metav1.Time `json:"ticketExpiryDate,omitempty"`

	// The state of the ticket generated. Can be Available, Unavailable
	// +optional
	Phase TicketPhase `json:"phase,omitempty"`

	// The MapR Ticket information which contains the details of the ticket.
	// +optional
	MaprTicketInfo string `json:"maprTicketInfo,omitempty"`

	// The MapR Ticket of the User
	// +optional
	MaprTicket string `json:"maprTicket,omitempty"`

	// The name of the ticket secret generated.
	// +optional
	TicketSecretName string `json:"ticketSecretName,omitempty"`

	// The namespace of the ticket secret generated.
	// +optional
	TicketSecretNamespace string `json:"ticketSecretNamespace,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// MapRTicket is the Schema for the maprtickets API
type MapRTicket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MapRTicketSpec   `json:"spec,omitempty"`
	Status MapRTicketStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MapRTicketList contains a list of MapRTicket
type MapRTicketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MapRTicket `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MapRTicket{}, &MapRTicketList{})
}

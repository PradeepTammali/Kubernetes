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
// +kubebuilder:docs-gen:collapse=Apache License

package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	validationutils "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"os"
	"regexp"
	"strconv"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:docs-gen:collapse=Go imports

// log is for logging in this package.
var maprticketlog = logf.Log.WithName("maprticket-resource")

func (r *MapRTicket) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-nsc-k8s-io-v1alpha1-maprticket,mutating=true,failurePolicy=fail,groups=nsc.k8s.io,resources=maprtickets,verbs=create;update,versions=v1alpha1,name=mmaprticket.kb.io

var _ webhook.Defaulter = &MapRTicket{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *MapRTicket) Default() {
	maprticketlog.Info("default", "name", r.Name)
	// TODO(user): fill in your defaulting logic.
	userID, err := strconv.ParseInt(os.Getenv("DEFAULT_USERID"), 10, 64)
	if err != nil {
		maprticketlog.Info("default", "UserID", "userID is not an integer.")
	}
	if r.Spec.UserID != userID {
		r.Spec.UserID = userID
	}
	var userName = os.Getenv("DEFAULT_USERNAME")
	if r.Spec.UserName != userName {
		r.Spec.UserName = userName
	}
	groupID, err := strconv.ParseInt(os.Getenv("DEFAULT_GROUPID"), 10, 64)
	if err != nil {
		maprticketlog.Info("default", "GroupID", "groupID is not an integer.")
	}
	if r.Spec.GroupID != groupID {
		r.Spec.GroupID = groupID
	}
	var groupName = os.Getenv("DEFAULT_GROUPNAME")
	if r.Spec.GroupName != groupName {
		r.Spec.GroupName = groupName
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-nsc-k8s-io-v1alpha1-maprticket,mutating=false,failurePolicy=fail,groups=nsc.k8s.io,resources=maprtickets,versions=v1alpha1,name=vmaprticket.kb.io

var _ webhook.Validator = &MapRTicket{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *MapRTicket) ValidateCreate() error {
	maprticketlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return r.validateMapRTicket()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *MapRTicket) ValidateUpdate(old runtime.Object) error {
	maprticketlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *MapRTicket) ValidateDelete() error {
	maprticketlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

// We validate the name and the spec of the MapRTicket.
func (r *MapRTicket) validateMapRTicket() error {
	maprticketlog.Info("validate MapRTicket", "name", r.Name)
	var allErrs field.ErrorList
	maprticketlog.Info("validate MapRTicket", "validation", "Validating MapRTicket spec items.")
	if err := r.validateMapRTicketSpec(); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := r.validateMapRTicketName(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "nsc.k8s.io", Kind: "MapRTicket"},
		r.Name, allErrs)
}

func (r *MapRTicket) validateMapRTicketName() *field.Error {
	if len(r.ObjectMeta.Name) > validationutils.DNS1123SubdomainMaxLength {
		// Kubernetes resources can have names up to 253 characters long.
		// The characters allowed in names are: digits (0-9), lower case letters (a-z), -, and .
		return field.Invalid(field.NewPath("metadata").Child("name"), r.Name, "must be no more than 253 characters")
	}
	return nil
}

// Validating MapRTicket Spec
func (r *MapRTicket) validateMapRTicketSpec() *field.Error {
	maprticketlog.Info("validate MapRTicketSpec", "name", r.Name)
	// The field helpers from the kubernetes API machinery help us return nicely
	// structured validation errors.
	var userIDRes = validateUserID(
		r.Spec.UserID,
		field.NewPath("spec").Child("userID"))
	if userIDRes != nil {
		return userIDRes
	}
	var userNameRes = validateUserName(
		r.Spec.UserName,
		field.NewPath("spec").Child("userName"))
	if userNameRes != nil {
		return userNameRes
	}
	var groupIDRes = validateGroupID(
		r.Spec.GroupID,
		field.NewPath("spec").Child("groupID"))
	if groupIDRes != nil {
		return groupIDRes
	}
	var groupNameRes = validateGroupName(
		r.Spec.GroupName,
		field.NewPath("spec").Child("groupName"))
	if groupNameRes != nil {
		return groupNameRes
	}
	return nil
}

// Validating UserID
func validateUserID(userID int64, fldPath *field.Path) *field.Error {
	// Validation of UserID
	if err := validationutils.IsValidUserID(userID); err != nil {
		// Checking if UserID is valid unix id
		return field.Invalid(fldPath, userID, "userID not a valid unix ID.")
	}
	return nil
}

// Validating UserName
func validateUserName(userName string, fldPath *field.Path) *field.Error {
	// Validation of UserName
	if _, err := regexp.MatchString("^[a-z][-a-z0-9_]*$", userName); err != nil {
		// Checking if UserID is valid unix id
		return field.Invalid(fldPath, userName, "userName is not a valid. Must match regex ^[a-z][-a-z0-9_]*$")
	}
	return nil
}

// Validating GroupID
func validateGroupID(groupID int64, fldPath *field.Path) *field.Error {
	// Validation of UserID
	if err := validationutils.IsValidGroupID(groupID); err != nil {
		// Checking if UserID is valid unix id
		return field.Invalid(fldPath, groupID, "groupID not a valid unix ID.")
	}
	return nil
}

// Validating GroupName
func validateGroupName(groupName string, fldPath *field.Path) *field.Error {
	// Validation of UserName
	if _, err := regexp.MatchString("^[a-z][-a-z0-9_]*$", groupName); err != nil {
		// Checking if UserID is valid unix id
		return field.Invalid(fldPath, groupName, "groupName is not a valid. Must match regex ^[a-z][-a-z0-9_]*$")
	}
	return nil
}

// +kubebuilder:docs-gen:collapse=Validate object name

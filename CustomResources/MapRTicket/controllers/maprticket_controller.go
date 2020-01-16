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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nscv1alpha1 "nsc/k8s/io/api/v1alpha1"

	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// MapRTicketReconciler reconciles a MapRTicket object
type MapRTicketReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *MapRTicketReconciler) requestLogger(req ctrl.Request) logr.Logger {
	return r.Log.WithValues("maprticket", req.NamespacedName)
}

func (r *MapRTicketReconciler) updateMapRTicketStatus(req ctrl.Request, maprticket *nscv1alpha1.MapRTicket) error {
	log := r.requestLogger(req)
	if status_error := r.Status().Update(context.Background(), maprticket); status_error != nil {
		log.Error(status_error, "Error while updating the status of the maprticket.")
		return status_error
	}
	return nil
}

// +kubebuilder:rbac:groups=nsc.k8s.io,resources=maprtickets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nsc.k8s.io,resources=maprtickets/status,verbs=get;update;patch
// As we create secret for mapr ticket we need RBACs to perform operations on secrets
// +kubebuilder:rbac:groups=batch,resources=secrets,verbs=get;list;watch;create;update;patch;delete
func (r *MapRTicketReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.requestLogger(req)

	// MapRTicket instance
	var maprTicket = &nscv1alpha1.MapRTicket{}
	log.Info("Fetching MapRticket Resource.")
	if err := r.Get(ctx, req.NamespacedName, maprTicket); err != nil {
		// log.Error(err, " error while fetching maprticket")
		if apierrors.IsNotFound(err) {
			log.Info("MapRTicket " + req.Name + " is not found.")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("Fetching MapRticket is done.")
	if maprTicket.Status.Phase == "" {
		maprTicket.Status.Phase = nscv1alpha1.Creating
		// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
		// maprTicket.Status.TicketExpiryDate = &metav1.Time{Time: time.Now()}
		// maprTicket.Status.TicketSecretName = maprTicket.Name
		// maprTicket.Status.TicketSecretNamespace = maprTicket.Namespace
		log.Info("Updating the status to creating.")
		if status_err := r.updateMapRTicketStatus(req, maprTicket); status_err != nil {
			return ctrl.Result{}, status_err
		}
		log.Info("Status updated.")
		// Secret Name validation.
		secret := &apiv1.Secret{}
		secretKey := client.ObjectKey{Namespace: maprTicket.Namespace, Name: maprTicket.Name}
		if secretErr := r.Get(ctx, secretKey, secret); secretErr != nil {
			if !apierrors.IsNotFound(secretErr) {
				// log.Error(secretErr, "Error occured while reading secret.")
				maprTicket.Status.Phase = nscv1alpha1.Failed
				r.updateMapRTicketStatus(req, maprTicket)
				return ctrl.Result{}, secretErr
			}
			log.Info("Ignore if it is not found error.")
		}
		if secret.Name == maprTicket.Name {
			// log.Error(apierrors.NewAlreadyExists(schema.GroupResource{}, "Secret "+secret.Name), "Already exist.")
			maprTicket.Status.Phase = nscv1alpha1.Failed
			r.updateMapRTicketStatus(req, maprTicket)
			return ctrl.Result{}, apierrors.NewAlreadyExists(schema.GroupResource{}, "Secret "+secret.Name)
		}
		// Validation success. No secret available witha same name.
		log.Info("Secret name validation done. No secret available with same name.")
		// Connect to MapR
		log.Info("Connecting to MapR to generate ticket.")
		if createErr := r.createMapRTicket(req, maprTicket, secret); createErr != nil {
			log.Error(createErr, "Error while creating MapRticket for " + secret.Name)
			return ctrl.Result{}, apierrors.NewInternalError(createErr)
		}
		// Create secret here and update the status
	}
	log.Info(maprTicket.Name + " " + maprTicket.Namespace + " " + maprTicket.Spec.UserName + " " + maprTicket.Spec.GroupName + " " + maprTicket.Status.TicketSecretName + " " + maprTicket.Status.TicketSecretNamespace + " " + string(maprTicket.Status.Phase))
	return ctrl.Result{}, nil
}

func (r *MapRTicketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nscv1alpha1.MapRTicket{}).
		Owns(&apiv1.Secret{}).
		Complete(r)
}

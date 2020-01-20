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
	"encoding/base64"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nscv1alpha1 "nsc/k8s/io/api/v1alpha1"

	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
)

// MapRTicketReconciler reconciles a MapRTicket object
type MapRTicketReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	Recorder record.EventRecorder
}

func (r *MapRTicketReconciler) requestLogger(req ctrl.Request) logr.Logger {
	return r.Log.WithValues("maprticket", req.NamespacedName)
}

func (r *MapRTicketReconciler) updateMapRTicketStatus(req ctrl.Request, maprticket *nscv1alpha1.MapRTicket) error {
	log := r.requestLogger(req)
	if statusError := r.Status().Update(context.Background(), maprticket); statusError != nil {
		log.Error(statusError, "Error while updating the status of the maprticket.")
		return statusError
	}
	return nil
}

// +kubebuilder:rbac:groups=nsc.k8s.io,resources=maprtickets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nsc.k8s.io,resources=maprtickets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=batch,resources=secrets,verbs=get;list;watch;create;update;patch;delete
func (r *MapRTicketReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.requestLogger(req)

	// MapRTicket instance
	var maprTicket = &nscv1alpha1.MapRTicket{}

	// Fetching MapRTicket Resource
	log.Info("Fetching MapRticket Resource.")
	if err := r.Get(ctx, req.NamespacedName, maprTicket); err != nil {
		// log.Error(err, " error while fetching maprticket")
		if apierrors.IsNotFound(err) {
			log.Info("MapRTicket " + req.Name + " is not found.")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("Fetching MapRticket is done.")

	// Creating MapRTicket Resource
	if maprTicket.Status.Phase == "" {
		log.Info("Phase is nil, so Creating MapR Ticket for the user " + maprTicket.Spec.UserName)
		maprTicket.Status.Phase = nscv1alpha1.Creating
		log.Info("Updating the status to creating.")
		if status_err := r.updateMapRTicketStatus(req, maprTicket); status_err != nil {
			log.Error(status_err, "ERROR - Error while updating the status of the Resource of type MapRTicket.")
			r.Recorder.Eventf(maprTicket, apiv1.EventTypeWarning, "FailedUpdate", "Error updating: MapRTicket \""+maprTicket.Name+"\" status update failed.")
			return ctrl.Result{}, status_err
		}
		log.Info("Status updated.")

		// Connecting to MapR to create MapR Ticket
		// r.Recorder.Eventf(maprTicket, apiv1.EventTypeNormal, "Connecting", "Establishing connecting to MapR Storage Cluster.")
		log.Info("Connecting to MapR to generate ticket.")
		if createErr := r.createMapRTicket(req, maprTicket); createErr != nil {
			log.Error(createErr, "Error while creating MapRticket for user "+maprTicket.Spec.UserName)
			maprTicket.Status.Phase = nscv1alpha1.Failed
			r.updateMapRTicketStatus(req, maprTicket)
			r.Recorder.Eventf(maprTicket, apiv1.EventTypeWarning, "FailedCreate", "Error creating: MapRTicket \""+maprTicket.Name+"\" create failed.")
			return ctrl.Result{}, apierrors.NewInternalError(createErr)
		}
		log.Info("MapRTicket is generated succesfully. Updating the status of the resource to completed.")
		maprTicket.Status.Phase = nscv1alpha1.Completed
		r.updateMapRTicketStatus(req, maprTicket)

		if maprTicket.Spec.CreateSecret {
			// r.Recorder.Eventf(maprTicket, apiv1.EventTypeNormal, "Creating", "Creating Secret \""+maprTicket.Name+"\"")
			log.Info("CreateSecret is true, so creating MapR Ticket in the current namespace with same name as MapRTicket Resource.")
			// Secret Name validation.
			secret := &apiv1.Secret{}
			secretKey := client.ObjectKey{Namespace: maprTicket.Namespace, Name: maprTicket.Name}
			if secretErr := r.Get(ctx, secretKey, secret); secretErr != nil {
				if !apierrors.IsNotFound(secretErr) {
					log.Error(secretErr, "ERROR- Error occured while reading secret.")
					maprTicket.Status.Phase = nscv1alpha1.Failed
					r.updateMapRTicketStatus(req, maprTicket)
					r.Recorder.Eventf(maprTicket, apiv1.EventTypeWarning, "FailedListing", "Error fetching: Can not list resources Secrets.")
					return ctrl.Result{}, secretErr
				}
				log.Info("Ingnoring if it is Secret not found error.")
			}
			if secret.Name == maprTicket.Name {
				log.Error(apierrors.NewAlreadyExists(schema.GroupResource{}, "Secret "+secret.Name), " Secret with the mentioned is already exist. Throwing already exist error. Updating status to Failed ...")
				maprTicket.Status.Phase = nscv1alpha1.Failed
				r.updateMapRTicketStatus(req, maprTicket)
				r.Recorder.Eventf(maprTicket, apiv1.EventTypeWarning, "FailedCreate", "Error creating: Secret \""+secret.Name+"\" already exist.")
				return ctrl.Result{}, apierrors.NewAlreadyExists(schema.GroupResource{}, "Secret "+secret.Name)
			}
			// Validation success. No secret available with same name.
			log.Info("Secret name validation done. No secret available with same name.")
			log.Info("Creating secret with encoded MapR Ticket.")
			log.Info("Preparing Secret Object.")
			secret = &apiv1.Secret{
				Type: apiv1.SecretTypeOpaque,
				ObjectMeta: metav1.ObjectMeta{
					Name:      maprTicket.Name,
					Namespace: maprTicket.Namespace,
					Labels:    maprTicket.GetLabels(),
				},
			}
			if maprTicket.Status.MaprTicket != "" || len(maprTicket.Status.MaprTicket) != 0 {
				secret.Data = map[string][]byte{
					"CONTAINER_TICKET": []byte(base64.StdEncoding.EncodeToString(([]byte(maprTicket.Status.MaprTicket)))),
				}
			}
			log.Info("Creating resource Secret in current namespace with MapR Ticket.")
			if secretCreateErr := r.Create(ctx, secret); secretCreateErr != nil {
				log.Error(secretCreateErr, "ERROR - unable to create MapRTicket Secret")
				maprTicket.Status.Phase = nscv1alpha1.Failed
				r.updateMapRTicketStatus(req, maprTicket)
				r.Recorder.Eventf(maprTicket, apiv1.EventTypeWarning, "FailedCreate", "Error creating: Secret \""+secret.Name+"\" is can not be created.")
				return ctrl.Result{}, secretCreateErr
			}
			log.Info("Created MapR Ticket Secret successfully.")

			// Update the status of resource maprticket.
			log.Info("Updating secretname and namespace in the status of the resource.")
			maprTicket.Status.TicketSecretName = secret.ObjectMeta.Name
			maprTicket.Status.TicketSecretNamespace = secret.ObjectMeta.Namespace
			r.updateMapRTicketStatus(req, maprTicket)
			// TODO: Update the events here
		} else {
			log.Info("CreateSecret is set to false.")
		}
	}
	return ctrl.Result{}, nil
}

var (
	secretOwnerKey = ".metadata.controller"
)

func (r *MapRTicketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(&apiv1.Secret{}, secretOwnerKey, func(rawObj runtime.Object) []string {
		// grab the job object, extract the owner...
		secret := rawObj.(*apiv1.Secret)
		owner := metav1.GetControllerOf(secret)
		if owner == nil {
			return nil
		}
		// ...make sure it's a MapRTicket...
		if owner.APIVersion != nscv1alpha1.GroupVersion.String() || owner.Kind != "MapRTicket" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&nscv1alpha1.MapRTicket{}).
		Owns(&apiv1.Secret{}).
		Complete(r)
}

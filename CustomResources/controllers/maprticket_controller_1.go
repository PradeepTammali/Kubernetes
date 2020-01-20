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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	nscv1alpha1 "nsc/k8s/io/api/v1alpha1"

	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

// MapRTicketReconciler reconciles a MapRTicket object
type MapRTicketReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=nsc.k8s.io,resources=maprtickets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nsc.k8s.io,resources=maprtickets/status,verbs=get;update;patch

// As we create secret for mapr ticket we need RBACs to perform operations on secrets
// +kubebuilder:rbac:groups=batch,resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *MapRTicketReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("maprticket", req.NamespacedName)

	// your logic here
	// Loading kubeconfig file
	/*log.Info("kubeconfig - loading kube config file from .kube folder.")
	var  kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "/home/pradeep/.kube/config", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	log.Info("kubeconfig - loading kube config file to get the kube context is successfull.")*/

	// or we can use the below method to get the config
	/*
		import "sigs.k8s.io/controller-runtime/pkg/client/config"
		cfg, err := config.GetConfig()
	*/

	config, err := config.GetConfig()
	// config, err := clientcmd.BuildConfigFromFlags("", "/home/pradeep/.kube/config")
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	log.Info("clientset - Building the config to create kubernetes resources.")

	// Method for MapRTicket create which creates maprticketsecret in the current namespace
	var maprTicket nscv1alpha1.MapRTicket
	// Fetching the MapRTicket object

	if err := r.Get(ctx, req.NamespacedName, &maprTicket); err != nil {
		log.Error(err, "unable to fetch MapRTicket")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Updating the status of MapRTicket with nil values
	// TODO: need to update with original values
	maprTicket.Status.TicketExpiryDate = nil
	maprTicket.Status.Phase = nscv1alpha1.Available
	maprTicket.Status.TicketSecretName = req.Name
	maprTicket.Status.TicketSecretName = req.Namespace
	if err := r.Status().Update(ctx, &maprTicket); err != nil {
		log.Error(err, "unable to update CronJob status")
		return ctrl.Result{}, err
	}

	// Creating secret with mapr ticket
	createMapRTicketSecret := func(ticketstring string, maprticket *nscv1alpha1.MapRTicket) (*apiv1.Secret, error) {
		log.Info("Secret - Inside the method of create secret createMapRTicketSecret.")
		secret := &apiv1.Secret{}
		log.Info("Secret -  checking if secret with same name already exist.")

		// we can create the secret. and got an error then we can if it is already exist error.
		// https://godoc.org/k8s.io/apimachinery/pkg/api/errors
		existing_secret, existing_secret_err := clientset.CoreV1().Secrets(maprticket.Namespace).Get(maprticket.Name, metav1.GetOptions{})
		if existing_secret_err == nil && existing_secret.GetObjectMeta().GetName() == maprticket.Name {
			log.Error(existing_secret_err, "Error creating secret for the MapRTicket. Secret with name "+maprticket.Name+" already exist.")
			// throw error
			return existing_secret, apierrors.NewAlreadyExists(schema.GroupResource{}, maprticket.Name + "Secret ")
		}
		/*secretKey := client.ObjectKey{Namespace: maprticket.Namespace, Name: maprticket.Name}
		if err := r.Get(ctx, secretKey, secret); err != nil {
			if !apierrors.IsNotFound(err) {
				return secret, errors.Wrap(err, "namespace already exists")
			}
		}*/
		log.Info("Secret - Secret object declaration.")
		secret = &apiv1.Secret{
			Type: apiv1.SecretTypeOpaque,
			ObjectMeta: metav1.ObjectMeta{
				Name:      maprticket.Name,
				Namespace: maprticket.Namespace,
				Labels:    maprticket.GetLabels(),
			},
		}
		if ticketstring != "" || len(ticketstring) != 0 {
			secret.Data = map[string][]byte{
				"CONTAINER_TICKET": []byte(ticketstring),
			}
		}
		/*if err := ctrl.SetControllerReference(maprticket, secret, r.Scheme); err != nil {
			return nil, err
		}*/
		log.Info("Secret - Creating secret object with corev1 api in namespace " + maprticket.Namespace)
		// secretOut, secret_err := clientset.CoreV1().Secrets(namespace).Create(secret)
		return secret, nil
	}
	// +kubebuilder:docs-gen:collapse=createMapRTicketSecret

	// Method to create MapRTicket and secret.
	createMapRTicket := func(maprticket *nscv1alpha1.MapRTicket) (*apiv1.Secret, error) {
		log.Info("createMapRTicket - creating mapr ticket.")
		log.Info("MapR - Connecting to MapR to create secret.")
		// write logic to connect to mapr and create ticket

		// Call create secret method to create secret.
		log.Info("Secret - Creating secret object with mapr ticket with name " + maprticket.Name + " in namespace " + maprticket.Namespace)
		var ticketstring = "bnNjc3RhZ2UwMS5lcmljc3Nvbi5jb20gWTBRK2RuS0VHQTdmT1czS0h2eDlsMnNmVVRvdDRQbnh4TzNmL2VLZ3dnemtXRnh6T1RGWWdFWXQ2d1RPOE9Bc2pPdVZnM1F1UUJPWWRIQU1tRTQzN0VZZHZLcVdJemliYW01dnFpekhMUE5DdWRSTnorZFhWNmZuYVRVOHdwa0NxOEgzQzltcytFdVVQVEJld2Y0b1ArZ0FWbDNrVGFncTIxQVFQNGkwSm0xbGpzNk1GZGRWVkxUcldaN0JKMG56WXJJdG5oN2ZlT0Z6aElJZmFWU05wVVdWMGxsa21mK1I3bVMxdG1Bbi9DUGxCOFprekRDK1lobGpmdCtWelJRWHNSakE0ZW96YVpYNEJiMEdsaFlibVhHajVPSE5EUFBGaE5WZmxWWlNqZ3ZFSTBjZXN2aG4zbkphZy9aTGMrTEFaeGJaU2MrbEFQREZBWnpoVDRxOTczNWZrM0pUZmo0MXlVOVNyZTEwCg=="
		secret, secret_err := createMapRTicketSecret(ticketstring, maprticket)
		if secret_err != nil {
			log.Error(secret_err, "Error creating secret for the MapRTicket secret object. ----------------------------")
			return secret, err
		}

		log.Info("Secret " + secret.Name + " Secret object created.")
		return secret, nil
	}
	// +kubebuilder:docs-gen:collapse=createMapRTicket

	// Creating ticket and secret
	secret, err := createMapRTicket(&maprTicket)
	if err != nil {
		log.Error(err, "unable to create MapRTicket secret.")
		return ctrl.Result{}, err
	}
	log.Info("Created Object of secret. Creating secret")
	if err := r.Create(ctx, secret); err != nil {
		log.Error(err, "unable to create MaprRTicket", "maprTicket", secret)
		return ctrl.Result{}, err
	}
	log.Info("Created MapRTicket.")
	return ctrl.Result{}, nil
}

/*
var (
	jobOwnerKey = ".metadata.controller"
	apiGVStr    = nscv1alpha1.GroupVersion.String()
)
*/

func (r *MapRTicketReconciler) SetupWithManager(mgr ctrl.Manager) error {

	/*
		// Adding current controller as owner
		if err := mgr.GetFieldIndexer().IndexField(&apiv1.Secret{}, jobOwnerKey, func(rawObj runtime.Object) []string {
			// grab the secret object, extract the owner...
			secret := rawObj.(*apiv1.Secret)
			owner := metav1.GetControllerOf(secret)
			if owner == nil {
				return nil
			}
			// ...make sure it's a MapRTicket...
			if owner.APIVersion != apiGVStr || owner.Kind != "MapRTicket" {
				return nil
			}

			// ...and if so, return it
			return []string{owner.Name}
		}); err != nil {
			return err
		}
	*/
	return ctrl.NewControllerManagedBy(mgr).
		For(&nscv1alpha1.MapRTicket{}).
		Owns(&apiv1.Secret{}).
		Complete(r)
}

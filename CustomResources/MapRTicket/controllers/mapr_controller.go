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
	apiv1 "k8s.io/api/core/v1"
	nscv1alpha1 "nsc/k8s/io/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *MapRTicketReconciler) createMapRTicket(req ctrl.Request, maprticket *nscv1alpha1.MapRTicket, secret *apiv1.Secret) error {
	log := r.requestLogger(req)
	log.Info("creating MapR Ticket for the user "  + maprticket.Spec.UserName)
	
	return nil
}

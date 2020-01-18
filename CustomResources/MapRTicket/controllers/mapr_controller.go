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
	"bytes"
	nscv1alpha1 "nsc/k8s/io/api/v1alpha1"
	"os"
	"os/exec"
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *MapRTicketReconciler) createMapRTicket(req ctrl.Request, maprticket *nscv1alpha1.MapRTicket) error {
	log := r.requestLogger(req)
	log.Info(string(maprticket.Spec.GroupID), maprticket.Spec.GroupName, string(maprticket.Spec.UserID), maprticket.Spec.UserName, "/tmp/maprticket_"+string(maprticket.Spec.UserID))
	log.Info("creating MapR Ticket for the user " + maprticket.Spec.UserName)
	log.Info("Setting environment variables for MAPR_CONTAINER_GID, MAPR_CONTAINER_GROUP, MAPR_CONTAINER_UID, MAPR_CONTAINER_USER and MAPR_TICKETFILE_LOCATION.")
	os.Setenv("MAPR_CONTAINER_GID", string(maprticket.Spec.GroupID))
	os.Setenv("MAPR_CONTAINER_GROUP", maprticket.Spec.GroupName)
	os.Setenv("MAPR_CONTAINER_UID", string(maprticket.Spec.UserID))
	os.Setenv("MAPR_CONTAINER_USER", maprticket.Spec.UserName)
	os.Setenv("MAPR_TICKETFILE_LOCATION", "/tmp/maprticket_"+string(maprticket.Spec.UserID))
	log.Info("Setting environment variables is done.")
	// Decrypt the password and setting as env var
	log.Info("Decrypting the password.")
	// TODO: Decrypt the password here

	password := maprticket.Spec.Password
	log.Info("Executing maprlogin command to create mapr ticket.")
	cmd := exec.Command("maprlogin", "password", "-user", maprticket.Spec.UserName)
	cmd.Stdin = strings.NewReader(password)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Error(err, "ERROR - Error executing mapr login password command.")
		return err
	}
	log.Info("MAPR LOGIN OUTPUT: " + out.String())
	// TODO: Check if invalid password or username or any errors in out.string.  or check if string contains MAPR_TICKETFILE_LOCATION values
	log.Info("INFO - MapR Ticket for user " + maprticket.Spec.UserName + " is generated successfully.")

	// Check if file /tmp/maprticket_MAPR_CONTAINER_UID exist
	log.Info("Checking if the ticket file exist.")
	fileCheck, fileCheckErr := os.Stat("/tmp/maprticket_" + string(maprticket.Spec.UserID))
	if os.IsNotExist(fileCheckErr) {
		log.Error(fileCheckErr, "ERROR - /tmp/maprticket_"+string(maprticket.Spec.UserID)+" File does not exist. Ticket Might not have been generated properly.")
		return fileCheckErr
	}
	if fileCheck.IsDir() {
		log.Error(fileCheckErr, "ERROR - /tmp/maprticket_"+string(maprticket.Spec.UserID)+" is a directory. Not a file. Plese check if mapr login is working properly.")
	}
	log.Info("/tmp/maprticket_" + string(maprticket.Spec.UserID) + " File exists.")

	// Fetching the ticket contents of mapr ticket file
	log.Info("Fetching the mapr ticket file contents.")
	fetchCmd := exec.Command("cat", "/tmp/maprticket_"+string(maprticket.Spec.UserID))
	var fetchOut bytes.Buffer
	fetchCmd.Stdout = &fetchOut
	fetchErr := fetchCmd.Run()
	if fetchErr != nil {
		log.Error(fetchErr, "ERROR - Error while fetching MapTicket file contents.")
		return fetchErr
	}
	log.Info("CAT " + "/tmp/maprticket_" + string(maprticket.Spec.UserID) + " is ran successfully.")
	log.Info("INFO - MapR Ticket for user " + maprticket.Spec.UserName + " is retrieved successfully.")
	log.Info("Updating MaprTicket in the Resource status.")
	maprticket.Status.MaprTicket = fetchOut.String()
	r.updateMapRTicketStatus(req, maprticket)
	log.Info("Updating the TicketInfo status of the MapRTicket.")
	maprticket.Status.TicketInfo = out.String()
	r.updateMapRTicketStatus(req, maprticket)
	// Unsetting the environment varibales
	log.Info("Unsetting the env variables.")
	os.Unsetenv("MAPR_CONTAINER_GID")
	os.Unsetenv("MAPR_CONTAINER_GROUP")
	os.Unsetenv("MAPR_CONTAINER_UID")
	os.Unsetenv("MAPR_CONTAINER_USER")
	os.Unsetenv("MAPR_TICKETFILE_LOCATION")
	log.Info("Unsetting env vars is done.")
	// remove the generated ticket file.
	var fileRemoveErr = os.Remove("/tmp/maprticket_" + string(maprticket.Spec.UserID))
	if fileRemoveErr != nil {
		log.Error(fileRemoveErr, "ERROR- Error while removing the ticket file.")
	}
	return nil
}

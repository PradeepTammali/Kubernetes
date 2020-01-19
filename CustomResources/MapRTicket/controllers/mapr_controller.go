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
	apiv1 "k8s.io/api/core/v1"
	nscv1alpha1 "nsc/k8s/io/api/v1alpha1"
	"os"
	"os/exec"
	ctrl "sigs.k8s.io/controller-runtime"
	"strconv"
	"strings"
)

func (r *MapRTicketReconciler) createMapRTicket(req ctrl.Request, maprticket *nscv1alpha1.MapRTicket) error {
	log := r.requestLogger(req)
	userID := strconv.FormatInt(maprticket.Spec.UserID, 10)
	userName := maprticket.Spec.UserName
	ticketFileLocation := "/tmp/maprticket_" + userID
	log.Info("creating MapR Ticket for the user " + userName)
	log.Info("Setting environment variables for MAPR_CONTAINER_GID, MAPR_CONTAINER_GROUP, MAPR_CONTAINER_UID, MAPR_CONTAINER_USER and MAPR_TICKETFILE_LOCATION.")
	os.Setenv("MAPR_CONTAINER_GID", strconv.FormatInt(maprticket.Spec.GroupID, 10))
	os.Setenv("MAPR_CONTAINER_GROUP", maprticket.Spec.GroupName)
	os.Setenv("MAPR_CONTAINER_UID", userID)
	os.Setenv("MAPR_CONTAINER_USER", userName)
	os.Setenv("MAPR_TICKETFILE_LOCATION", ticketFileLocation)
	log.Info("Setting environment variables is done.")

	// Decrypt the password and setting as env var
	log.Info("Decrypting the password.")
	// TODO: Decrypt the password here

	password := maprticket.Spec.Password
	// r.Recorder.Eventf(maprticket, apiv1.EventTypeNormal, "Creating", "Creating MapR Ticket.")
	log.Info("Executing maprlogin command to create mapr ticket.")
	cmd := exec.Command("maprlogin", "password", "-user", userName)
	cmd.Stdin = strings.NewReader(password)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Error(err, "ERROR - Error executing mapr login password command.")
		r.Recorder.Eventf(maprticket, apiv1.EventTypeWarning, "FailedCreate", "Error creating: MapRTicket \""+maprticket.Name+"\" can not be generated.")
		return err
	}
	log.Info("MAPR LOGIN OUTPUT: " + out.String())
	// TODO: Check if invalid password or username or any errors in out.string.  or check if string contains MAPR_TICKETFILE_LOCATION values
	log.Info("INFO - MapR Ticket for user " + userName + " is generated successfully.")

	// Check if file /tmp/maprticket_MAPR_CONTAINER_UID exist
	log.Info("Checking if the ticket file exist.")
	fileCheck, fileCheckErr := os.Stat(ticketFileLocation)
	if os.IsNotExist(fileCheckErr) {
		log.Error(fileCheckErr, "ERROR - /tmp/maprticket_"+userID+" File does not exist. Ticket Might not have been generated properly.")
		r.Recorder.Eventf(maprticket, apiv1.EventTypeWarning, "Failed", "TicketFetchErr: MapRTicket \""+maprticket.Name+"\" can not be retrieved.")
		return fileCheckErr
	}
	// Check if it a directory
	if fileCheck.IsDir() {
		log.Error(fileCheckErr, "ERROR - /tmp/maprticket_"+userID+" is a directory. Not a file. Plese check if mapr login is working properly.")
		r.Recorder.Eventf(maprticket, apiv1.EventTypeWarning, "Failed", "TicketFetchErr: MapRTicket \""+maprticket.Name+"\" can not be retrieved.")
		return os.ErrNotExist
	}
	log.Info(ticketFileLocation + " File exists.")

	// Fetching the ticket contents of mapr ticket file
	// r.Recorder.Eventf(maprticket, apiv1.EventTypeNormal, "Fetching", "Retrieving the MapR Ticket.")
	log.Info("Fetching the mapr ticket file contents.")
	fetchCmd := exec.Command("cat", ticketFileLocation)
	var fetchOut bytes.Buffer
	fetchCmd.Stdout = &fetchOut
	fetchErr := fetchCmd.Run()
	if fetchErr != nil {
		log.Error(fetchErr, "ERROR - Error while fetching MapTicket file contents.")
		r.Recorder.Eventf(maprticket, apiv1.EventTypeWarning, "Failed", "TicketFetchErr: MapRTicket \""+maprticket.Name+"\" can not be recovered.")
		return fetchErr
	}
	log.Info("CAT " + ticketFileLocation + " is ran successfully.")
	log.Info("INFO - MapR Ticket for user " + userName + " is retrieved successfully.")

	// Updating the status of the resource
	log.Info("Updating MaprTicket in the Resource status.")
	maprticket.Status.MaprTicket = fetchOut.String()
	r.updateMapRTicketStatus(req, maprticket)

	// Fetching ticket info
	// r.Recorder.Eventf(maprticket, apiv1.EventTypeNormal, "Printing", "Fetching MapR Ticket Info.")
	log.Info("Executing maprlogin print to fetch ticket information.")
	printCmd := exec.Command("maprlogin", "print", "-ticketfile", ticketFileLocation)
	var printOut bytes.Buffer
	printCmd.Stdout = &printOut
	printErr := printCmd.Run()
	if printErr != nil {
		log.Error(printErr, "ERROR - Error executing mapr login print command.")
		r.Recorder.Eventf(maprticket, apiv1.EventTypeWarning, "Failed", "TicketPrintErr: MapRTicket \""+maprticket.Name+"\" info can not be retrieved.")
		return printErr
	}
	log.Info("MAPRLOGIN PRINT OUTPUT: " + printOut.String())
	log.Info("INFO - MapR ticket info fetched successfully.")

	// Updating the status of the Resource.
	log.Info("Updating the MaprTicketInfo status of the MapRTicket.")
	maprticket.Status.MaprTicketInfo = printOut.String()
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
	var fileRemoveErr = os.Remove(ticketFileLocation)
	if fileRemoveErr != nil {
		log.Error(fileRemoveErr, "ERROR- Error while removing the ticket file.")
	}

	return nil
}

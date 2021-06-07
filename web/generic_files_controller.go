package web

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/APTrust/registry/common"
	"github.com/APTrust/registry/constants"
	"github.com/APTrust/registry/helpers"
	"github.com/APTrust/registry/pgmodels"
	"github.com/gin-gonic/gin"
)

// GenericFileRequestDelete shows a confirmation message asking
// if user really wants to delete a file.
// DELETE /files/request_delete/:id
func GenericFileRequestDelete(c *gin.Context) {
	req := NewRequest(c)
	gf, err := pgmodels.GenericFileByID(req.Auth.ResourceID)
	if AbortIfError(c, err) {
		return
	}
	req.TemplateData["file"] = gf
	req.TemplateData["error"] = err
	c.HTML(http.StatusOK, "files/_request_delete.html", req.TemplateData)
}

// GenericFileDelete deletes a user.
// DELETE /files/delete/:id
func GenericFileDelete(c *gin.Context) {

}

// GenericFileIndex shows list of objects.
// GET /files
func GenericFileIndex(c *gin.Context) {

}

// GenericFileShow returns the object with the specified id.
// GET /files/show/:id
func GenericFileShow(c *gin.Context) {
	req := NewRequest(c)
	file, err := pgmodels.GenericFileByID(req.Auth.ResourceID)
	req.TemplateData["file"] = file
	req.TemplateData["error"] = err
	c.HTML(http.StatusOK, "files/show.html", req.TemplateData)
}

// GenericFileRequestRestore shows a confirmation message asking whether
// user really wants to restore a file.
func GenericFileRequestRestore(c *gin.Context) {
	req := NewRequest(c)
	gf, err := pgmodels.GenericFileByID(req.Auth.ResourceID)
	if AbortIfError(c, err) {
		return
	}
	req.TemplateData["file"] = gf
	req.TemplateData["error"] = err
	c.HTML(http.StatusOK, "files/_request_restore.html", req.TemplateData)
}

// GenericFileInitRestore creates a file restoration request,
// which is really just a WorkItem that gets queued. Restoration can take
// seconds or hours, depending on where the file is stored and how big it is.
// POST /files/init_restore/:id
func GenericFileInitRestore(c *gin.Context) {
	req := NewRequest(c)
	gf, err := pgmodels.GenericFileByID(req.Auth.ResourceID)
	if AbortIfError(c, err) {
		return
	}

	// Make sure there are no pending work items...
	pendingWorkItems, err := pgmodels.WorkItemsPendingForFile(gf.ID)
	if AbortIfError(c, err) {
		return
	}
	if len(pendingWorkItems) > 0 {
		AbortIfError(c, common.ErrPendingWorkItems)
		return
	}

	// Create the new restoration work item
	obj, err := pgmodels.IntellectualObjectByID(gf.IntellectualObjectID)
	if AbortIfError(c, err) {
		return
	}
	workItem, err := pgmodels.NewRestorationItem(obj, gf, req.CurrentUser)
	if AbortIfError(c, err) {
		return
	}

	// Queue the new work item in NSQ
	topic, err := constants.TopicFor(workItem.Action, workItem.Stage)
	if AbortIfError(c, err) {
		return
	}
	ctx := common.Context()
	err = ctx.NSQClient.Enqueue(topic, workItem.ID)
	if AbortIfError(c, err) {
		return
	}

	workItem.QueuedAt = time.Now().UTC()
	err = workItem.Save()
	if AbortIfError(c, err) {
		return
	}

	// Respond
	msg := fmt.Sprintf("File %s has been queued for restoration.", gf.Identifier)
	helpers.SetFlashCookie(c, msg)
	redirectUrl := fmt.Sprintf("/objects/show/%d", obj.ID)
	c.Redirect(http.StatusFound, redirectUrl)
}

// GenericFileInitDelete occurs when user clicks the button confirming
// they want to delete a file. This creates a deletion confirmation message,
// which will be emailed to institutional admins for approval.
func GenericFileInitDelete(c *gin.Context) {
	req := NewRequest(c)

	del, err := NewDeletionForFile(req)
	if AbortIfError(c, err) {
		return
	}

	_, err = del.CreateRequestAlert()
	if AbortIfError(c, err) {
		return
	}

	req.TemplateData["fileIdentifier"] = del.DeletionRequest.FirstFile().Identifier
	c.HTML(http.StatusCreated, "files/deletion_requested.html", req.TemplateData)
}

// GenericFileReviewDelete displays a page on which an institutional
// admin can review a requested file deletion and choose whether to approve
// or cancel it.
// GET /files/review_delete/:id?token=<token>
func GenericFileReviewDelete(c *gin.Context) {
	req := NewRequest(c)

	// Find the deletion request
	deletionRequest, err := pgmodels.DeletionRequestByID(req.Auth.ResourceID)
	if AbortIfError(c, err) {
		return
	}

	// Make sure the token is valid for that deletion request
	token := c.Query("token")
	if !common.ComparePasswords(deletionRequest.EncryptedConfirmationToken, token) {
		AbortIfError(c, common.ErrInvalidToken)
		return
	}

	// Present the page describing the request, and if it hasn't
	// already been cancelled or approved, give the user the option
	// to cancel or approve.
	req.TemplateData["deletionRequest"] = deletionRequest
	req.TemplateData["token"] = token

	template := "files/review_deletion.html"
	if deletionRequest.ConfirmedByID > 0 {
		template = "files/deletion_already_approved.html"
	} else if deletionRequest.CancelledByID > 0 {
		template = "files/deletion_already_cancelled.html"
	}
	c.HTML(http.StatusOK, template, req.TemplateData)
}

// GenericFileApproveDelete handles the case where an institutional
// admin approves a file deletion request.
// POST /files/approve_delete/:id
func GenericFileApproveDelete(c *gin.Context) {
	req := NewRequest(c)

	del, err := NewDeletionForReview(req)
	if AbortIfError(c, err) {
		return
	}
	del.DeletionRequest.Confirm(req.CurrentUser)
	_, err = del.CreateAndQueueWorkItem()
	if AbortIfError(c, err) {
		return
	}

	_, err = del.CreateApprovalAlert()
	if AbortIfError(c, err) {
		return
	}

	req.TemplateData["deletionRequest"] = del.DeletionRequest
	c.HTML(http.StatusOK, "files/deletion_approved.html", req.TemplateData)
}

// GenericFileCancelDelete handles the case where an institutional
// admin cancels (rejects) a file deletion request.
// POST /files/cancel_delete/:id
func GenericFileCancelDelete(c *gin.Context) {
	req := NewRequest(c)

	// ---------------------------------------------------------------------
	// TODO: Refactor and break this up.
	// ---------------------------------------------------------------------

	// Find the deletion request
	deletionRequest, err := pgmodels.DeletionRequestByID(req.Auth.ResourceID)
	if AbortIfError(c, err) {
		return
	}

	// Make sure the token is valid for that deletion request
	token := c.PostForm("token")
	if !common.ComparePasswords(deletionRequest.EncryptedConfirmationToken, token) {
		AbortIfError(c, common.ErrInvalidToken)
		return
	}

	deletionRequest.CancelledByID = req.CurrentUser.ID
	deletionRequest.CancelledAt = time.Now().UTC()
	err = deletionRequest.Save()
	if AbortIfError(c, err) {
		return
	}

	// Explicitly set CancelledBy, because it wasn't set when we
	// loaded the DeletionRequest object.
	deletionRequest.CancelledBy = req.CurrentUser

	alertData := map[string]interface{}{
		"deletionRequest": deletionRequest,
	}
	tmpl := common.TextTemplates["alerts/deletion_cancelled.txt"]
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, alertData)
	if AbortIfError(c, err) {
		return
	}

	// Create deletion approved alert (refactor later)
	// Get list of inst admins.
	gf := deletionRequest.GenericFiles[0]
	adminsQuery := pgmodels.NewQuery().Where("institution_id", "=", gf.InstitutionID).Where("role", "=", constants.RoleInstAdmin)
	instAdmins, err := pgmodels.UserSelect(adminsQuery)
	if AbortIfError(c, err) {
		return
	}

	cancellationAlert := &pgmodels.Alert{
		InstitutionID:     deletionRequest.InstitutionID,
		Type:              constants.AlertDeletionCancelled,
		DeletionRequestID: deletionRequest.ID,
		Content:           buf.String(),
		CreatedAt:         time.Now().UTC(),
		Users:             instAdmins,
	}

	err = cancellationAlert.Save()
	if AbortIfError(c, err) {
		return
	}

	// Temporary local debugging
	envName := common.Context().Config.EnvName
	if envName == "dev" || envName == "test" {
		fmt.Println("***********************")
		fmt.Println(cancellationAlert.Content)
		fmt.Println("***********************")
	}

	req.TemplateData["deletionRequest"] = deletionRequest
	c.HTML(http.StatusOK, "files/deletion_cancelled.html", req.TemplateData)
}

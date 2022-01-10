package admin_api_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/APTrust/registry/constants"
	"github.com/APTrust/registry/db"
	"github.com/APTrust/registry/pgmodels"
	"github.com/APTrust/registry/web/api"
	tu "github.com/APTrust/registry/web/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenericFileShow(t *testing.T) {
	tu.InitHTTPTests(t)

	gf, err := pgmodels.GenericFileByID(1)
	require.Nil(t, err)
	require.NotNil(t, gf)

	// Sysadmin should be able to get this.
	// This is a pass-through to the common api endpoint,
	// but we want to make sure it's available at this URL.
	resp := tu.SysAdminClient.GET("/admin-api/v3/files/show/{id}", gf.ID).Expect().Status(http.StatusOK)
	record := &pgmodels.GenericFile{}
	err = json.Unmarshal([]byte(resp.Body().Raw()), record)
	require.Nil(t, err)
	assert.Equal(t, gf.ID, record.ID)
	assert.Equal(t, gf.InstitutionID, record.InstitutionID)

	// Non-admins should get an error. They have to go through
	// the member API.
	for _, client := range tu.AllClients {
		if client == tu.SysAdminClient {
			continue
		}
		tu.Inst2AdminClient.GET("/admin-api/v3/files/show/{id}", gf.ID).
			Expect().
			Status(http.StatusForbidden)
	}
}

func TestGenericFileIndex(t *testing.T) {
	tu.InitHTTPTests(t)

	// This endpoint should work for sys admin
	// but not for others.
	resp := tu.SysAdminClient.GET("/admin-api/v3/files").
		WithQuery("page", 2).
		WithQuery("per_page", 5).
		Expect().Status(http.StatusOK)

	list := api.GenericFileViewList{}
	err := json.Unmarshal([]byte(resp.Body().Raw()), &list)
	require.Nil(t, err)
	assert.Equal(t, 62, list.Count)
	assert.Equal(t, "/admin-api/v3/files?page=3&per_page=5", list.Next)
	assert.Equal(t, "/admin-api/v3/files?page=1&per_page=5", list.Previous)
	assert.Equal(t, tu.Inst1User.InstitutionID, list.Results[0].InstitutionID)

	// Test some filters. This object has 1 deleted, 4 active files.
	resp = tu.SysAdminClient.GET("/admin-api/v3/files").
		WithQuery("intellectual_object_id", 3).
		WithQuery("state", "A").
		Expect().Status(http.StatusOK)

	list = api.GenericFileViewList{}
	err = json.Unmarshal([]byte(resp.Body().Raw()), &list)
	require.Nil(t, err)
	assert.Equal(t, 4, list.Count)
	assert.Equal(t, 4, len(list.Results))
	for _, file := range list.Results {
		assert.Equal(t, int64(3), file.IntellectualObjectID)
		assert.Equal(t, "institution1.edu/glass", file.ObjectIdentifier)
		assert.Equal(t, "institution1.edu", file.InstitutionIdentifier)
		assert.Equal(t, constants.AccessConsortia, file.Access)
		assert.Equal(t, "A", file.State)
		assert.True(t, file.Size > 0)
	}

	// Non-admins are forbidden. They have to go through
	// the member API.
	for _, client := range tu.AllClients {
		if client == tu.SysAdminClient {
			continue
		}
		tu.Inst2AdminClient.GET("/admin-api/v3/files").
			Expect().
			Status(http.StatusForbidden)
	}
}

func TestFileCreateUpdateDelete(t *testing.T) {
	// Reset DB after this test so we don't screw up others.
	defer db.ForceFixtureReload()
	tu.InitHTTPTests(t)
	gf := testFileCreate(t)
	updatedFile := testFileUpdate(t, gf)

	createFileDeletionPreConditions(t, updatedFile)
	testFileDelete(t, updatedFile)
}

func testFileCreate(t *testing.T) *pgmodels.GenericFile {
	obj, err := pgmodels.IntellectualObjectGet(
		pgmodels.NewQuery().
			Where("institution_id", "=", 4).
			Limit(1))
	require.Nil(t, err)
	require.NotNil(t, obj)
	gf := pgmodels.RandomGenericFile(obj.ID, obj.Identifier)
	resp := tu.SysAdminClient.POST("/admin-api/v3/files/create/{id}", gf.InstitutionID).WithJSON(gf).Expect()
	resp.Status(http.StatusCreated)

	savedFile := &pgmodels.GenericFile{}
	err = json.Unmarshal([]byte(resp.Body().Raw()), savedFile)
	require.Nil(t, err)
	assert.True(t, savedFile.ID > int64(0))
	assert.Equal(t, gf.Identifier, savedFile.Identifier)
	assert.Equal(t, gf.InstitutionID, savedFile.InstitutionID)
	assert.Equal(t, gf.Size, savedFile.Size)
	assert.Equal(t, gf.FileFormat, savedFile.FileFormat)
	assert.Equal(t, gf.StorageOption, savedFile.StorageOption)
	assert.NotEmpty(t, savedFile.CreatedAt)
	assert.NotEmpty(t, savedFile.UpdatedAt)
	return savedFile
}

func testFileUpdate(t *testing.T, gf *pgmodels.GenericFile) *pgmodels.GenericFile {
	origUpdatedAt := gf.UpdatedAt
	copyOfGf := gf
	copyOfGf.Size = gf.Size + 200
	copyOfGf.FileFormat = "txt/screed"

	resp := tu.SysAdminClient.PUT("/admin-api/v3/files/update/{id}", gf.ID).WithJSON(copyOfGf).Expect()
	resp.Status(http.StatusOK)

	updatedGf := &pgmodels.GenericFile{}
	err := json.Unmarshal([]byte(resp.Body().Raw()), updatedGf)
	require.Nil(t, err)

	assert.Equal(t, copyOfGf.Size, updatedGf.Size)
	assert.Equal(t, copyOfGf.FileFormat, updatedGf.FileFormat)
	assert.Equal(t, gf.CreatedAt, updatedGf.CreatedAt)
	assert.True(t, updatedGf.UpdatedAt.After(origUpdatedAt))

	return updatedGf
}

// Registry business rules won't allow deletions without the following:
//
// - Ingest event at ingest.
// - Deletion request when a user clicks the delete file button
//   in the web UI.
// - WorkItem when an inst admin has approved the deletion request.
//
// Here, we create them just so we can complete our test.
func createFileDeletionPreConditions(t *testing.T, gf *pgmodels.GenericFile) {
	// Deletion checks for last ingest event on this object.
	event := pgmodels.RandomPremisEvent(constants.EventIngestion)
	event.IntellectualObjectID = gf.IntellectualObjectID
	event.GenericFileID = gf.ID
	event.GenericFileID = 0
	event.InstitutionID = gf.InstitutionID
	require.Nil(t, event.Save())

	// Also requires an approved Deletion work item
	item := pgmodels.RandomWorkItem(
		"TestBagName.tar",
		constants.ActionDelete,
		gf.IntellectualObjectID,
		gf.ID)
	item.User = "admin@test.edu"
	item.InstApprover = "admin@test.edu"
	item.Status = constants.StatusStarted
	require.Nil(t, item.Save())
	require.True(t, item.ID > 0)

	// Requires approved deletion request
	now := time.Now().UTC()
	req, err := pgmodels.NewDeletionRequest()
	require.Nil(t, err)
	req.GenericFiles = append(req.GenericFiles, gf)
	req.InstitutionID = gf.InstitutionID
	req.RequestedByID = 8 // admin@test.edu
	req.RequestedAt = now
	req.ConfirmedByID = 8
	req.ConfirmedAt = now
	req.WorkItemID = item.ID
	require.Nil(t, req.Save())
}

func testFileDelete(t *testing.T, gf *pgmodels.GenericFile) {
	resp := tu.SysAdminClient.DELETE("/admin-api/v3/files/delete/{id}", gf.ID).Expect()
	resp.Status(http.StatusOK)

	deletedFile := &pgmodels.GenericFile{}
	err := json.Unmarshal([]byte(resp.Body().Raw()), deletedFile)
	require.Nil(t, err)

	assert.Equal(t, gf.ID, deletedFile.ID)
	assert.Equal(t, constants.StateDeleted, deletedFile.State)

	// Test for deletion event
}

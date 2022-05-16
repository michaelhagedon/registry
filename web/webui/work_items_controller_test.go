package webui_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/APTrust/registry/common"
	"github.com/APTrust/registry/constants"
	"github.com/APTrust/registry/pgmodels"
	"github.com/APTrust/registry/web/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkItemShow(t *testing.T) {
	testutil.InitHTTPTests(t)

	items := []string{
		"1f594a4e5bb944e59c74aefe781a3726",
		"institution1.edu/photos",
		"aptrust.receiving.institution1.edu",
		"system@aptrust.org",
		"Delete",
		"Cleanup",
		"Success",
		"Item deleted successfuly",
	}

	adminActions := []string{
		"/work_items/edit/30",
		"/work_items/edit/30",
	}

	for _, client := range testutil.AllClients {
		html := client.GET("/work_items/show/30").Expect().
			Status(http.StatusOK).Body().Raw()
		testutil.AssertMatchesAll(t, html, items)
		if client == testutil.SysAdminClient {
			testutil.AssertMatchesAll(t, html, adminActions)
		} else {
			testutil.AssertMatchesNone(t, html, adminActions)
		}
	}
}

func TestWorkItemIndex(t *testing.T) {
	testutil.InitHTTPTests(t)

	links := []string{
		"/work_items/show/5",
		"/work_items/show/6",
		"/work_items/show/7",
		"/work_items/show/8",
	}

	commonFilters := []string{
		`select name="action"`,
		`select name="stage"`,
		`select name="status"`,
		`type="text" name="name"`,
		`type="text" name="etag"`,
		`type="date" name="date_processed__gteq"`,
		`type="date" name="date_processed__lteq"`,
		`select name="needs_admin_review"`,
		`type="text" name="object_identifier"`,
		`type="text" name="generic_file_identifier"`,
		`select name="storage_option"`,
		`type="text" name="alt_identifier"`,
		`type="text" name="bag_group_identifier"`,
		`type="text" name="user"`,
		`select name="bagit_profile_identifier"`,
		`type="number" name="size__gteq"`,
		`type="number" name="size__lteq"`,
	}

	adminFilters := []string{
		`select name="institution_id"`,
	}

	for _, client := range testutil.AllClients {
		html := client.GET("/work_items").Expect().
			Status(http.StatusOK).Body().Raw()
		testutil.AssertMatchesAll(t, html, links)
		testutil.AssertMatchesAll(t, html, commonFilters)
		if client == testutil.SysAdminClient {
			testutil.AssertMatchesAll(t, html, adminFilters)
			testutil.AssertMatchesResultCount(t, html, 33)
		} else {
			testutil.AssertMatchesNone(t, html, adminFilters)
			testutil.AssertMatchesResultCount(t, html, 18)
		}
	}

	// Apply a filter
	objRestorationLinks := []string{
		"/work_items/show/33",
	}
	for _, client := range testutil.AllClients {
		html := client.GET("/work_items").
			WithQuery("action", constants.ActionRestoreObject).
			Expect().
			Status(http.StatusOK).Body().Raw()
		testutil.AssertMatchesAll(t, html, objRestorationLinks)
	}

}

func TestWorkItemEditUpdate(t *testing.T) {
	testutil.InitHTTPTests(t)

	workItem := createWorkItem(t, "unit_test_bag1.tar")

	// Sys Admin should should be able to see the edit page for this item
	testutil.SysAdminClient.GET("/work_items/edit/{id}", workItem.ID).
		Expect().Status(http.StatusOK)

	// InstAdmin and InstUser cannot edit work items
	testutil.Inst1AdminClient.GET("/work_items/edit/{id}", workItem.ID).
		Expect().Status(http.StatusForbidden)
	testutil.Inst1UserClient.GET("/work_items/edit/{id}", workItem.ID).
		Expect().Status(http.StatusForbidden)

	// Change some values
	workItem.Stage = constants.StageStorageValidation
	workItem.Status = constants.StatusPending
	workItem.Retry = true
	workItem.NeedsAdminReview = false
	workItem.Note = "This has been edited"
	workItem.Node = ""
	workItem.PID = 0

	// SysAdmin should be able to PUT this
	testutil.SysAdminClient.PUT("/work_items/edit/{id}", workItem.ID).
		WithHeader("Referer", testutil.BaseURL).
		WithHeader(constants.CSRFHeaderName, testutil.SysAdminToken).
		WithForm(workItem).Expect().Status(http.StatusOK)

	// Make sure changes stuck.
	item, err := pgmodels.WorkItemByID(workItem.ID)
	require.Nil(t, err)
	require.NotNil(t, item)
	assert.Equal(t, workItem.Stage, item.Stage)
	assert.Equal(t, workItem.Status, item.Status)
	assert.Equal(t, workItem.Retry, item.Retry)
	assert.Equal(t, workItem.NeedsAdminReview, item.NeedsAdminReview)
	assert.Equal(t, workItem.Note, item.Note)
	assert.Equal(t, workItem.Node, item.Node)
	assert.Equal(t, workItem.PID, item.PID)

	// And make sure these roles cannot update work items
	testutil.Inst1AdminClient.PUT("/work_items/edit/{id}", workItem.ID).
		WithHeader("Referer", testutil.BaseURL).
		WithHeader(constants.CSRFHeaderName, testutil.Inst1AdminToken).
		WithForm(workItem).Expect().Status(http.StatusForbidden)
	testutil.Inst1UserClient.PUT("/work_items/edit/{id}", workItem.ID).
		WithHeader("Referer", testutil.BaseURL).
		WithHeader(constants.CSRFHeaderName, testutil.Inst1UserToken).
		WithForm(workItem).Expect().Status(http.StatusForbidden)
}

func TestWorkItemRequeue(t *testing.T) {
	testutil.InitHTTPTests(t)

	workItem := createWorkItem(t, "unit_test_bag2.tar")

	// SysAdmin can requeue
	testutil.SysAdminClient.PUT("/work_items/requeue/{id}", workItem.ID).
		WithHeader("Referer", testutil.BaseURL).
		WithHeader(constants.CSRFHeaderName, testutil.SysAdminToken).
		WithFormField("Stage", constants.StageReingestCheck).
		Expect().Status(http.StatusOK)

	// Make sure that worked. It should set not only the stage,
	// but the other values noted below.
	item, err := pgmodels.WorkItemByID(workItem.ID)
	require.Nil(t, err)
	require.NotNil(t, item)
	assert.Equal(t, constants.StageReingestCheck, item.Stage)
	assert.True(t, item.Retry)
	assert.False(t, item.NeedsAdminReview)
	assert.Empty(t, item.PID)
	assert.Empty(t, item.Node)
	assert.Equal(t, constants.StatusPending, item.Status)

	// Make sure other roles cannot requeue
	testutil.Inst1AdminClient.PUT("/work_items/requeue/{id}", workItem.ID).
		WithHeader("Referer", testutil.BaseURL).
		WithHeader(constants.CSRFHeaderName, testutil.Inst1AdminToken).
		WithFormField("Stage", constants.StageReingestCheck).
		Expect().Status(http.StatusForbidden)
	testutil.Inst1UserClient.PUT("/work_items/requeue/{id}", workItem.ID).
		WithHeader("Referer", testutil.BaseURL).
		WithHeader(constants.CSRFHeaderName, testutil.Inst1UserToken).
		WithFormField("Stage", constants.StageReingestCheck).
		Expect().Status(http.StatusForbidden)

}

/*
Note:

You can also test this manually by running redis-cli at the command line
and entering:

> hset "22" "object:institution1.edu/glass" '{"object": "value", "yadda": [1,2,3,4]}'

Then run ./registry serve, log in as sys admin and go to
http://localhost:8080/work_items/show/22
*/
func TestWorkItemRedisInfo(t *testing.T) {
	ctx := common.Context()

	items := []string{
		"obj_key1",
		"obj_value1",
		"Delete Redis Data",
	}

	itemID, _ := createRedisRecord(t, ctx)

	// Sys Admin should see this JSON from Redis
	html := testutil.SysAdminClient.GET("/work_items/show/{id}", itemID).
		Expect().
		Status(http.StatusOK).Body().Raw()
	testutil.AssertMatchesAll(t, html, items)

	// Other users should not.
	// Don't bother testing users outside of inst 1,
	// because they can't even view this item.
	html = testutil.Inst1UserClient.GET("/work_items/show/{id}", itemID).
		Expect().
		Status(http.StatusOK).Body().Raw()
	testutil.AssertMatchesNone(t, html, items)

	html = testutil.Inst1AdminClient.GET("/work_items/show/{id}", itemID).
		Expect().
		Status(http.StatusOK).Body().Raw()
	testutil.AssertMatchesNone(t, html, items)
}

func TestWorkItemRedisDelete(t *testing.T) {
	ctx := common.Context()

	items := []string{
		"Redis data for this work item has been deleted.",
	}

	itemID, objIdentifier := createRedisRecord(t, ctx)

	// Sys Admin should be able to delete this object
	// and should see confirmation that it was deleted.
	html := testutil.SysAdminClient.DELETE("/work_items/redis_delete/{id}", itemID).
		WithHeader("Referer", testutil.BaseURL).
		WithHeader(constants.CSRFHeaderName, testutil.SysAdminToken).
		Expect().
		Status(http.StatusOK).
		Body().Raw()
	testutil.AssertMatchesAll(t, html, items)

	// Check redis. Make sure the record is really gone.
	str, err := ctx.RedisClient.IngestObjectGet(itemID, objIdentifier)
	require.NotNil(t, err)
	require.Empty(t, str)

	// Recreate the Redis record and make sure non-admins
	// cannot delete it.
	itemID, objIdentifier = createRedisRecord(t, ctx)

	testutil.Inst1UserClient.DELETE("/work_items/redis_delete/{id}", itemID).
		WithHeader("Referer", testutil.BaseURL).
		WithHeader(constants.CSRFHeaderName, testutil.Inst1UserToken).
		Expect().
		Status(http.StatusForbidden)

	testutil.Inst1AdminClient.DELETE("/work_items/redis_delete/{id}", itemID).
		WithHeader("Referer", testutil.BaseURL).
		WithHeader(constants.CSRFHeaderName, testutil.Inst1AdminToken).
		Expect().
		Status(http.StatusForbidden)
}

func createRedisRecord(t *testing.T, ctx *common.APTContext) (int64, string) {
	// These items are in our fixture data.
	// Obj 3 is processed in work item 22.
	obj, err := pgmodels.IntellectualObjectByID(3)
	require.Nil(t, err)
	require.NotNil(t, obj)
	key := fmt.Sprintf("object:%s", obj.Identifier)
	err = ctx.RedisClient.SaveItem(22, key, `{"object":{"obj_key1":"obj_value1"}}`)
	require.Nil(t, err)

	// Make sure the data is there.
	str, err := ctx.RedisClient.IngestObjectGet(22, obj.Identifier)
	require.Nil(t, err)
	require.NotEmpty(t, str)

	return 22, obj.Identifier
}

func createWorkItem(t *testing.T, name string) *pgmodels.WorkItem {
	now := time.Now().UTC()
	workItem := &pgmodels.WorkItem{
		Name:             name,
		ETag:             "54321543215432154321000000000000",
		InstitutionID:    testutil.Inst1User.InstitutionID,
		Bucket:           "aptrust.receiving.yadda.yadda",
		User:             "system@aptrust.org",
		Note:             "Wheel her in, Homer! I'm not a picky man.",
		Action:           constants.ActionIngest,
		Stage:            constants.StageRecord,
		Status:           constants.StatusStarted,
		Outcome:          "Ourcome? I ain't done yet.",
		BagDate:          now,
		DateProcessed:    now,
		Retry:            false,
		Node:             "oh god, not Node!",
		PID:              3344,
		NeedsAdminReview: true,
		QueuedAt:         now,
		Size:             8900,
		StageStartedAt:   now,
	}
	require.Nil(t, workItem.Save())
	return workItem
}

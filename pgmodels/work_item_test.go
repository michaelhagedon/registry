package pgmodels_test

import (
	"testing"

	"github.com/APTrust/registry/common"
	"github.com/APTrust/registry/constants"
	"github.com/APTrust/registry/db"
	"github.com/APTrust/registry/pgmodels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkItemValidation(t *testing.T) {
	item := &pgmodels.WorkItem{}
	err := item.Validate()
	require.NotNil(t, err)

	assert.Equal(t, pgmodels.ErrItemName, err.Errors["Name"])
	assert.Equal(t, pgmodels.ErrItemETag, err.Errors["ETag"])
	assert.Equal(t, pgmodels.ErrItemBagDate, err.Errors["BagDate"])
	assert.Equal(t, pgmodels.ErrItemBucket, err.Errors["Bucket"])
	assert.Equal(t, pgmodels.ErrItemUser, err.Errors["User"])
	assert.Equal(t, pgmodels.ErrItemInstID, err.Errors["InstitutionID"])
	assert.Equal(t, pgmodels.ErrItemDateProcessed, err.Errors["DateProcessed"])
	assert.Equal(t, pgmodels.ErrItemNote, err.Errors["Note"])
	assert.Equal(t, pgmodels.ErrItemAction, err.Errors["Action"])
	assert.Equal(t, pgmodels.ErrItemStage, err.Errors["Stage"])
	assert.Equal(t, pgmodels.ErrItemStatus, err.Errors["Status"])
	assert.Equal(t, pgmodels.ErrItemOutcome, err.Errors["Outcome"])
}

func TestWorkItemGetID(t *testing.T) {
	item := &pgmodels.WorkItem{
		ID: 199,
	}
	assert.Equal(t, int64(199), item.GetID())
}

func TestWorkItemByID(t *testing.T) {
	db.LoadFixtures()
	item, err := pgmodels.WorkItemByID(int64(23))
	require.Nil(t, err)
	require.NotNil(t, item)
	assert.Equal(t, int64(23), item.ID)
}

func TestWorkItemGet(t *testing.T) {
	db.LoadFixtures()
	query := pgmodels.NewQuery().Where("name", "=", "pdfs.tar")
	item, err := pgmodels.WorkItemGet(query)
	require.Nil(t, err)
	require.NotNil(t, item)
	assert.Equal(t, "pdfs.tar", item.Name)
}

func TestWorkItemSelect(t *testing.T) {
	db.LoadFixtures()
	query := pgmodels.NewQuery()
	query.Where("name", "!=", "pdfs.tar")
	query.Where("name", "!=", "coal.tar")
	query.OrderBy("name asc")
	items, err := pgmodels.WorkItemSelect(query)
	require.Nil(t, err)
	require.NotEmpty(t, items)
	assert.True(t, (len(items) > 20 && len(items) < 30))
	for _, item := range items {
		assert.NotEqual(t, "pdfs.tar", item)
		assert.NotEqual(t, "coal.tar", item)
	}
}

func TestWorkItemSave(t *testing.T) {
	db.LoadFixtures()
	item := &pgmodels.WorkItem{
		Name:          "unit_00001.tar",
		ETag:          "12345678901234567890123456789099",
		InstitutionID: 4,
		User:          "system@aptrust.org",
		Bucket:        "aptrust.receiving.test.test.edu",
		Action:        constants.ActionIngest,
		Stage:         constants.StageRequested,
		Status:        constants.StatusPending,
		Note:          "Item is awaiting ingest.",
		Outcome:       "I said item is awaiting ingest.",
		BagDate:       TestDate,
		DateProcessed: TestDate,
		Retry:         true,
		Size:          8000,
	}
	err := item.Save()
	require.Nil(t, err)

	// pg library should set ID, BeforeInsert hook should set other values
	assert.True(t, item.ID > int64(0))
	assert.Equal(t, "unit_00001.tar", item.Name)
	assert.Equal(t, int64(4), item.InstitutionID)
	assert.NotEmpty(t, item.CreatedAt)
	assert.NotEmpty(t, item.UpdatedAt)
}

func TestWorkItemHasCompleted(t *testing.T) {
	item := &pgmodels.WorkItem{}
	for _, status := range constants.IncompleteStatusValues {
		item.Status = status
		assert.False(t, item.HasCompleted())
	}
	for _, status := range constants.CompletedStatusValues {
		item.Status = status
		assert.True(t, item.HasCompleted())
	}
}

func TestWorkItemSetForRequeue(t *testing.T) {
	db.LoadFixtures()
	item := &pgmodels.WorkItem{
		Name:          "unit_00002.tar",
		ETag:          "12345678901234567890123456789022",
		InstitutionID: 4,
		User:          "system@aptrust.org",
		Bucket:        "aptrust.receiving.test.test.edu",
		Action:        constants.ActionIngest,
		Stage:         constants.StageStore,
		Status:        constants.StatusStarted,
		Note:          "Item is being stored.",
		Outcome:       "I said item is being stored.",
		BagDate:       TestDate,
		DateProcessed: TestDate,
		Retry:         true,
		Size:          8000,
	}
	err := item.Save()
	require.Nil(t, err)

	err = item.SetForRequeue(constants.StageFormatIdentification)
	require.Nil(t, err)

	assert.Equal(t, constants.StageFormatIdentification, item.Stage)
	assert.Equal(t, constants.StatusPending, item.Status)
	assert.True(t, item.Retry)
	assert.False(t, item.NeedsAdminReview)
	assert.Empty(t, item.Node)
	assert.Empty(t, item.PID)
	assert.Equal(t, "Requeued for Format Identification", item.Note)

	// This should fail, because Package is not a valid stage for Ingest.
	err = item.SetForRequeue(constants.StagePackage)
	require.NotNil(t, err)
	assert.ErrorIs(t, err, common.ErrInvalidRequeue)
}

func TestWorkItemsPendingForObject(t *testing.T) {
	db.LoadFixtures()

	item := &pgmodels.WorkItem{
		Name:          "pending.tar",
		ETag:          "12345678901234567890123456789022",
		InstitutionID: 4,
		User:          "system@aptrust.org",
		Bucket:        "aptrust.receiving.test.test.edu",
		Action:        constants.ActionIngest,
		Stage:         constants.StageStore,
		Status:        constants.StatusStarted,
		Note:          "Item is being stored.",
		Outcome:       "I said item is being stored.",
		BagDate:       TestDate,
		DateProcessed: TestDate,
		Retry:         true,
		Size:          8000,
	}
	err := item.Save()
	require.Nil(t, err)

	// Should return nothing, because inst ID doesn't match.
	itemsInProgress, err := pgmodels.WorkItemsPendingForObject(3, "pending.tar")
	require.Nil(t, err)
	assert.Equal(t, 0, len(itemsInProgress))

	// This should get the item above
	itemsInProgress, err = pgmodels.WorkItemsPendingForObject(4, "pending.tar")
	require.Nil(t, err)
	assert.Equal(t, 1, len(itemsInProgress))

	item = itemsInProgress[0]
	item.Status = constants.StatusCancelled
	err = item.Save()
	require.Nil(t, err)

	// It should not come back this time because it has a completed status.
	itemsInProgress, err = pgmodels.WorkItemsPendingForObject(4, "pending.tar")
	require.Nil(t, err)
	assert.Equal(t, 0, len(itemsInProgress))

}

func TestWorkItemsPendingForFile(t *testing.T) {
	db.LoadFixtures()

}

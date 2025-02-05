package pgmodels_test

import (
	"testing"

	"github.com/APTrust/registry/constants"
	"github.com/APTrust/registry/db"
	"github.com/APTrust/registry/pgmodels"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntellectualObjectView(t *testing.T) {
	db.LoadFixtures()
	objView, err := pgmodels.IntellectualObjectViewByID(1)
	require.Nil(t, err)
	require.NotNil(t, objView)

	objView, err = pgmodels.IntellectualObjectViewByIdentifier(objView.Identifier)
	require.Nil(t, err)
	require.NotNil(t, objView)

	assert.Equal(t, int64(1), objView.GetID())

	query := pgmodels.NewQuery().
		Where("state", "=", constants.StateActive).
		OrderBy("created_at", "asc").
		Limit(10)
	objViews, err := pgmodels.IntellectualObjectViewSelect(query)
	require.Nil(t, err)
	assert.Equal(t, 10, len(objViews))
}

func TestSmallestObjectNotRestoredInXDays(t *testing.T) {
	db.LoadFixtures()
	defer db.ForceFixtureReload()

	// Ask for any object from this inst that hasn't been restored
	// in the past two years.
	obj, err := pgmodels.SmallestObjectNotRestoredInXDays(2, 200, 730)
	require.Nil(t, err)
	require.NotNil(t, obj)
	assert.Equal(t, int64(1), obj.ID)
	assert.Equal(t, int64(1657065000), obj.Size)

	// Ask for an object with a larger min size.
	obj, err = pgmodels.SmallestObjectNotRestoredInXDays(2, 1657065999, 730)
	require.Nil(t, err)
	require.NotNil(t, obj)
	assert.Equal(t, int64(3), obj.ID)
	assert.Equal(t, int64(13779270000), obj.Size)

	// Create a recent successful restoration for the object above.
	workItem := pgmodels.RandomWorkItem(obj.BagName, constants.ActionRestoreObject, obj.ID, 0)
	workItem.Stage = constants.StageResolve
	workItem.Status = constants.StatusSuccess
	require.Nil(t, workItem.Save())

	// Now query again, and we should get a different object
	// because the last one was recently restored.
	obj, err = pgmodels.SmallestObjectNotRestoredInXDays(2, 1657065999, 730)
	require.Nil(t, err)
	require.NotNil(t, obj)
	assert.Equal(t, int64(10), obj.ID)
	assert.Equal(t, int64(28234280000), obj.Size)
}

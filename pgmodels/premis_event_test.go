package pgmodels_test

import (
	"testing"
	"time"

	"github.com/APTrust/registry/constants"
	"github.com/APTrust/registry/db"
	"github.com/APTrust/registry/pgmodels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObjectEventCount(t *testing.T) {
	db.LoadFixtures()
	count, err := pgmodels.ObjectEventCount(3)
	require.Nil(t, err)
	assert.Equal(t, 1, count)

	// No events for non-existent object
	count, err = pgmodels.ObjectEventCount(0)
	require.Nil(t, err)
	assert.Equal(t, 0, count)

}

func TestIdForEventIdentifier(t *testing.T) {
	db.LoadFixtures()
	id, err := pgmodels.IdForEventIdentifier("77e16041-8887-4739-af04-9d35e5cab4dc")
	require.Nil(t, err)
	assert.Equal(t, int64(49), id)

	id, err = pgmodels.IdForEventIdentifier("274e230a-dc6b-48a1-a96c-709e5728632b")
	require.Nil(t, err)
	assert.Equal(t, int64(35), id)

	id, err = pgmodels.IdForEventIdentifier("bad identifier")
	require.NotNil(t, err)
}

func TestEventValidate(t *testing.T) {
	event := &pgmodels.PremisEvent{}
	valErr := event.Validate()
	require.NotNil(t, valErr)
	assert.Equal(t, 11, len(valErr.Errors))

	assert.NotEmpty(t, valErr.Errors["Agent"])
	assert.NotEmpty(t, valErr.Errors["DateTime"])
	assert.NotEmpty(t, valErr.Errors["Detail"])
	assert.NotEmpty(t, valErr.Errors["EventType"])
	assert.NotEmpty(t, valErr.Errors["Identifier"])
	assert.NotEmpty(t, valErr.Errors["InstitutionID"])
	assert.NotEmpty(t, valErr.Errors["IntellectualObjectID"])
	assert.NotEmpty(t, valErr.Errors["Object"])
	assert.NotEmpty(t, valErr.Errors["Outcome"])
	assert.NotEmpty(t, valErr.Errors["OutcomeDetail"])
	assert.NotEmpty(t, valErr.Errors["OutcomeInformation"])

	event.Agent = "Agent 99"
	event.DateTime = time.Now().UTC()
	event.Detail = "Some little detail"
	event.EventType = "*** Not a valid event type ***"
	event.Identifier = "*** Not a valid uuid ***"
	event.InstitutionID = 4
	event.IntellectualObjectID = 21
	event.Object = "The apple of my eye"
	event.Outcome = "*** Not a valid outcome ***"
	event.OutcomeDetail = "The proof is in the pudding"
	event.OutcomeInformation = "The info"

	valErr = event.Validate()
	require.NotNil(t, valErr)
	assert.Equal(t, 3, len(valErr.Errors))

	assert.Equal(t, "Event requires a valid EventType", valErr.Errors["EventType"])
	assert.Equal(t, "Event identifier should be a UUID", valErr.Errors["Identifier"])
	assert.Equal(t, "Event requires a valid Outcome value", valErr.Errors["Outcome"])

	// Make everything valid...
	event.EventType = constants.EventIngestion
	event.Identifier = "1ce24eee-1bf9-437f-99c8-166dcadcc190"
	event.Outcome = constants.OutcomeSuccess

	// ... and make sure the validator accepts it.
	valErr = event.Validate()
	require.Nil(t, valErr)
}

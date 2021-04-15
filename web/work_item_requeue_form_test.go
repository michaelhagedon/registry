package web_test

import (
	"testing"

	"github.com/APTrust/registry/common"
	"github.com/APTrust/registry/constants"
	"github.com/APTrust/registry/pgmodels"
	"github.com/APTrust/registry/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func WorkItemRequeueFormCompleted(t *testing.T) {
	query := pgmodels.NewQuery().Where("status", "=", constants.StageCleanup).Limit(1)
	items, err := pgmodels.WorkItemSelect(query)
	require.Nil(t, err)
	require.Equal(t, 1, len(items))
	item := items[0]
	req := &web.Request{
		ResourceID: item.ID,
	}
	_, err = web.NewWorkItemRequeueForm(req)
	require.NotNil(t, err)
	assert.Equal(t, common.ErrNotSupported, err)
}

func WorkItemRequeueFormIngest(t *testing.T) {
	query := pgmodels.NewQuery().Where("action", "=", constants.ActionIngest).Limit(1)
	items, err := pgmodels.WorkItemSelect(query)
	require.Nil(t, err)
	require.Equal(t, 1, len(items))
	item := items[0]

	// Save as new item with desired stage.
	// Receive is first stage, so we should only be able to requeue
	// to that.
	item.ID = 0
	item.Stage = constants.StageReceive
	err = item.Save()
	require.Nil(t, err)

	req := &web.Request{
		ResourceID: item.ID,
	}
	form, err := web.NewWorkItemRequeueForm(req)
	require.Nil(t, err)
	require.NotNil(t, form)
	require.NotNil(t, form.Fields["Stage"])
	require.Equal(t, 1, len(form.Fields["Stage"].Options))
	assert.Equal(t, constants.StageReceive, form.Fields["Stage"].Options[0].Value)

	// Store is sixth stage, so we should have six stage
	// options in the requeue list.
	item.Stage = constants.StageStore
	err = item.Save()
	require.Nil(t, err)

	req = &web.Request{
		ResourceID: item.ID,
	}
	form, err = web.NewWorkItemRequeueForm(req)
	require.Nil(t, err)
	require.NotNil(t, form)
	require.NotNil(t, form.Fields["Stage"])
	require.Equal(t, 6, len(form.Fields["Stage"].Options))
	opts := form.Fields["Stage"].Options
	assert.Equal(t, constants.StageReceive, opts[0].Value)
	assert.Equal(t, constants.StageValidate, opts[1].Value)
	assert.Equal(t, constants.StageReingestCheck, opts[2].Value)
	assert.Equal(t, constants.StageCopyToStaging, opts[3].Value)
	assert.Equal(t, constants.StageFormatIdentification, opts[4].Value)
	assert.Equal(t, constants.StageStore, opts[5].Value)

	// Cleanup is final stage, so all stage options
	// should appear.
	item.Stage = constants.StageCleanup
	err = item.Save()
	require.Nil(t, err)

	req = &web.Request{
		ResourceID: item.ID,
	}
	form, err = web.NewWorkItemRequeueForm(req)
	require.Nil(t, err)
	require.NotNil(t, form)
	require.NotNil(t, form.Fields["Stage"])
	require.Equal(t, 6, len(form.Fields["Stage"].Options))
	opts = form.Fields["Stage"].Options
	assert.Equal(t, constants.StageReceive, opts[0].Value)
	assert.Equal(t, constants.StageValidate, opts[1].Value)
	assert.Equal(t, constants.StageReingestCheck, opts[2].Value)
	assert.Equal(t, constants.StageCopyToStaging, opts[3].Value)
	assert.Equal(t, constants.StageFormatIdentification, opts[4].Value)
	assert.Equal(t, constants.StageStore, opts[5].Value)
	assert.Equal(t, constants.StageStorageValidation, opts[6].Value)
	assert.Equal(t, constants.StageRecord, opts[7].Value)
	assert.Equal(t, constants.StageCleanup, opts[8].Value)

}

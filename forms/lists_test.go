package forms_test

import (
	"fmt"
	"testing"

	"github.com/APTrust/registry/constants"
	"github.com/APTrust/registry/db"
	"github.com/APTrust/registry/forms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListInstitutions(t *testing.T) {
	db.LoadFixtures()
	options, err := forms.ListInstitutions(false)
	require.Nil(t, err)
	require.NotEmpty(t, options)
	assert.True(t, len(options) >= 4)
	expected := []forms.ListOption{
		{"1", "APTrust"},
		{"5", "Example Institution (for integration tests)"},
		{"2", "Institution One"},
		{"3", "Institution Two"},
		{"4", "Test Institution (for integration tests)"},
		{"6", "Unit Test Institution"},
	}
	for i, option := range options {
		assert.Equal(t, expected[i].Value, option.Value)
		assert.Equal(t, expected[i].Text, option.Text)
	}
}

func TestOptions(t *testing.T) {
	options := forms.Options(constants.AccessSettings)
	require.NotEmpty(t, options)
	for i, option := range options {
		assert.Equal(t, constants.AccessSettings[i], option.Value)
		assert.Equal(t, constants.AccessSettings[i], option.Text)
	}
}

func TestListUsers(t *testing.T) {
	db.LoadFixtures()
	options, err := forms.ListUsers(3)
	require.Nil(t, err)
	require.NotEmpty(t, options)
	assert.Equal(t, 2, len(options))
	expected := []forms.ListOption{
		{"5", "Inst Two Admin"},
		{"7", "Inst Two User"},
	}
	fmt.Println(options)
	for i, option := range options {
		assert.Equal(t, expected[i].Value, option.Value)
		assert.Equal(t, expected[i].Text, option.Text)
	}
}

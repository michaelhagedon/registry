package helpers_test

import (
	"html/template"
	"testing"
	"time"

	"github.com/APTrust/registry/constants"
	"github.com/APTrust/registry/helpers"
	"github.com/APTrust/registry/pgmodels"
	"github.com/stretchr/testify/assert"
)

var testDate, _ = time.Parse(time.RFC3339, "2021-04-16T15:04:05Z")
var textString = "The Academic Preservation Trust (APTrust) is committed to the creation and management of a sustainable environment for digital preservation."
var truncatedString = "The Academic Preservation Trust..."

func TestTruncate(t *testing.T) {
	assert.Equal(t, truncatedString, helpers.Truncate(textString, 31))
}

func TestDateUS(t *testing.T) {
	assert.Equal(t, "Apr 16, 2021", helpers.DateUS(testDate))
}

func TestDateISO(t *testing.T) {
	assert.Equal(t, "2021-04-16", helpers.DateISO(testDate))
}

func TestStrEq(t *testing.T) {
	assert.True(t, helpers.StrEq("4", int8(4)))
	assert.True(t, helpers.StrEq("200", int16(200)))
	assert.True(t, helpers.StrEq("200", int32(200)))
	assert.True(t, helpers.StrEq("200", int64(200)))

	assert.True(t, helpers.StrEq("true", true))
	assert.True(t, helpers.StrEq("true", "true"))
	assert.True(t, helpers.StrEq(true, true))
	assert.True(t, helpers.StrEq(true, "true"))
	assert.True(t, helpers.StrEq(false, "false"))

	assert.False(t, helpers.StrEq("true", false))
	assert.False(t, helpers.StrEq("200", 909))
}

func TestEscapeAttr(t *testing.T) {
	assert.Equal(t, template.HTMLAttr("O'Blivion's"), helpers.EscapeAttr("O'Blivion's"))
}

func TestEscapeHTML(t *testing.T) {
	assert.Equal(t, template.HTML("<em>escape!</em>"), helpers.EscapeHTML("<em>escape!</em>"))
}

func TestUserCan(t *testing.T) {
	admin := &pgmodels.User{
		Role:          constants.RoleSysAdmin,
		InstitutionID: 1,
	}
	assert.True(t, helpers.UserCan(admin, constants.UserCreate, 1))
	assert.True(t, helpers.UserCan(admin, constants.UserCreate, 2))
	assert.True(t, helpers.UserCan(admin, constants.UserCreate, 100))

	instAdmin := &pgmodels.User{
		Role:          constants.RoleInstAdmin,
		InstitutionID: 1,
	}
	assert.True(t, helpers.UserCan(instAdmin, constants.UserCreate, 1))
	assert.False(t, helpers.UserCan(instAdmin, constants.UserCreate, 2))
	assert.False(t, helpers.UserCan(instAdmin, constants.UserCreate, 100))

	instUser := &pgmodels.User{
		Role:          constants.RoleInstUser,
		InstitutionID: 1,
	}
	assert.False(t, helpers.UserCan(instUser, constants.UserCreate, 1))
	assert.False(t, helpers.UserCan(instUser, constants.UserCreate, 2))
	assert.False(t, helpers.UserCan(instUser, constants.UserCreate, 100))

}

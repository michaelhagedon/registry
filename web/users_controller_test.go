package web_test

import (
	"net/http"
	"testing"

	"github.com/APTrust/registry/constants"
	"github.com/APTrust/registry/pgmodels"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserShow(t *testing.T) {
	initHTTPTests(t)

	items := []string{
		"Force Password Reset",
		"Change Password",
		"Deactivate",
		"Edit",
		"Inst One User",
		"User at Institution One",
		"user@inst1.edu",
		"14345551212",
	}

	// Sys Admin can see any user
	html := sysAdminClient.GET("/users/show/3").Expect().
		Status(http.StatusOK).Body().Raw()
	AssertMatchesAll(t, html, items)

	// Inst admin can see users at their own institution
	html = instAdminClient.GET("/users/show/3").Expect().
		Status(http.StatusOK).Body().Raw()
	AssertMatchesAll(t, html, items)

	// Inst admin cannot view the user belonging to other institution
	instUserClient.GET("/users/show/1").Expect().Status(http.StatusForbidden)

	// Regular user cannot view the user show page, even their own
	instUserClient.GET("/users/show/3").Expect().Status(http.StatusForbidden)

}

func TestUserIndex(t *testing.T) {
	initHTTPTests(t)

	items := []string{
		"New",
		"Name",
		"Email",
	}

	instUserLinks := []string{
		"/users/show/2",
		"/users/show/3",
		"/users/show/4",
	}

	nonInst1Links := []string{
		"/users/show/1",
		"/users/show/5",
	}

	// Sys Admin sees filters because list of all users is long.
	// Inst admin does not see filters, because most institutions
	// have only 4-6 users.
	adminFilters := []string{
		`type="text" name="name__contains"`,
		`type="text" name="email__contains"`,
		`select name="role"`,
		`select name="institution_id"`,
		"Filter",
	}

	html := sysAdminClient.GET("/users").Expect().
		Status(http.StatusOK).Body().Raw()
	AssertMatchesAll(t, html, items)
	AssertMatchesAll(t, html, instUserLinks)
	AssertMatchesAll(t, html, nonInst1Links)
	AssertMatchesAll(t, html, adminFilters)
	AssertMatchesResultCount(t, html, 5)

	html = instAdminClient.GET("/users").Expect().
		Status(http.StatusOK).Body().Raw()
	AssertMatchesAll(t, html, items)
	AssertMatchesAll(t, html, instUserLinks)
	AssertMatchesNone(t, html, nonInst1Links)
	AssertMatchesNone(t, html, adminFilters)
	AssertMatchesResultCount(t, html, 3)

	// Regular user cannot view the user list page
	instUserClient.GET("/users").Expect().Status(http.StatusForbidden)

}

func TestUserCreateEditDeleteUndelete(t *testing.T) {
	initHTTPTests(t)

	// Make sure admins can get to this page and regular users cannot.
	sysAdminClient.GET("/users/new").Expect().Status(http.StatusOK)
	instAdminClient.GET("/users/new").Expect().Status(http.StatusOK)
	instUserClient.GET("/users/new").Expect().Status(http.StatusForbidden)

	formData := map[string]interface{}{
		"Name":           "Unit Test User",
		"Email":          "utest-user@inst1.edu",
		"PhoneNumber":    "+12025559815",
		"institution_id": inst1Admin.InstitutionID,
		"Role":           constants.RoleInstUser,
	}

	instAdminClient.POST("/users/new").
		WithForm(formData).Expect().Status(http.StatusOK)

	// Make sure the new user exists and has the correct info
	user, err := pgmodels.UserByEmail("utest-user@inst1.edu")
	require.Nil(t, err)
	require.NotNil(t, user)
	assert.Equal(t, formData["Name"], user.Name)
	assert.Equal(t, formData["Email"], user.Email)
	assert.Equal(t, formData["PhoneNumber"], user.PhoneNumber)
	assert.Equal(t, formData["institution_id"], user.InstitutionID)
	assert.Equal(t, formData["Role"], user.Role)
	assert.NotEmpty(t, user.EncryptedPassword)

	// Get the edit page for the new user
	instAdminClient.GET("/users/edit/{id}", user.ID).
		Expect().Status(http.StatusOK)

	// Update the user
	formData["Name"] = "Unit Test User (edited)"
	formData["PhoneNumber"] = "+15058981234"
	instAdminClient.PUT("/users/edit/{id}", user.ID).
		WithForm(formData).Expect().Status(http.StatusOK)

	// Make sure the edits were saved
	user, err = pgmodels.UserByEmail("utest-user@inst1.edu")
	require.Nil(t, err)
	require.NotNil(t, user)
	assert.Equal(t, formData["Name"], user.Name)
	assert.Equal(t, formData["PhoneNumber"], user.PhoneNumber)

	// Delete the user. This winds up with an OK because of redirect.
	instAdminClient.DELETE("/users/delete/{id}", user.ID).
		Expect().Status(http.StatusOK)

	// Undelete the user. Again, we get a redirect ending with an OK.
	instAdminClient.GET("/users/undelete/{id}", user.ID).
		Expect().Status(http.StatusOK)

}

func TestUserSignInSignOut(t *testing.T) {
	initHTTPTests(t)

	client := getAnonymousClient(t)

	// Make sure anonymous client can access the sign-in page
	client.GET("/").Expect().Status(http.StatusOK)

	// Make sure they can sign in and are redirected to dashboard
	html := client.POST("/users/sign_in").
		WithFormField("email", "user@inst1.edu").
		WithFormField("password", "password").
		Expect().Status(http.StatusOK).Body().Raw()

	dashboardItems := []string{
		"Recent Work Items",
		"Notifications",
		"Deposits by Storage Option",
	}
	AssertMatchesAll(t, html, dashboardItems)

	// Make sure user can sign out.
	client.GET("/users/sign_out").Expect().Status(http.StatusOK)

	// After signout, attempts to access valid pages should return
	// unauthorized.
	client.GET("/dashboard").Expect().Status(http.StatusUnauthorized)
}

func TestUserChangePassword(t *testing.T) {
	initHTTPTests(t)
}

func TestUserForcePasswordReset(t *testing.T) {
	initHTTPTests(t)
}

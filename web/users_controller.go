package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/APTrust/registry/common"
	"github.com/APTrust/registry/constants"
	"github.com/APTrust/registry/forms"
	"github.com/APTrust/registry/helpers"
	"github.com/APTrust/registry/pgmodels"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserCreate a new user. Handles submission of new user form.
// POST /users/new
func UserCreate(c *gin.Context) {
	saveUserForm(c)
}

// UserDelete deletes a user.
// DELETE or POST /users/delete/:id
func UserDelete(c *gin.Context) {
	req := NewRequest(c)
	user, err := pgmodels.UserByID(req.Auth.ResourceID)
	if AbortIfError(c, err) {
		return
	}
	err = user.Delete()
	if AbortIfError(c, err) {
		return
	}
	c.Redirect(http.StatusFound, "/users")
}

// UserUndelete reactivates a user.
// POST or PUT /users/undelete/:id
func UserUndelete(c *gin.Context) {
	req := NewRequest(c)
	user, err := pgmodels.UserByID(req.Auth.ResourceID)
	if AbortIfError(c, err) {
		return
	}
	err = user.Undelete()
	if AbortIfError(c, err) {
		return
	}
	location := fmt.Sprintf("/users/show/%d", user.ID)
	c.Redirect(http.StatusFound, location)
}

// UserIndex shows list of users.
// GET /users
func UserIndex(c *gin.Context) {
	req := NewRequest(c)
	template := "users/index.html"
	var users []*pgmodels.UserView
	err := req.LoadResourceList(&users, "name asc", forms.NewUserFilterForm)
	if AbortIfError(c, err) {
		return
	}
	c.HTML(http.StatusOK, template, req.TemplateData)
}

// UserNew returns a blank form for the user to create a new user.
// GET /users/new
func UserNew(c *gin.Context) {
	req := NewRequest(c)
	form, err := forms.NewUserForm(&pgmodels.User{}, req.CurrentUser)
	if AbortIfError(c, err) {
		return
	}
	req.TemplateData["form"] = form
	c.HTML(http.StatusOK, form.Template, req.TemplateData)
}

// UserShow returns the user with the specified id.
// GET /users/show/:id
func UserShow(c *gin.Context) {
	req := NewRequest(c)
	user, err := pgmodels.UserByID(req.Auth.ResourceID)
	if AbortIfError(c, err) {
		return
	}
	req.TemplateData["user"] = user
	c.HTML(http.StatusOK, "users/show.html", req.TemplateData)
}

// UserUpdate saves changes to an exiting user.
// PUT /users/edit/:id
func UserUpdate(c *gin.Context) {
	saveUserForm(c)
}

// UserEdit shows a form to edit an exiting user.
// GET /users/edit/:id
func UserEdit(c *gin.Context) {
	req := NewRequest(c)
	userToEdit, err := pgmodels.UserByID(req.Auth.ResourceID)
	if AbortIfError(c, err) {
		return
	}
	form, err := forms.NewUserForm(userToEdit, req.CurrentUser)
	if AbortIfError(c, err) {
		return
	}
	req.TemplateData["form"] = form
	c.HTML(http.StatusOK, form.Template, req.TemplateData)
}

// UserSignInShow shows the user sign-in form.
// GET /users/sign_in
func UserSignInShow(c *gin.Context) {
	c.HTML(200, "users/sign_in.html", gin.H{
		"cover":             helpers.GetCover(),
		"preFillTestLogins": common.Context().Config.EnvName == "test",
	})
}

// UserSignIn signs the user in.
// POST /users/sign_in
func UserSignIn(c *gin.Context) {
	status, redirectTo, err := SignInUser(c)
	if err == nil {
		c.Redirect(status, redirectTo)
	} else {
		c.HTML(status, "users/sign_in.html", gin.H{
			"error": err.Error(),
			"cover": helpers.GetCover(),
		})
	}
}

// UserSignOut signs the user out.
// GET /users/sign_out
func UserSignOut(c *gin.Context) {
	user := helpers.CurrentUser(c)
	if user != nil {
		user.SignOut()
	}
	helpers.DeleteSessionCookie(c)
	helpers.DeleteCSRFCookie(c)
	c.HTML(http.StatusOK, "users/sign_in.html", gin.H{
		"cover":             helpers.GetCover(),
		"preFillTestLogins": common.Context().Config.EnvName == "test",
	})
}

// UserShowChangePassword displays the change password page
// for the user with the specified ID.
//
// GET /users/change_password/:id
func UserShowChangePassword(c *gin.Context) {
	req, userToEdit, err := reqAndUserForPwdEdit(c)
	if AbortIfError(c, err) {
		return
	}
	form := forms.NewPasswordResetForm(userToEdit)
	req.TemplateData["form"] = form
	req.TemplateData["user"] = userToEdit

	// Not the prettiest solution, but for now, don't show
	// top and side nav if user is editing their own password.
	// On a forced password reset, we want the user to change
	// their password, not navigate to another page.
	//
	// This may annoy users changing their own password, and
	// it won't deter anyone from manually typing in a URL,
	// but it will suffice for now. No sense building in complex
	// logic now if ST is going to redo the UI anyway.
	// We'll come back to this one.
	if req.CurrentUser.ResetPasswordToken != "" {
		req.TemplateData["suppressTopNav"] = true
		req.TemplateData["suppressSideNav"] = true
	}

	c.HTML(http.StatusOK, form.Template, req.TemplateData)
}

// UserChangePassword changes a user's password. The user gets
// to specify what the new password will be.
//
// POST /users/change_password/:id
func UserChangePassword(c *gin.Context) {
	req, userToEdit, err := reqAndUserForPwdEdit(c)
	if AbortIfError(c, err) {
		return
	}
	pwd := c.PostForm("NewPassword")
	confirm := c.PostForm("ConfirmNewPassword")
	if pwd != confirm {
		err := fmt.Errorf("Passwords don't match.")
		AbortIfError(c, err)
		return
	}
	if !common.PasswordMeetsRequirements(pwd) {
		AbortIfError(c, common.ErrPasswordReqs)
		return
	}
	encPassword, err := common.EncryptPassword(pwd)
	if AbortIfError(c, err) {
		return
	}
	userToEdit.EncryptedPassword = encPassword
	userToEdit.PasswordChangedAt = time.Now().UTC()
	userToEdit.ResetPasswordToken = ""
	userToEdit.ResetPasswordSentAt = time.Time{}
	err = userToEdit.Save()
	if AbortIfError(c, err) {
		return
	}

	// Create a password changed alert, so we know this
	// happened and user knows too. If user gets a suspicious
	// "password changed" alert, they can contact us.
	_, err = CreatePasswordChangedAlert(req, userToEdit)

	helpers.SetFlashCookie(c, "Password changed.")
	redirectURL := fmt.Sprintf("/users/show/%d", userToEdit.ID)
	if !req.CurrentUser.HasPermission(constants.UserRead, userToEdit.InstitutionID) {
		redirectURL = "/dashboard"
	}
	c.Redirect(http.StatusFound, redirectURL)
}

func reqAndUserForPwdEdit(c *gin.Context) (*Request, *pgmodels.User, error) {
	req := NewRequest(c)
	userToEdit, err := pgmodels.UserByID(req.Auth.ResourceID)
	if err != nil {
		return req, nil, err
	}

	// Let's be real clear about the permissions governing who can
	// change a user's password.

	// Is the current user editing him/herself?
	isEditingSelf := req.CurrentUser.ID == req.Auth.ResourceID

	// Is current user an inst admin editing a user at their own institution?
	canEditInstUser := req.CurrentUser.HasPermission(constants.UserUpdate, userToEdit.InstitutionID)

	// Is current user sys admin?
	isSysAdmin := req.CurrentUser.IsAdmin()

	// Does the current user meet any of the three use cases above
	// that would allow them to change the password of the subject
	// user (userToEdit)?
	canEditSubject := isEditingSelf || canEditInstUser || isSysAdmin

	if !canEditSubject {
		return nil, nil, common.ErrPermissionDenied
	}

	return req, userToEdit, err
}

// UserInitPasswordReset resets a user's password to something
// random and sends them an email with a link that has an embedded
// login token. When they follow the link, they'll be automatically
// logged in and will have to choose a new password.
//
// TODO: This should be POST or PUT, not GET
//
// GET /users/init_password_reset/:id
func UserInitPasswordReset(c *gin.Context) {
	// This is admin triggering a password reset for another user,
	// so current user id does not need to match subject user id.
	req, userToEdit, err := reqAndUserForPwdEdit(c)
	if AbortIfError(c, err) {
		return
	}
	token := common.RandomToken()
	encryptedToken, err := common.EncryptPassword(token)
	if AbortIfError(c, err) {
		return
	}
	userToEdit.ResetPasswordToken = encryptedToken
	userToEdit.ForcePasswordUpdate = true
	err = userToEdit.Save()
	if AbortIfError(c, err) {
		return
	}
	_, err = CreatePasswordResetAlert(req, userToEdit, token)
	if AbortIfError(c, err) {
		return
	}

	req.TemplateData["user"] = userToEdit
	c.HTML(http.StatusOK, "users/reset_password_initiated.html", req.TemplateData)
}

// UserCompletePasswordReset allows a user to complete the password
// reset process. Note that this is one of the few pages that does
// not require a login.
//
// GET /users/complete_password_reset/:id
func UserCompletePasswordReset(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if userID == 0 || err != nil {
		AbortIfError(c, common.ErrInvalidParam)
		return
	}
	token := c.Query("token")
	if token == "" {
		AbortIfError(c, common.ErrInvalidToken)
		return
	}
	user, err := pgmodels.UserByID(userID)
	if AbortIfError(c, err) {
		return
	}
	if !user.DeactivatedAt.IsZero() {
		AbortIfError(c, common.ErrAccountDeactivated)
		return
	}
	if !common.ComparePasswords(user.ResetPasswordToken, token) {
		AbortIfError(c, common.ErrInvalidToken)
		return
	}

	user.SignInCount = user.SignInCount + 1
	if user.CurrentSignInIP != "" {
		user.LastSignInIP = user.CurrentSignInIP
	}
	if user.CurrentSignInAt.IsZero() {
		user.LastSignInAt = user.CurrentSignInAt
	}
	user.CurrentSignInIP = c.ClientIP()
	user.CurrentSignInAt = time.Now().UTC()
	err = user.Save()
	if AbortIfError(c, err) {
		return
	}

	err = helpers.SetSessionCookie(c, user)
	if AbortIfError(c, err) {
		return
	}
	c.Set("CurrentUser", user)
	pageURL := fmt.Sprintf("/users/change_password/%d", user.ID)
	c.Redirect(http.StatusFound, pageURL)
}

// UserGetAPIKey issues a new API key for the user, which replaces the
// user's existing API key. This key will be displayed once only.
//
// POST /users/get_api_key/:id
func UserGetAPIKey(c *gin.Context) {
	req := NewRequest(c)
	if req.CurrentUser.ID != req.Auth.ResourceID {
		common.Context().Log.Warn().Msgf("Permission denied: User %d requested API key for user %d", req.CurrentUser.ID, req.Auth.ResourceID)
		AbortIfError(c, common.ErrPermissionDenied)
	}
	apiKey := common.RandomToken()
	encKey, err := common.EncryptPassword(apiKey)
	if AbortIfError(c, err) {
		return
	}
	req.CurrentUser.EncryptedAPISecretKey = encKey
	err = req.CurrentUser.Save()
	if AbortIfError(c, err) {
		return
	}

	req.TemplateData["user"] = req.CurrentUser
	req.TemplateData["apiKey"] = apiKey
	c.HTML(http.StatusOK, "users/show_api_key.html", req.TemplateData)
}

// UserMyAccount displays the user's account info. From this page, they
// can see account details, change their password, and get an API key.
//
// GET /users/my_account
func UserMyAccount(c *gin.Context) {
	req := NewRequest(c)
	c.HTML(http.StatusOK, "users/my_account.html", req.TemplateData)
}

// SignInUser signs the user in with email and password.
// If user has two-factor auth, this is the first step of login.
func SignInUser(c *gin.Context) (int, string, error) {
	redirectTo := "/users/sign_in"
	user, err := pgmodels.UserSignIn(
		c.PostForm("email"),
		c.PostForm("password"),
		c.ClientIP(),
	)
	if err != nil {
		c.Error(err)
		helpers.DeleteSessionCookie(c)
		return http.StatusBadRequest, redirectTo, err
	}

	// Set this flag for two factor users.
	user.AwaitingSecondFactor = user.IsTwoFactorUser()

	err = helpers.SetSessionCookie(c, user)
	if err != nil {
		return http.StatusInternalServerError, redirectTo, err
	}
	err = helpers.SetCSRFCookie(c)
	if AbortIfError(c, err) {
		return http.StatusInternalServerError, redirectTo, err
	}
	c.Set("CurrentUser", user)

	redirectTo = "/dashboard"
	if user.IsTwoFactorUser() {
		redirectTo = "/users/choose_second_factor"
	}
	return http.StatusFound, redirectTo, nil
}

// UserTwoFactorChoose shows a list of radio button options so a user
// can choose their two-factor auth method (Authy, Backup Code, SMS).
// We show this page only to users who have enabled two-factor auth.
//
// GET /users/2fa_choose/
func UserTwoFactorChoose(c *gin.Context) {
	// Present the two-factor choice form.
	// Show only the options confirmed for current user
	//   - (sms and/or authy, plus backup codes)

	// If two-factor is SMS, generate token, send SMS and redirect
	// to a page where the user can enter the SMS OTP.
	//
	// If Authy, send approval request and wait for response.
	// Show a spinner. Or just Rickroll the bastards.
	// Authy is one-touch. We don't use the time-based codes.

}

// UserTwoFactorEnter shows a form with a single text input where a user
// can enter an SMS verfication code or a backup code.
//
// GET /users/2fa_enter/
func UserTwoFactorEnter(c *gin.Context) {
	// Show a form where user can enter SMS code or backup code.
	// Form should include the type user is entering (SMS/backup).
}

// UserTwoFactorPush initiates a push request to the user's authentication
// app asking them to approve the login. This method waits for a response
// from the authentication service. It's a POST to avoid GET spam.
// POST form includes CSRF token.
//
// POST /users/2fa_push/
func UserTwoFactorPush(c *gin.Context) {
	// On approval, redirect to dashboard.
	// On rejection or timeout, log user out and redirect to sign-in
}

// UserTwoFactorResend resends the SMS two-factor auth code and then
// re-displays TwoFactorEnter. This is a post, because we don't want
// hackers spamming us with GETs. The post form includes a CSRF token.
//
// POST /users/2fa_resend/
func UserTwoFactorResend(c *gin.Context) {

}

// UserTwoFactorVerify verifies the SMS or backup code that the user
// entered on TwoFactorEnter.
//
// POST /users/2fa_verify/
func UserTwoFactorVerify(c *gin.Context) {
	// If SMS, verify that, else verify backup.
	// Success redirects to dashboard.
	// Failure redirects to TwoFactorEnter,
	// incrementinf failed login attempt count
}

func getIndexQuery(c *gin.Context) (*pgmodels.Query, error) {
	allowedFilters := []string{
		"institution_id",
	}
	fc := pgmodels.NewFilterCollection()
	for _, key := range allowedFilters {
		fc.Add(key, c.QueryArray(key))
	}
	return fc.ToQuery()
}

func saveUserForm(c *gin.Context) {
	req := NewRequest(c)
	userToEdit := &pgmodels.User{}
	var err error
	if req.Auth.ResourceID > 0 {
		// Load existing user.
		userToEdit, err = pgmodels.UserByID(req.Auth.ResourceID)
		if AbortIfError(c, err) {
			return
		}
	} else {
		// Assign random password to new user. They'll get an email
		// asking them to reset their password.
		encPwd, err := common.EncryptPassword(uuid.New().String())
		if AbortIfError(c, err) {
			return
		}
		userToEdit.EncryptedPassword = encPwd
	}

	// Bind submitted form values in case we have to
	// re-display the form with an error message.
	c.ShouldBind(userToEdit)
	form, err := forms.NewUserForm(userToEdit, req.CurrentUser)
	if AbortIfError(c, err) {
		return
	}
	req.TemplateData["form"] = form
	if form.Save() {
		// If we just created a new user, send them a welcome email.
		if req.Auth.ResourceID == 0 {
			err = createNewUserAlert(req, userToEdit)
			if AbortIfError(c, err) {
				return
			}
		}
		c.Redirect(form.Status, form.PostSaveURL())
	} else {
		req.TemplateData["FormError"] = form.Error
		c.HTML(form.Status, form.Template, req.TemplateData)
	}
}

func createNewUserAlert(req *Request, newUser *pgmodels.User) error {
	token := common.RandomToken()
	encryptedToken, err := common.EncryptPassword(token)
	if err != nil {
		return err
	}
	newUser.ResetPasswordToken = encryptedToken
	newUser.ForcePasswordUpdate = true
	err = newUser.Save()
	if err != nil {
		return err
	}
	_, err = CreateNewAccountAlert(req, newUser, token)
	return err
}

package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/APTrust/registry/common"
	"github.com/APTrust/registry/pgmodels"
	"github.com/gin-gonic/gin"
)

// Authenticate eusures the current user is logged in for all requests other
// than those going to "/" or static resources.
func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		LoadCookies(c)
		var user *pgmodels.User
		var err error
		if !ExemptFromAuth(c) {
			user, err = GetUserFromSession(c)
			if err != nil {
				c.HTML(http.StatusUnauthorized, "errors/show.html", gin.H{
					"suppressSideNav": true,
					"suppressTopNav":  false,
					"error":           "Please log in",
					"redirectURL":     "/",
				})
				c.Abort()
			} else {
				c.Set("CurrentUser", user)
			}
		}
		if forceCompletionOfPasswordChange(c, user) {
			c.HTML(http.StatusUnauthorized, "errors/show.html", gin.H{
				"suppressSideNav": true,
				"suppressTopNav":  false,
				"error":           "Please finish changing your password",
			})
			c.Abort()
		}
		c.Next()
	}
}

// GetUserFromSession returns the User for the current session.
func GetUserFromSession(c *gin.Context) (user *pgmodels.User, err error) {
	ctx := common.Context()
	cookie, err := c.Cookie(ctx.Config.Cookies.SessionCookie)
	if err == nil {
		value := ""
		if err = ctx.Config.Cookies.Secure.Decode(ctx.Config.Cookies.SessionCookie, cookie, &value); err != nil {
			return nil, common.ErrDecodeCookie
		}
		userID, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, common.ErrWrongDataType
		}
		user, err = pgmodels.UserByID(userID)
	}
	return user, err
}

// LoadCookies loads the user's flash and preference cookes into
// the request context.
func LoadCookies(c *gin.Context) error {
	ctx := common.Context()
	err := LoadCookie(c, ctx.Config.Cookies.FlashCookie)
	if err != nil && err != http.ErrNoCookie {
		return err
	}
	err = LoadCookie(c, ctx.Config.Cookies.PrefsCookie)
	if err != nil && err != http.ErrNoCookie {
		return err
	}
	return nil
}

// LoadCookie loads a cookie's value into the request context.
func LoadCookie(c *gin.Context, name string) error {
	ctx := common.Context()
	cookie, err := c.Cookie(name)
	if err != nil {
		return err
	}
	value := ""
	if err = ctx.Config.Cookies.Secure.Decode(name, cookie, &value); err != nil {
		return common.ErrDecodeCookie
	}
	c.Set(name, value)
	return nil
}

func forceCompletionOfPasswordChange(c *gin.Context, currentUser *pgmodels.User) bool {
	if currentUser == nil {
		return false // user isn't even signed in
	}
	p := c.FullPath()
	return currentUser.ResetPasswordToken != "" &&
		!strings.HasPrefix(p, "/users/change_password/") &&
		!strings.HasPrefix(p, "/errors/show/")
}

func ExemptFromAuth(c *gin.Context) bool {
	p := c.FullPath()
	return p == "/" ||
		p == "/users/sign_in" ||
		p == "/users/sign_out" ||
		strings.HasPrefix(p, "/static") ||
		strings.HasPrefix(p, "/favicon") ||
		strings.HasPrefix(p, "/error") ||
		strings.HasPrefix(p, "/users/complete_password_reset/")
}

package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/APTrust/registry/common"
	"github.com/APTrust/registry/models"
	"github.com/gin-gonic/gin"
)

// Auth eusures the current user is logged in for all requests other
// than those going to "/" or static resources.
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ExemptFromAuth(c) {
			user, err := GetUserFromSession(c)
			if err != nil {
				c.AbortWithStatus(http.StatusUnauthorized)
			} else {
				c.Set("CurrentUser", user)
			}
		}
		c.Next()
	}
}

// GetUserFromSession returns the User for the current session.
func GetUserFromSession(c *gin.Context) (user *models.User, err error) {
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
		user, err = models.UserFind(userID)
	}
	return user, err
}

func ExemptFromAuth(c *gin.Context) bool {
	p := c.FullPath()
	return p == "/" || p == "/users/sign_in" || strings.HasPrefix(p, "/static") || strings.HasPrefix(p, "/favicon")
}

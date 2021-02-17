package controllers

import (
	"net/http"

	"github.com/APTrust/registry/common"
	"github.com/APTrust/registry/helpers"
	"github.com/gin-gonic/gin"
)

func ErrorShow(c *gin.Context) {
	templateData := helpers.TemplateVars(c)
	status := http.StatusInternalServerError
	err, _ := c.Get("err")
	if err != nil {
		templateData["error"] = err.(error).Error()
		status = StatusCodeForError(err.(error))
	}
	c.HTML(status, "errors/show.html", templateData)
}

// StatusCodeForError returns the http.StatusCode for the specified
// error. If the error doesn't map to a code, this returns 500 by
// default.
func StatusCodeForError(err error) (status int) {
	switch err {
	case common.ErrInvalidLogin:
		status = http.StatusUnauthorized
	case common.ErrAccountDeactivated:
		status = http.StatusForbidden
	case common.ErrPermissionDenied:
		status = http.StatusForbidden
	case common.ErrParentRecordNotFound:
		status = http.StatusNotFound
	case common.ErrWrongDataType:
		status = http.StatusBadRequest
	case common.ErrDecodeCookie:
		status = http.StatusBadRequest
	case common.ErrNotSupported:
		status = http.StatusMethodNotAllowed
	case common.ErrInternal:
		status = http.StatusInternalServerError
	default:
		status = http.StatusInternalServerError
	}
	return status
}

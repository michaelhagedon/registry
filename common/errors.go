package common

import (
	"errors"
	"fmt"
	"strings"
)

// ErrNotSignedIn means user has not signed in.
var ErrNotSignedIn = errors.New("user is not signed in")

// ErrInvalidLogin means the user supplied the wrong login name
// or password while trying to sign in.
var ErrInvalidLogin = errors.New("invalid login or password")

// ErrAccountDeactivated means the user logged in to a deactivated
// account.
var ErrAccountDeactivated = errors.New("account deactivated")

// ErrPermissionDenied means the user tried to access a resource
// withouth sufficient permission.
var ErrPermissionDenied = errors.New("permission denied")

// ErrNotSupported is an internal error that occurs, for example,
// when we try to delete an object that does not support deletion.
// This represents a programmer error and should not occur.
var ErrNotSupported = errors.New("operation not supported")

// ErrDecodeCookie occurs when we get a bad authentication cookie.
// We consider it a 400/Bad Request because we don't set bad cookies.
var ErrDecodeCookie = errors.New("error decoding cookie")

// ErrInvalidParam means the HTTP request contained an invalid
// parameter.
var ErrInvalidParam = errors.New("invalid parameter")

// ErrWrongDataType occurs when the user submits data of the wrong type,
// such string data that cannot be converted to a number, bool, date,
// or whatever type the application is expecting.
var ErrWrongDataType = errors.New("wrong data type")

// ErrParentRecordNotFound occurs when we cannot find the parent
// record required to check a user's permission. For example, when
// a user requests a Checksum, we first need to know if the user
// is allowed to access the Checksum's parent, which is a Generic
// File. If that record is missing, we get this error.
var ErrParentRecordNotFound = errors.New("parent record not found")

// ErrResourcePermission occurs when Authorization middleware cannot
// determine which permission is required to access the specified resoruce.
var ErrResourcePermission = errors.New("cannot determine permission type for requested requested resource")

// ErrInvalidRequeue occurs when someone attempts to requeue an item to the
// wrong stage, or to a stage for which no NSQ topic exists.
var ErrInvalidRequeue = errors.New("item cannot be requeued to the specified stage")

// ErrPendingWorkItems occurs when a user wants to restore or delete an
// object or file but the WorkItems list shows other operations are pending
// on that item. For example, we can't delete or restore an object or file
// while another version of that object/file is pending ingest. Doing so
// would cause newly ingested files to be deleted as soon as they're sent
// to preservation, or would cause a restoration to contain a mix of new
// and old versions of a bag's files.
var ErrPendingWorkItems = errors.New("task cannot be completed because this object has pending work items")

// ErrInvalidToken means that the token presented for an action like
// password reset or deletion confirmation does not match the encrypted
// token in the database. When this error occurs, the user may not
// proceed with the requested action.
var ErrInvalidToken = errors.New("invalid token")

// ErrPasswordReqs indicates that the password the user is trying to set
// does not meet minimum requirements.
var ErrPasswordReqs = errors.New("password does not meet minimum requirements")

// ErrInternal is a runtime error that is not the user's fault, hence
// probably the programmer's fault.
var ErrInternal = errors.New("internal server error")

type ValidationError struct {
	Errors map[string]string
}

func NewValidationError() *ValidationError {
	return &ValidationError{
		Errors: make(map[string]string),
	}
}

func (v *ValidationError) Error() string {
	if v.Errors == nil {
		return ""
	}
	errs := make([]string, len(v.Errors))
	i := 0
	for field, errMsg := range v.Errors {
		errs[i] = fmt.Sprintf("%s: %s", field, errMsg)
		i++
	}
	return strings.Join(errs, "\n")
}

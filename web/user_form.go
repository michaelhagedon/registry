package web

import (
	"strconv"

	"github.com/APTrust/registry/pgmodels"
)

type UserForm struct {
	Form
	instOptions []ListOption
}

func NewUserForm(request *Request) (*UserForm, error) {
	var err error
	user := &pgmodels.User{}
	if request.ResourceID > 0 {
		user, err = pgmodels.UserByID(request.ResourceID)
		if err != nil {
			return nil, err
		}
	}
	// Bind submitted form values in case we have to
	// re-display the form with an error message.
	request.GinContext.ShouldBind(user)

	userForm := &UserForm{
		Form: NewForm(request, user),
	}

	if request.CurrentUser.IsAdmin() {
		userForm.instOptions, err = ListInstitutions(false)
		if err != nil {
			return nil, err
		}
	} else {
		instID := strconv.Itoa(int(request.CurrentUser.InstitutionID))
		userForm.instOptions = []ListOption{
			{instID, request.CurrentUser.Institution.Name},
		}
		user.InstitutionID = request.CurrentUser.InstitutionID
	}
	userForm.init()
	userForm.setValues()
	return userForm, err
}

func (f *UserForm) init() {
	f.Fields["Name"] = &Field{
		Name:        "Name",
		ErrMsg:      pgmodels.ErrUserName,
		Label:       "Name",
		Placeholder: "Name",
		Attrs: map[string]string{
			"required": "",
		},
	}
	f.Fields["Email"] = &Field{
		Name:        "Email",
		ErrMsg:      pgmodels.ErrUserEmail,
		Label:       "Email Address",
		Placeholder: "Email Address",
		Attrs: map[string]string{
			"required": "",
		},
	}
	f.Fields["PhoneNumber"] = &Field{
		Name:        "PhoneNumber",
		ErrMsg:      pgmodels.ErrUserPhone,
		Label:       "Phone",
		Placeholder: "Phone in format 212-555-1212",
		Attrs:       map[string]string{
			//"pattern": "[0-9]{10,11}",
		},
	}
	f.Fields["OTPRequiredForLogin"] = &Field{
		Name:    "OTPRequiredForLogin",
		Label:   "Require Two-Factor Auth",
		ErrMsg:  pgmodels.ErrUser2Factor,
		Options: YesNoList,
		Attrs: map[string]string{
			"required": "",
		},
	}
	f.Fields["GracePeriod"] = &Field{
		Name:        "GracePeriod",
		Label:       "Must enable two-factor auth by",
		Placeholder: "mm/dd/yyyy",
		ErrMsg:      pgmodels.ErrUserGracePeriod,
		Attrs: map[string]string{
			"min": "2021-01-01",
			"max": "2099-12-31",
		},
	}
	f.Fields["InstitutionID"] = &Field{
		Name:    "InstitutionID",
		Label:   "Institution",
		ErrMsg:  pgmodels.ErrUserInst,
		Options: f.instOptions,
		Attrs: map[string]string{
			"required": "",
		},
	}
	rolesList := InstRolesList
	if f.Request.CurrentUser.IsAdmin() {
		rolesList = AllRolesList
	}
	f.Fields["Role"] = &Field{
		Name:    "Role",
		ErrMsg:  pgmodels.ErrUserRole,
		Label:   "Role",
		Options: rolesList,
		Attrs: map[string]string{
			"required": "",
		},
	}
}

// setValues sets the form values to match the User values.
func (f *UserForm) setValues() {
	user := f.Model.(*pgmodels.User)
	f.Fields["Name"].Value = user.Name
	f.Fields["Email"].Value = user.Email
	f.Fields["PhoneNumber"].Value = user.PhoneNumber
	f.Fields["OTPRequiredForLogin"].Value = user.OTPRequiredForLogin
	f.Fields["InstitutionID"].Value = user.InstitutionID
	f.Fields["Role"].Value = user.Role

	// Don't set date to 0001-01-01, because it makes
	// the date picker hard to use. User has to scroll
	// ahead 2000 years.
	if !user.GracePeriod.IsZero() {
		f.Fields["GracePeriod"].Value = user.GracePeriod.Format("2006-01-02")
	}
}

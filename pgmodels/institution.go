package pgmodels

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/APTrust/registry/common"
	"github.com/APTrust/registry/constants"
	v "github.com/asaskevich/govalidator"
	"github.com/go-pg/pg/v10"
)

const (
	ErrInstName       = "Name must contain 5-100 characters."
	ErrInstIdentifier = "Identifier must be a domain name."
	ErrInstState      = "State must be 'A' or 'D'."
	ErrInstType       = "Please choose an institution type."
	ErrInstReceiving  = "Receiving bucket name is not valid."
	ErrInstRestore    = "Restoration bucket name is not valid."
	ErrInstMemberID   = "Please choose a parent institution."
)

var InstitutionFilters = []string{
	"identifier__contains",
	"name__contains",
}

type Institution struct {
	ID                  int64     `json:"id" form:"id" pg:"id"`
	Name                string    `json:"name" pg:"name"`
	Identifier          string    `json:"identifier" pg:"identifier"`
	State               string    `json:"state" pg:"state"`
	Type                string    `json:"type" pg:"type"`
	MemberInstitutionID int64     `json:"member_institution_id" pg:"member_institution_id"`
	DeactivatedAt       time.Time `json:"deactivated_at" pg:"deactivated_at"`
	OTPEnabled          bool      `json:"otp_enabled" pg:"otp_enabled"`
	ReceivingBucket     string    `json:"receiving_bucket" pg:"receiving_bucket"`
	RestoreBucket       string    `json:"restore_bucket" pg:"restore_bucket"`
	CreatedAt           time.Time `json:"created_at" pg:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" pg:"updated_at"`
}

// InstitutionByID returns the institution with the specified id.
// Returns pg.ErrNoRows if there is no match.
func InstitutionByID(id int64) (*Institution, error) {
	query := NewQuery().Where("id", "=", id)
	return InstitutionGet(query)
}

// InstitutionByIdentifier returns the institution with the specified
// identifier. Returns pg.ErrNoRows if there is no match.
func InstitutionByIdentifier(identifier string) (*Institution, error) {
	query := NewQuery().Where("identifier", "=", identifier)
	return InstitutionGet(query)
}

// InstitutionGet returns the first institution matching the query.
func InstitutionGet(query *Query) (*Institution, error) {
	var institution Institution
	err := query.Select(&institution)
	return &institution, err
}

// InstitutionSelect returns all institutions matching the query.
func InstitutionSelect(query *Query) ([]*Institution, error) {
	var institutions []*Institution
	err := query.Select(&institutions)
	return institutions, err
}

func (inst *Institution) GetID() int64 {
	return inst.ID
}

// Save saves this institution to the database. This will peform an insert
// if Institution.ID is zero. Otherwise, it updates.
func (inst *Institution) Save() error {
	if inst.ID == int64(0) {
		return insert(inst)
	}
	return update(inst)
}

// Delete soft-deletes this institution by setting State to 'D' and
// the DeletedAt timestamp to now. You can undo this with Undelete.
func (inst *Institution) Delete() error {
	inst.State = constants.StateDeleted
	inst.DeactivatedAt = time.Now().UTC()
	return update(inst)
}

// Undelete reactivates this institution by setting State to 'A' and
// clearing the DeletedAt timestamp.
func (inst *Institution) Undelete() error {
	inst.State = constants.StateActive
	inst.DeactivatedAt = time.Time{}
	return update(inst)
}

// The following statements have no effect other than to force a compile-time
// check that ensures our Institution model properly implements these hook
// interfaces.
var (
	_ pg.BeforeInsertHook = (*Institution)(nil)
	_ pg.BeforeUpdateHook = (*Institution)(nil)
)

// BeforeInsert sets timestamps and bucket names on creation.
func (inst *Institution) BeforeInsert(c context.Context) (context.Context, error) {
	now := time.Now().UTC()
	inst.CreatedAt = now
	inst.UpdatedAt = now
	inst.ReceivingBucket = inst.bucket("receiving")
	inst.RestoreBucket = inst.bucket("restore")
	inst.State = constants.StateActive

	// WTF? The following:
	//
	// return c, inst.Validate()
	//
	// causes transaction to fail, even when inst.Validate() returns nil.
	// So we have to do this BS.
	err := inst.Validate()
	if err == nil {
		return c, nil
	}
	return c, err
}

// BeforeUpdate sets the UpdatedAt timestamp.
func (inst *Institution) BeforeUpdate(c context.Context) (context.Context, error) {
	inst.UpdatedAt = time.Now().UTC()
	err := inst.Validate()
	if err == nil {
		return c, nil
	}
	return c, err
}

// bucket returns a valid bucket name for this institution.
// Param name should be "receiving" or "restore"
func (inst *Institution) bucket(name string) string {
	ctx := common.Context()
	return fmt.Sprintf("aptrust.%s%s.%s", name, ctx.Config.BucketQualifier(), inst.Identifier)
}

// Validate validates the model. This is called automatically on insert
// and update.
func (inst *Institution) Validate() *common.ValidationError {
	errors := make(map[string]string)
	if !v.IsByteLength(inst.Name, 5, 200) {
		errors["Name"] = ErrInstName
	}
	// DNS names without dots, such as "localhost" are valid,
	// but we require a DNS name with at least one dot.
	if !v.IsDNSName(inst.Identifier) || !strings.Contains(inst.Identifier, ".") {
		errors["Identifier"] = ErrInstIdentifier
	}
	if !v.IsIn(inst.State, constants.States...) {
		errors["State"] = ErrInstState
	}
	if !v.IsIn(inst.Type, constants.InstTypes...) {
		errors["Type"] = ErrInstType
	}
	if inst.ReceivingBucket != inst.bucket("receiving") {
		errors["ReceivingBucket"] = ErrInstReceiving
	}
	if inst.RestoreBucket != inst.bucket("restore") {
		errors["RestoreBucket"] = ErrInstRestore
	}
	if inst.Type == constants.InstTypeSubscriber && inst.MemberInstitutionID < int64(1) {
		errors["MemberInstitutionID"] = ErrInstMemberID
	}
	if len(errors) > 0 {
		return &common.ValidationError{errors}
	}
	return nil
}

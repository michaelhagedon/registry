package pgmodels

import (
	"time"

	"github.com/APTrust/registry/common"
	"github.com/APTrust/registry/constants"
	"github.com/stretchr/stew/slice"
)

// Filters are defined in IntellectualObjectView, since we query on the view.

type IntellectualObject struct {
	ID                        int64     `json:"id"`
	Title                     string    `json:"title"`
	Description               string    `json:"description"`
	Identifier                string    `json:"identifier"`
	AltIdentifier             string    `json:"alt_identifier"`
	Access                    string    `json:"access"`
	BagName                   string    `json:"bag_name"`
	InstitutionID             int64     `json:"institution_id"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
	State                     string    `json:"state"`
	ETag                      string    `json:"etag" pg:"etag"`
	BagGroupIdentifier        string    `json:"bag_group_identifier"`
	StorageOption             string    `json:"storage_option"`
	BagItProfileIdentifier    string    `json:"bagit_profile_identifier" pg:"bagit_profile_identifier"`
	SourceOrganization        string    `json:"source_organization"`
	InternalSenderIdentifier  string    `json:"internal_sender_identifier"`
	InternalSenderDescription string    `json:"internal_sender_description"`

	Institution  *Institution   `json:"institution" pg:"rel:has-one"`
	GenericFiles []*GenericFile `json:"generic_files" pg:"rel:has-many"`
	PremisEvents []*PremisEvent `json:"premis_events" pg:"rel:has-many"`
}

// IntellectualObjectByID returns the object with the specified id.
// Returns pg.ErrNoRows if there is no match.
func IntellectualObjectByID(id int64) (*IntellectualObject, error) {
	query := NewQuery().Where(`"intellectual_object"."id"`, "=", id)
	return IntellectualObjectGet(query)
}

// IntellectualObjectByIdentifier returns the object with the specified
// identifier. Returns pg.ErrNoRows if there is no match.
func IntellectualObjectByIdentifier(identifier string) (*IntellectualObject, error) {
	query := NewQuery().Where(`"intellectual_object"."identifier"`, "=", identifier)
	return IntellectualObjectGet(query)
}

// IdForFileIdentifier returns the ID of the IntellectualObject
// having the specified identifier.
func IdForObjIdentifier(identifier string) (int64, error) {
	query := NewQuery().Columns("id").Where(`"intellectual_object"."identifier"`, "=", identifier)
	var object IntellectualObject
	err := query.Select(&object)
	return object.ID, err
}

// IntellectualObjectGet returns the first object matching the query.
func IntellectualObjectGet(query *Query) (*IntellectualObject, error) {
	var object IntellectualObject
	err := query.Relations("Institution").Select(&object)
	return &object, err
}

// IntellectualObjectSelect returns all objects matching the query.
func IntellectualObjectSelect(query *Query) ([]*IntellectualObject, error) {
	var objects []*IntellectualObject
	err := query.Select(&objects)
	return objects, err
}

func (obj *IntellectualObject) GetID() int64 {
	return obj.ID
}

// Save saves this object to the database. This will peform an insert
// if IntellectualObject.ID is zero. Otherwise, it updates.
func (obj *IntellectualObject) Save() error {
	if obj.ID == int64(0) {
		return insert(obj)
	}
	return update(obj)
}

// IsGlacierOnly returns true if this object is stored only
// in Glacier.
func (obj *IntellectualObject) IsGlacierOnly() bool {
	return isGlacierOnly(obj.StorageOption)
}

// Delete soft-deletes this object by setting State to 'D' and
// the DeletedAt timestamp to now. You can undo this with Undelete.
func (obj *IntellectualObject) Delete() error {
	obj.State = constants.StateDeleted
	obj.UpdatedAt = time.Now().UTC()

	// TODO: Create PremisEvents, update WorkItem

	return update(obj)
}

func isGlacierOnly(storageOption string) bool {
	return slice.Contains(constants.GlacierOnlyOptions, storageOption)
}

func (obj *IntellectualObject) ValidateChanges(updatedObj *IntellectualObject) error {
	if obj.ID != updatedObj.ID {
		return common.ErrIDMismatch
	}
	if obj.InstitutionID != updatedObj.InstitutionID {
		return common.ErrInstIDChange
	}
	if obj.Identifier != updatedObj.Identifier {
		return common.ErrIdentifierChange
	}
	// Caller should force storage option of updated object to
	// match existing object before calling this validation function.
	if obj.State == constants.StateActive && obj.StorageOption != updatedObj.StorageOption {
		return common.ErrStorageOptionChange
	}
	return nil
}

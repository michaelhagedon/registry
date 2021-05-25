package pgmodels

import (
	"time"

	"github.com/APTrust/registry/common"
	"github.com/APTrust/registry/constants"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

const (
	ErrDeletionInstitutionID = "Deletion request requires institution id."
	ErrDeletionRequesterID   = "Deletion request requires requester id."
	ErrDeletionWrongInst     = "Deletion request user belongs to wrong institution."
	ErrDeletionWrongRole     = "Deletion confirmer/canceller must be institutional admin."
	ErrDeletionUserNotFound  = "User does not exist."
	ErrDeletionUserInactive  = "User has been deactivated."
	ErrTokenNotEncrypted     = "Token must be encrypted."
)

// init does some setup work so go-pg can recognize many-to-many
// relations. Go automatically calls this function once when package
// is imported.
func init() {
	orm.RegisterTable((*DeletionRequestsGenericFiles)(nil))
	orm.RegisterTable((*DeletionRequestsIntellectualObjects)(nil))
}

type DeletionRequest struct {
	ID                         int64                 `json:"id"`
	InstitutionID              int64                 `json:"institution_id"`
	RequestedByID              int64                 `json:"-"`
	RequestedAt                time.Time             `json:"requested_at"`
	ConfirmationToken          string                `json:"-" pg:"-"`
	EncryptedConfirmationToken string                `json:"-"`
	ConfirmedByID              int64                 `json:"-"`
	ConfirmedAt                time.Time             `json:"confirmed_at"`
	CancelledByID              int64                 `json:"-"`
	CancelledAt                time.Time             `json:"cancelled_at"`
	RequestedBy                *User                 `json:"requested_by" pg:"rel:has-one"`
	ConfirmedBy                *User                 `json:"confirmed_by" pg:"rel:has-one"`
	CancelledBy                *User                 `json:"cancelled_by" pg:"rel:has-one"`
	GenericFiles               []*GenericFile        `json:"generic_files" pg:"many2many:deletion_requests_generic_files"`
	IntellectualObjects        []*IntellectualObject `json:"intellectual_objects" pg:"many2many:deletion_requests_intellectual_objects"`
}

type DeletionRequestsGenericFiles struct {
	tableName         struct{} `pg:"deletion_requests_generic_files"`
	DeletionRequestID int64
	GenericFileID     int64
}

type DeletionRequestsIntellectualObjects struct {
	tableName            struct{} `pg:"deletion_requests_intellectual_objects"`
	DeletionRequestID    int64
	IntellectualObjectID int64
}

func NewDeletionRequest() (*DeletionRequest, error) {
	confToken := common.RandomToken()
	encConfToken, err := common.EncryptPassword(confToken)
	if err != nil {
		return nil, err
	}
	return &DeletionRequest{
		ConfirmationToken:          confToken,
		EncryptedConfirmationToken: encConfToken,
		GenericFiles:               make([]*GenericFile, 0),
		IntellectualObjects:        make([]*IntellectualObject, 0),
	}, nil
}

// DeletionRequestByID returns the institution with the specified id.
// Returns pg.ErrNoRows if there is no match.
func DeletionRequestByID(id int64) (*DeletionRequest, error) {
	query := NewQuery().Relations("RequestedBy", "ConfirmedBy", "CancelledBy", "GenericFiles", "IntellectualObjects").Where(`"deletion_request"."id"`, "=", id)
	return DeletionRequestGet(query)
}

// DeletionRequestGet returns the first deletion request matching the query.
func DeletionRequestGet(query *Query) (*DeletionRequest, error) {
	var request DeletionRequest
	err := query.Select(&request)
	return &request, err
}

// DeletionRequestSelect returns all deletion requests matching the query.
func DeletionRequestSelect(query *Query) ([]*DeletionRequest, error) {
	var requests []*DeletionRequest
	err := query.Select(&requests)
	return requests, err
}

func (request *DeletionRequest) GetID() int64 {
	return request.ID
}

// Save saves this requestitution to the database. This will peform an insert
// if DeletionRequest.ID is zero. Otherwise, it updates.
func (request *DeletionRequest) Save() error {
	registryContext := common.Context()
	db := registryContext.DB
	return db.RunInTransaction(db.Context(), func(*pg.Tx) error {
		var err error
		if request.ID == 0 {
			_, err = db.Model(request).Insert()
		} else {
			_, err = db.Model(request).WherePK().Update()
		}
		if err != nil {
			registryContext.Log.Error().Msgf("Transaction failed. Model: %v. Error: %v", request, err)
		}
		return request.saveRelations(db)
	})
}

// Validation enforces business rules, including who can request and
// confirm deletions. Although our general security middleware should
// prevent any of these problems from ever occurring, we want to
// double check everything here because we're a preservation archive
// and deletion is a destructive action. We must be sure deletion is a
// deliberate act initiated and confirmed by authorized individuals.
func (request *DeletionRequest) Validate() *common.ValidationError {
	errors := make(map[string]string)

	if request.InstitutionID < 1 {
		errors["InstitutionID"] = ErrDeletionInstitutionID
	}

	// Make sure requester is valid
	if request.RequestedByID < 1 {
		errors["RequestedByID"] = ErrDeletionRequesterID
	}
	if request.RequestedByID > 0 && request.RequestedBy == nil {
		request.RequestedBy, _ = UserByID(request.RequestedByID)
	}
	if request.RequestedBy == nil {
		errors["RequestedByID"] = ErrDeletionRequesterID
	} else if request.RequestedBy.InstitutionID != request.InstitutionID {
		errors["RequestedByID"] = ErrDeletionWrongInst
	}

	// Make sure approver has admin role at the right institution
	if request.ConfirmedByID > 0 {
		if request.ConfirmedBy == nil {
			user, err := UserByID(request.ConfirmedByID)
			if err != nil {
				errors["ConfirmedByID"] = ErrDeletionUserNotFound
			}
			if user.InstitutionID != request.InstitutionID {
				errors["ConfirmedByID"] = ErrDeletionWrongInst
			}
			if user.Role != constants.RoleInstAdmin {
				errors["ConfirmedByID"] = ErrDeletionWrongRole
			}
		}
	}

	// Make sure canceller has admin role at the right institution
	if request.CancelledByID > 0 {
		if request.CancelledBy == nil {
			user, err := UserByID(request.CancelledByID)
			if err != nil {
				errors["CancelledByID"] = ErrDeletionUserNotFound
			}
			if user.InstitutionID != request.InstitutionID {
				errors["CancelledByID"] = ErrDeletionWrongInst
			}
			if user.Role != constants.RoleInstAdmin {
				errors["CancelledByID"] = ErrDeletionWrongRole
			}
		}
	}

	// Make sure tokens are actually encrypted
	if !common.LooksEncrypted(request.EncryptedConfirmationToken) {
		errors["EncryptedConfirmationToken"] = ErrTokenNotEncrypted
	}

	// TODO: Ensure that all objects and files actually belong to the
	// specified institution.

	if len(errors) > 0 {
		return &common.ValidationError{Errors: errors}
	}
	return nil
}

func (request *DeletionRequest) AddFile(gf *GenericFile) {
	if request.GenericFiles == nil {
		request.GenericFiles = make([]*GenericFile, 0)
	}
	request.GenericFiles = append(request.GenericFiles, gf)
}

func (request *DeletionRequest) AddObject(obj *IntellectualObject) {
	if request.IntellectualObjects == nil {
		request.IntellectualObjects = make([]*IntellectualObject, 0)
	}
	request.IntellectualObjects = append(request.IntellectualObjects, obj)
}

func (request *DeletionRequest) saveRelations(db *pg.DB) error {
	err := request.saveFiles(db)
	if err != nil {
		return err
	}
	return request.saveObjects(db)
}

func (request *DeletionRequest) saveFiles(db *pg.DB) error {
	sql := "insert into deletion_requests_generic_files (deletion_request_id, generic_file_id) values (?, ?) on conflict do nothing"
	for _, gf := range request.GenericFiles {
		_, err := db.Exec(sql, request.ID, gf.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (request *DeletionRequest) saveObjects(db *pg.DB) error {
	sql := "insert into deletion_requests_intellectual_objects (deletion_request_id, intellectual_object_id) values (?, ?) on conflict do nothing"
	for _, obj := range request.IntellectualObjects {
		_, err := db.Exec(sql, request.ID, obj.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

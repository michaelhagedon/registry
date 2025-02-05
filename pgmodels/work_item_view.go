package pgmodels

import (
	"fmt"
	"regexp"
	"time"

	"github.com/APTrust/registry/common"
	"github.com/APTrust/registry/constants"
	"github.com/stretchr/stew/slice"
)

var tarExtension = regexp.MustCompile("\\.tar$")

var WorkItemFilters = []string{
	"action",
	"action__in",
	"alt_identifier",
	"bag_date__gteq",
	"bag_date__lteq",
	"bag_group_identifier",
	"bagit_profile_identifier",
	"bucket",
	"date_processed__gteq",
	"date_processed__lteq",
	"etag",
	"generic_file_id",
	"generic_file_id__is_null",
	"generic_file_identifier",
	"institution_id",
	"intellectual_object_id",
	"intellectual_object_id__is_null",
	"name",
	"needs_admin_review",
	"node__is_null",
	"node__not_null",
	"object_identifier",
	"queued_at__is_null",
	"queued_at__not_null",
	"retry",
	"size__gteq",
	"size__lteq",
	"stage",
	"stage__in",
	"status",
	"status__in",
	"storage_option",
	"user",
}

// WorkItemView is a read-only model for querying. It flattens out
// WorkItem and some of its one-to-one relations for easy querying.
type WorkItemView struct {
	tableName                struct{}  `pg:"work_items_view"`
	ID                       int64     `json:"id" pg:"id"`
	Name                     string    `json:"name" pg:"name"`
	ETag                     string    `json:"etag" pg:"etag"`
	InstitutionID            int64     `json:"institution_id" pg:"institution_id"`
	InstitutionName          string    `json:"institution_name" pg:"institution_name"`
	InstitutionIdentifier    string    `json:"institution_identifier" pg:"institution_identifier"`
	IntellectualObjectID     int64     `json:"intellectual_object_id" pg:"intellectual_object_id"`
	ObjectIdentifier         string    `json:"object_identifier" pg:"object_identifier"`
	AltIdentifier            string    `json:"alt_identifier" pg:"alt_identifier"`
	BagGroupIdentifier       string    `json:"bag_group_identifier" pg:"bag_group_identifier"`
	StorageOption            string    `json:"storage_option" pg:"storage_option"`
	BagItProfileIdentifier   string    `json:"bagit_profile_identifier" pg:"bagit_profile_identifier"`
	SourceOrganization       string    `json:"source_organization" pg:"source_organization"`
	InternalSenderIdentifier string    `json:"internal_sender_identifier" pg:"internal_sender_identifier"`
	GenericFileID            int64     `json:"generic_file_id" pg:"generic_file_id"`
	GenericFileIdentifier    string    `json:"generic_file_identifier" pg:"generic_file_identifier"`
	Bucket                   string    `json:"bucket" pg:"bucket"`
	User                     string    `json:"user" pg:"user"`
	Note                     string    `json:"note" pg:"note"`
	Action                   string    `json:"action" pg:"action"`
	Stage                    string    `json:"stage" pg:"stage"`
	Status                   string    `json:"status" pg:"status"`
	Outcome                  string    `json:"outcome" pg:"outcome"`
	BagDate                  time.Time `json:"bag_date" pg:"bag_date"`
	DateProcessed            time.Time `json:"date_processed" pg:"date_processed"`
	Retry                    bool      `json:"retry" pg:"retry"`
	Node                     string    `json:"node" pg:"node"`
	PID                      int       `json:"pid" pg:"pid"`
	NeedsAdminReview         bool      `json:"needs_admin_review" pg:"needs_admin_review"`
	QueuedAt                 time.Time `json:"queued_at" pg:"queued_at"`
	Size                     int64     `json:"size" pg:"size"`
	StageStartedAt           time.Time `json:"stage_started_at" pg:"stage_started_at"`
	APTrustApprover          string    `json:"aptrust_approver" pg:"aptrust_approver"`
	InstApprover             string    `json:"inst_approver" pg:"inst_approver"`
	CreatedAt                time.Time `json:"created_at" pg:"created_at"`
	UpdatedAt                time.Time `json:"updated_at" pg:"updated_at"`
}

// WorkItemViewByID returns the work item with the specified id.
// Returns pg.ErrNoRows if there is no match.
func WorkItemViewByID(id int64) (*WorkItemView, error) {
	query := NewQuery().Where("id", "=", id)
	return WorkItemViewGet(query)
}

// WorkItemViewGet returns the first work item matching the query.
func WorkItemViewGet(query *Query) (*WorkItemView, error) {
	var item WorkItemView
	err := query.Select(&item)
	return &item, err
}

// WorkItemViewSelect returns all work items matching the query.
func WorkItemViewSelect(query *Query) ([]*WorkItemView, error) {
	var items []*WorkItemView
	err := query.Select(&items)
	return items, err
}

// GetID returns the ID of this WorkItem.
func (item *WorkItemView) GetID() int64 {
	return item.ID
}

// Validate is a no-op. This view is not writable, so we can't save to it.
// This method is here to satisfy the Model interface.
func (item *WorkItemView) Validate() *common.ValidationError {
	return nil
}

// HasCompleted returns true if this item has completed processing.
func (item *WorkItemView) HasCompleted() bool {
	return slice.Contains(constants.CompletedStatusValues, item.Status)
}

// ObjIdentifier returns this item's ObjectIdentifier if it exists, or
// what the object identifier would be if the item were fully ingested.
// Note that for ingest items, no object identifier is assigned until
// ingest is complete. However, to get a look at preservation services'
// internal Redis data, we need to know this not-yet-assigned identifier
// to compose the Redis key.
func (item *WorkItemView) GetObjIdentifier() string {
	if item.ObjectIdentifier != "" {
		return item.ObjectIdentifier
	}
	itemName := tarExtension.ReplaceAllString(item.Name, "")
	return fmt.Sprintf("%s/%s", item.InstitutionIdentifier, itemName)
}

// FindIngestedObject sounds like something the vet would do with your cat.
// It's actually used to fix a specific case where Registry thinks an ingest
// has failed, even though it actually succeeded. If we have to manually
// push the WorkItem to Ingest/Cleanup/Success, we need to link it to the
// ingested object. That object should match the identifier and etag of
// the WorkItem.
//
// We rarely need to do this. This case occurs only when preserv tries
// to re-record an already completed ingest and accidentally overwrites
// the WorkItem's object id. We're still trying to track down the conditions
// that lead to that event, but again, it's rare.
func (item *WorkItemView) FindIngestedObject() (*IntellectualObject, error) {
	query := NewQuery().Where("intellectual_object.identifier", "=", item.GetObjIdentifier()).Where("intellectual_object.etag", "=", item.ETag).Limit(1).Offset(0)
	return IntellectualObjectGet(query)
}

// IngestObjectLinkIsMissing returns true if this WorkItemView should
// have an associated IntellectualObjectID but does not have it.
func (item *WorkItemView) IngestObjectLinkIsMissing() bool {
	return item.Action == constants.ActionIngest && item.Status == constants.StatusSuccess && item.IntellectualObjectID == 0
}

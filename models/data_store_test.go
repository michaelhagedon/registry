package models_test

import (
	"testing"

	"github.com/APTrust/registry/common"
	"github.com/APTrust/registry/constants"
	"github.com/APTrust/registry/db"
	"github.com/APTrust/registry/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note that ds is defined in common_test.go and is created
// with SysAdmin user, so that ds has all privileges.

func TestChecksumFind(t *testing.T) {
	db.LoadFixtures()
	cs, err := ds.ChecksumFind(int64(1))
	require.Nil(t, err)
	require.NotNil(t, cs)
	assert.Equal(t, int64(1), cs.ID)
	assert.EqualValues(t, 1, cs.GenericFileID)
	assert.EqualValues(t, "md5", cs.Algorithm)
	assert.Equal(t, "12345678", cs.Digest)
}

func TestChecksumsList(t *testing.T) {
	db.LoadFixtures()
	query := models.NewQuery().Where("generic_file_id", "=", int64(21)).OrderBy("created_at desc").OrderBy("algorithm asc")
	checksums, err := ds.ChecksumList(query)
	require.Nil(t, err)
	require.NotEmpty(t, checksums)
	algs := []string{
		"md5",
		"sha1",
		"sha256",
		"sha512",
	}
	for i, cs := range checksums {
		assert.Equal(t, int64(21), cs.GenericFileID)
		assert.Equal(t, algs[i], checksums[i].Algorithm)
	}
}

func TestChecksumSave(t *testing.T) {
	db.LoadFixtures()
	cs := &models.Checksum{
		Algorithm:     constants.AlgMd5,
		DateTime:      TestDate,
		Digest:        "0987654321abcdef",
		GenericFileID: int64(20),
	}

	err := ds.ChecksumSave(cs)
	require.Nil(t, err)
	require.NotNil(t, cs)
	assert.True(t, cs.ID > int64(0))
	assert.EqualValues(t, 20, cs.GenericFileID)
	assert.EqualValues(t, constants.AlgMd5, cs.Algorithm)
	assert.Equal(t, "0987654321abcdef", cs.Digest)

	// We should get an error here because we're not allowed
	// to update existing checksums.
	cs.Digest = "----------------"
	err = ds.ChecksumSave(cs)
	require.NotNil(t, err)
	assert.Equal(t, common.ErrNotSupported, err)
}

func TestGenericFileSaveDeleteUndelete(t *testing.T) {
	db.LoadFixtures()
	gf := &models.GenericFile{
		FileFormat:           "text/plain",
		Size:                 int64(400),
		Identifier:           "institution2.edu/toads/test-file.txt",
		IntellectualObjectID: int64(6),
		State:                "A",
		LastFixityCheck:      TestDate,
		InstitutionID:        int64(3),
		StorageOption:        constants.StorageOptionStandard,
		UUID:                 "811b9a46-f91f-4379-a2f7-7b1bc8125a7c",
	}

	err := ds.GenericFileSave(gf)
	require.Nil(t, err)
	assert.True(t, gf.ID > int64(0))
	assert.Equal(t, "A", gf.State)
	assert.False(t, gf.CreatedAt.IsZero())
	assert.False(t, gf.UpdatedAt.IsZero())

	err = ds.GenericFileDelete(gf)
	require.Nil(t, err)
	assert.Equal(t, "D", gf.State)

	err = ds.GenericFileUndelete(gf)
	require.Nil(t, err)
	assert.Equal(t, "A", gf.State)
}

func TestGenericFileFind(t *testing.T) {
	db.LoadFixtures()
	gf, err := ds.GenericFileFind(int64(1))
	require.Nil(t, err)
	require.NotNil(t, gf)
	assert.Equal(t, int64(1), gf.ID)
	assert.Equal(t, "institution1.edu/photos/picture1", gf.Identifier)
	assert.Equal(t, int64(48771), gf.Size)
}

func TestGenericFileFindByIdentifier(t *testing.T) {
	db.LoadFixtures()
	gf, err := ds.GenericFileFindByIdentifier("institution1.edu/photos/picture1")
	require.Nil(t, err)
	require.NotNil(t, gf)
	assert.Equal(t, int64(1), gf.ID)
	assert.Equal(t, "institution1.edu/photos/picture1", gf.Identifier)
	assert.Equal(t, int64(48771), gf.Size)
}

func TestGenericFileList(t *testing.T) {
	db.LoadFixtures()
	query := models.NewQuery().Where("intellectual_object_id", "=", 1).OrderBy("identifier asc")
	files, err := ds.GenericFileList(query)
	require.Nil(t, err)

	expected := []string{
		"institution1.edu/photos/picture1",
		"institution1.edu/photos/picture2",
		"institution1.edu/photos/picture3",
	}
	assert.Equal(t, len(expected), len(files))
	for i, gf := range files {
		assert.Equal(t, expected[i], gf.Identifier)
	}
}

func TestInstitutionFind(t *testing.T) {
	db.LoadFixtures()
	inst, err := ds.InstitutionFind(int64(1))
	require.Nil(t, err)
	require.NotNil(t, inst)
	assert.Equal(t, int64(1), inst.ID)
	assert.Equal(t, "aptrust.org", inst.Identifier)
}

func TestInstitutionFindByIdentifier(t *testing.T) {
	db.LoadFixtures()
	inst, err := ds.InstitutionFindByIdentifier("aptrust.org")
	require.Nil(t, err)
	require.NotNil(t, inst)
	assert.Equal(t, int64(1), inst.ID)
	assert.Equal(t, "aptrust.org", inst.Identifier)
}

func TestInstitutionSaveDeleteUndelete(t *testing.T) {
	db.LoadFixtures()
	inst := &models.Institution{
		Name:            "Unit Test Institution",
		Identifier:      "unittest.edu",
		State:           "A",
		Type:            constants.InstTypeMember,
		ReceivingBucket: "aptrust.yadda.receiving.unittest.edu",
		RestoreBucket:   "aptrust.yadda.restore.unittest.edu",
	}
	err := ds.InstitutionSave(inst)
	require.Nil(t, err)
	assert.Equal(t, "A", inst.State)
	assert.True(t, inst.ID > int64(0))
	assert.False(t, inst.CreatedAt.IsZero())
	assert.False(t, inst.UpdatedAt.IsZero())
	assert.True(t, inst.DeactivatedAt.IsZero())

	err = ds.InstitutionDelete(inst)
	require.Nil(t, err)
	assert.Equal(t, "D", inst.State)
	assert.False(t, inst.DeactivatedAt.IsZero())

	err = ds.InstitutionUndelete(inst)
	require.Nil(t, err)
	assert.Equal(t, "A", inst.State)
	assert.True(t, inst.DeactivatedAt.IsZero())
}

func TestInstitutionList(t *testing.T) {
	db.LoadFixtures()
	query := models.NewQuery().Where("identifier", "LIKE", "%.edu").Where("state", "=", "A").OrderBy("name asc")
	institutions, err := ds.InstitutionList(query)
	require.Nil(t, err)

	expected := []string{
		"Example Institution (for integration tests)",
		"Institution One",
		"Institution Two",
		"Test Institution (for integration tests)",
		"Unit Test Institution",
	}
	assert.Equal(t, len(expected), len(institutions))
	for i, inst := range institutions {
		assert.Equal(t, expected[i], inst.Name)
	}
}

func TestIntellectualObjectFind(t *testing.T) {
	db.LoadFixtures()
	obj, err := ds.IntellectualObjectFind(int64(1))
	require.Nil(t, err)
	require.NotNil(t, obj)
	assert.Equal(t, int64(1), obj.ID)
	assert.Equal(t, "institution1.edu/photos", obj.Identifier)
	assert.Equal(t, "First Object for Institution One", obj.Title)
}

func TestIntellectualObjectFindByIdentifier(t *testing.T) {
	db.LoadFixtures()
	obj, err := ds.IntellectualObjectFindByIdentifier("institution1.edu/photos")
	require.Nil(t, err)
	require.NotNil(t, obj)
	assert.Equal(t, int64(1), obj.ID)
	assert.Equal(t, "institution1.edu/photos", obj.Identifier)
	assert.Equal(t, "First Object for Institution One", obj.Title)
}

func TestIntellectualObjectSaveDeleteUndelete(t *testing.T) {
	obj := &models.IntellectualObject{
		Title:                     "Unit Test Bag #100",
		Description:               "Bag created during unit tests",
		Identifier:                "institution1.edu/UnitTestBag100",
		AltIdentifier:             "Alt Identifier 100",
		Access:                    constants.AccessInstitution,
		BagName:                   "UnitTestBag.tar",
		InstitutionID:             2,
		State:                     "A",
		ETag:                      "etag-phone-home",
		BagGroupIdentifier:        "unit-test-group",
		StorageOption:             constants.StorageOptionWasabiVA,
		BagItProfileIdentifier:    constants.DefaultProfileIdentifier,
		SourceOrganization:        "UVA",
		InternalSenderIdentifier:  "test-internal-id",
		InternalSenderDescription: "test-internal-desc",
	}
	err := ds.IntellectualObjectSave(obj)
	require.Nil(t, err)
	assert.True(t, obj.ID > int64(0))
	assert.Equal(t, "UnitTestBag.tar", obj.BagName)
	assert.Equal(t, "A", obj.State)
	assert.False(t, obj.CreatedAt.IsZero())
	assert.False(t, obj.UpdatedAt.IsZero())

	err = ds.IntellectualObjectDelete(obj)
	assert.Equal(t, "D", obj.State)

	err = ds.IntellectualObjectUndelete(obj)
	assert.Equal(t, "A", obj.State)
}

func TestIntellectualObjectList(t *testing.T) {
	db.LoadFixtures()
}

func TestPremisEventFind(t *testing.T) {
	event, err := ds.PremisEventFind(int64(1))
	require.Nil(t, err)
	require.NotNil(t, event)
	assert.Equal(t, int64(1), event.ID)
	assert.EqualValues(t, 14, event.GenericFileID)
	assert.EqualValues(t, 3, event.InstitutionID)
	assert.Equal(t, "a966ca54-ee5b-4606-81bd-7653dd5f3a63", event.Identifier)
}

func TestPremisEventFindByIdentifier(t *testing.T) {
	event, err := ds.PremisEventFindByIdentifier("a966ca54-ee5b-4606-81bd-7653dd5f3a63")
	require.Nil(t, err)
	require.NotNil(t, event)
	assert.Equal(t, int64(1), event.ID)
	assert.EqualValues(t, 14, event.GenericFileID)
	assert.EqualValues(t, 3, event.InstitutionID)
	assert.Equal(t, "a966ca54-ee5b-4606-81bd-7653dd5f3a63", event.Identifier)
}

func TestPremisEventList(t *testing.T) {
	query := models.NewQuery().Where("generic_file_id", "=", int64(3)).OrderBy("event_type asc", "date_time asc")
	events, err := ds.PremisEventList(query)
	require.Nil(t, err)
	require.NotNil(t, events)
	expected := []string{
		"d1dd9047-d25c-4ba3-adc4-e17914eda1e9", // ingestion
		"6e9e665a-4f7e-41f4-9594-d511f9fc1edf", // ingestion
		"549a9b7f-3a61-42b3-8af4-13d01ef13f41", // message digest calculation
		"3bd67ede-0fca-430a-9bb3-652c0a95b471", // message digest calculation
	}
	assert.Equal(t, len(expected), len(events))
	for i, event := range events {
		assert.Equal(t, expected[i], event.Identifier)
	}
}

func TestPremisEventSave(t *testing.T) {
	event := &models.PremisEvent{
		Identifier:           "",
		EventType:            constants.EventDecryption,
		DateTime:             TestDate,
		OutcomeDetail:        "Pistol whip? I don't like the sound of that!",
		Detail:               "Mmm! Pistol whip!",
		Object:               "Duff",
		Agent:                "Moe",
		IntellectualObjectID: int64(1),
		GenericFileID:        int64(20),
		Outcome:              "Doh!",
		InstitutionID:        int64(4),
	}
	err := ds.PremisEventSave(event)
	require.Nil(t, err)
	assert.True(t, event.ID > int64(0))

	// This should cause an error because updating events
	// is not allowed.
	event.Outcome = "Ooh!"
	err = ds.PremisEventSave(event)
	require.NotNil(t, err)
	assert.Equal(t, common.ErrNotSupported, err)
}

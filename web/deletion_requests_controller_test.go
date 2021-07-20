package web_test

import (
	"net/http"
	"testing"
)

func TestDeletionRequestShow(t *testing.T) {
	initHTTPTests(t)

	items := []string{
		"Deletion Request",
		"Requested By",
		"Requested At",
		"Work Item",
		"Files",
		"institution1.edu/glass/shard1",
		"institution1.edu/glass/shard2",
		"institution1.edu/glass/shard3",
	}

	for _, client := range allClients {
		html := client.GET("/deletions/show/1").Expect().
			Status(http.StatusOK).Body().Raw()
		MatchesAll(t, html, items)
	}

}

func TestDeletionRequestIndex(t *testing.T) {
	initHTTPTests(t)

	// All users should see these filters on the index page.
	commonFilters := []string{
		`select name="stage"`,
		`select name="status"`,
		`name="requested_at__gteq"`,
		`name="requested_at__lteq"`,
	}

	// Only sys admin should see these filters.
	sysAdminFilters := []string{
		`select name="institution_id"`,
	}

	deletionLinks := []string{
		"/deletions/show/1",
		"/deletions/show/2",
		"/deletions/show/3",
	}

	for _, client := range allClients {
		html := client.GET("/deletions").Expect().
			Status(http.StatusOK).Body().Raw()
		MatchesAll(t, html, deletionLinks)
		MatchesAll(t, html, commonFilters)
		if client == sysAdminClient {
			MatchesAll(t, html, sysAdminFilters)
		} else {
			MatchesNone(t, html, sysAdminFilters)
		}
	}
}

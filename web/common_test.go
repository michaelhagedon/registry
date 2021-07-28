package web_test

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/APTrust/registry/app"
	"github.com/APTrust/registry/db"
	"github.com/APTrust/registry/forms"
	"github.com/APTrust/registry/pgmodels"
	"github.com/gavv/httpexpect/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var appEngine *gin.Engine
var baseURL = "http://localhost"
var fixturesReloaded = false
var sysAdminClient *httpexpect.Expect
var instAdminClient *httpexpect.Expect
var instUserClient *httpexpect.Expect
var allClients []*httpexpect.Expect

var sysAdmin *pgmodels.User
var inst1Admin *pgmodels.User
var inst1User *pgmodels.User
var inst2Admin *pgmodels.User

var sysAdminToken string
var instAdminToken string
var instUserToken string

var allInstNames []string
var allUserNames []string

var userFor map[*httpexpect.Expect]*pgmodels.User
var tokenFor map[*httpexpect.Expect]string

func initHTTPTests(t *testing.T) {
	// Force fixture reload to get rid of any records
	// that the pgmodels tests may have inserted or changed.
	// This gives us a known set of fixtures to work with.
	if fixturesReloaded == false {
		err := db.ForceFixtureReload()
		require.Nil(t, err)
		fixturesReloaded = true
	}
	if appEngine == nil {
		appEngine = app.InitAppEngine(true)
		sysAdminClient, sysAdminToken = initClient(t, "system@aptrust.org")
		instAdminClient, instAdminToken = initClient(t, "admin@inst1.edu")
		instUserClient, instUserToken = initClient(t, "user@inst1.edu")
		allClients = []*httpexpect.Expect{
			sysAdminClient,
			instAdminClient,
			instUserClient,
		}

		sysAdmin = initUser(t, "system@aptrust.org")
		inst1Admin = initUser(t, "admin@inst1.edu")
		inst1User = initUser(t, "user@inst1.edu")
		inst2Admin = initUser(t, "admin@inst2.edu")

		userFor = make(map[*httpexpect.Expect]*pgmodels.User)
		userFor[sysAdminClient] = sysAdmin
		userFor[instAdminClient] = inst1Admin
		userFor[instUserClient] = inst1User

		tokenFor = make(map[*httpexpect.Expect]string)
		tokenFor[sysAdminClient] = sysAdminToken
		tokenFor[instAdminClient] = instAdminToken
		tokenFor[instUserClient] = instUserToken
	}
}

func initClient(t *testing.T, email string) (*httpexpect.Expect, string) {
	client := getAnonymousClient(t)

	// In fixture data, password for all users is 'password'
	signInForm := map[string]string{
		"email":    email,
		"password": "password",
	}

	// Sign the user in, and be sure we got on OK.
	// The client cookie jar will store the session
	// cookie for this user.
	html := client.POST("/users/sign_in").
		WithForm(signInForm).Expect().Status(http.StatusOK).Body().Raw()
	csrfToken := extractCSRFToken(t, html)

	return client, csrfToken
}

func getAnonymousClient(t *testing.T) *httpexpect.Expect {
	client := httpexpect.WithConfig(httpexpect.Config{
		BaseURL: baseURL,
		Client: &http.Client{
			Transport: httpexpect.NewBinder(appEngine),
			Jar:       httpexpect.NewJar(),
			Timeout:   time.Second * 3,
		},

		// NewAssertReporter or NewRequireReporter
		Reporter: httpexpect.NewAssertReporter(t),
		//Printers: []httpexpect.Printer{
		//	httpexpect.NewDebugPrinter(t, true),
		//},
	})
	return client
}

func initUser(t *testing.T, email string) *pgmodels.User {
	user, err := pgmodels.UserByEmail(email)
	require.Nil(t, err)
	require.NotNil(t, user)
	return user
}

// Extract csrf token from http response, so we can include it
// it POST and PUT requests.
func extractCSRFToken(t *testing.T, html string) string {
	re := regexp.MustCompile(`<meta name="csrf_token" content="(.+)">`)
	m := re.FindAllStringSubmatch(html, 1)
	require.True(t, len(m) > 0)
	require.True(t, len(m[0]) > 0)
	token := m[0][1]
	require.NotEmpty(t, token)
	return token
}

// OptionLabels returns the text labels from a list of HTML options.
func OptionLabels(options []forms.ListOption) []string {
	labels := make([]string, len(options))
	for i, opt := range options {
		labels[i] = opt.Text
	}
	return labels
}

// AllInstitutionNames returns a list of all institution names
// in our test data, in no guaranteed order. We use this to ensure
// that pages containing institution lists do indeed display all
// institutions.
func AllInstitutionNames(t *testing.T) []string {
	if len(allInstNames) == 0 {
		options, err := forms.ListInstitutions(false)
		require.Nil(t, err)
		allInstNames = OptionLabels(options)
	}
	return allInstNames
}

// AllUserNames returns a list of all user names in our test data,
// in no guaranteed order. We use this to ensure that pages containing
// user lists do indeed display all users.
func AllUserNames(t *testing.T) []string {
	if len(allUserNames) == 0 {
		options, err := forms.ListUsers(0)
		require.Nil(t, err)
		allUserNames = OptionLabels(options)
	}
	return allUserNames
}

// InstUserNames returns the names of all users at an institution.
func InstUserNames(t *testing.T, institutionID int64) []string {
	options, err := forms.ListUsers(0)
	require.Nil(t, err)
	return OptionLabels(options)
}

// Note on match functions:
// httpexpect.String includes good matching functions, but they
// don't behave well in loops. We get panics instead of proper
// test failure reports.

// AssertMatchesAll asserts that all strings in items appear in body.
func AssertMatchesAll(t *testing.T, body string, items []string) {
	for _, item := range items {
		assert.True(t, strings.Contains(body, item), "Missing expected string: %s", item)
	}
}

// AssertMatchesNone asserts that no strings in items appear in body.
func AssertMatchesNone(t *testing.T, body string, items []string) {
	for _, item := range items {
		assert.False(t, strings.Contains(body, item), "Found unexpected string: %s", item)
	}
}

// MatchResult count asserts that the result count at the bottom of
// a list/index page matches the expected count. In the HTML pager,
// the result count appears in the format "1 - 20 of 215".
func AssertMatchesResultCount(t *testing.T, body string, count int) {
	countStr := fmt.Sprintf("%d", count)
	re := regexp.MustCompile(`\d+ - \d+ of (\d+)`)
	matches := re.FindAllStringSubmatch(body, 1)
	assert.NotNil(t, matches, "Did not find result count string '1 - N of N'")
	if matches != nil {
		assert.Equal(t, countStr, matches[0][1], "Expected result count %d; got %s. Full string: %s", count, matches[0][1], matches[0][0])
	}
}

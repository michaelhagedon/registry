package pgmodels

import (
	"fmt"
	"time"

	"github.com/APTrust/registry/common"
)

// Note: chart_metric and report_type are ignored by backend.
// Used only in front-end.
var DepositStatsFilters = []string{
	"chart_metric",
	"end_date",
	"institution_id",
	"report_type",
	"start_date",
	"storage_option",
}

// DepositStats contains info about member deposits and the costs
// of those deposits. This struct does not implement the usual pgmodel
// interface, nor does it map to a single underlying table or view.
// This struct merely represents to the output of a reporting query.
//
// Note that we store CostGBPerMonth and MonthlyCost in the table
// rather than calculating them because cost per GB per month may
// change over time, and we want to capture the actual historical
// cost for each month.
type DepositStats struct {
	InstitutionID       int64     `json:"institution_id"`
	MemberInstitutionID int64     `json:"member_institution_id"`
	InstitutionName     string    `json:"institution_name"`
	StorageOption       string    `json:"storage_option"`
	ObjectCount         int64     `json:"object_count"`
	FileCount           int64     `json:"file_count"`
	TotalBytes          int64     `json:"total_bytes"`
	TotalGB             float64   `json:"total_gb" pg:"total_gb"`
	TotalTB             float64   `json:"total_tb" pg:"total_tb"`
	CostGBPerMonth      float64   `json:"cost_gb_per_month" pg:"cost_gb_per_month"`
	MonthlyCost         float64   `json:"monthly_cost"`
	EndDate             time.Time `json:"end_date"`
	PrimarySort         string    `json:"-"`
	SecondarySort       string    `json:"-"`
}

// ----------------------------------------------------------------------------
//
// Why is this so complicated? Because deposit stats, being extremely
// expensive to calculate, are stored in a one table (historical_deposit_stats)
// and one materialized view (current_deposit_stats), both of which are
// populated periodically in the background through a fairly complex
// query that already includes a number of sums and rollups.
//
// If we roll up sub account subtotals into these tables/materialized views,
// it becomes difficult to query the data without returning extranous rows.
//
// So the not-too-pretty solution here is run a second query to
// calculate sub-account totals. C'est la guerre. Ne pleurez pas.
//
// ----------------------------------------------------------------------------

// DepositStatsSelect returns info about materials a depositor updated
// in our system before a given date. This breaks down deposits by
// storage option and institution. To report on all institutions, use
// zero for institutionID. To report on all storage options, pass an
// empty string for storageOption.
func DepositStatsSelect(institutionID int64, storageOption string, endDate time.Time) ([]*DepositStats, error) {
	var stats []*DepositStats
	statsQuery := getDepositStatsQuery(institutionID, storageOption, endDate)
	//fmt.Println(statsQuery, "INST", institutionID, "STOR", storageOption, "END", endDate)
	//fmt.Println(statsQuery)
	_, err := common.Context().DB.Query(&stats, statsQuery,
		institutionID, institutionID, institutionID,
		storageOption, storageOption,
		endDate, endDate)

	// If we happen to get a query for a date before 2014,
	// we'll get no results. We don't want to return nil, because
	// the caller is likely expecting something that can be serialized
	// to JSON. Give the caller an actual answer, saying there was
	// nothing in the system on the date they inquired about.
	if stats == nil {
		stats = make([]*DepositStats, 1)
		stats[0] = &DepositStats{
			InstitutionName: "Total",
			StorageOption:   "Total",
			EndDate:         endDate,
		}
	}
	if err != nil {
		return stats, err
	}

	// We have to get these separately
	subAcctStats, err := calculateSubAccountRollup(institutionID, storageOption, endDate)
	if err == nil && subAcctStats != nil {
		stats = append(stats, subAcctStats...)
	}

	calculateMontlyTotals(stats)

	return stats, err
}

func DepositStatsOverTime(institutionID int64, storageOption string, startDate, endDate time.Time) ([]*DepositStats, error) {
	var stats []*DepositStats
	statsQuery := getDepositTimelineQuery(institutionID)
	var err error
	if institutionID == 0 {
		// Omit inst id if zero because it will go into the startDate param slot and cause an error.
		_, err = common.Context().DB.Query(&stats, statsQuery, startDate, endDate)
	} else {
		_, err = common.Context().DB.Query(&stats, statsQuery, institutionID, startDate, endDate)
	}
	return stats, err
}

func getDepositTimelineQuery(institutionID int64) string {
	// Basic depost stats query. Use the "is null / or" trick to deal with
	// filters that may or may not be present. Also note that historical
	// deposit stats uses EXACT FIRST-OF-MONTH dates, so we look for
	// "end_date = " not "<" or "<=".
	q := `select 
			institution_id, 
			member_institution_id,
			institution_name, 
			storage_option, 
			file_count, 
			object_count, 
			total_bytes, 
			total_gb, 
			total_tb, 
			monthly_cost, 
			end_date,
			primary_sort,
			secondary_sort 
			from historical_deposit_stats  
		where institution_id %s
		and end_date >= ?
		and end_date <= ?
		order by primary_sort, secondary_sort, end_date`
	op := " = ? "
	if institutionID == 0 {
		op = " is null "
	}
	return fmt.Sprintf(q, op)
}

func getDepositStatsQuery(institutionID int64, storageOption string, endDate time.Time) string {
	// Basic depost stats query. Use the "is null / or" trick to deal with
	// filters that may or may not be present. Also note that historical
	// deposit stats uses EXACT FIRST-OF-MONTH dates, so we look for
	// "end_date = " not "<" or "<=".
	q := `select institution_id, 
				institution_name, 
				storage_option, 
				file_count, 
				object_count, 
				total_bytes, 
				total_gb, 
				total_tb, 
				cost_gb_per_month,
				monthly_cost, 
				end_date 
				from %s 
				where (? = 0 or institution_id = ? or member_institution_id = ?)
				and (? = '' or storage_option = ?) `

	tableName, dateClause := getTableNameAndDateClause(endDate)
	q += dateClause
	q += " order by end_date, primary_sort, secondary_sort"
	return fmt.Sprintf(q, tableName)
}

func calculateSubAccountRollup(institutionID int64, storageOption string, endDate time.Time) ([]*DepositStats, error) {
	common.Context().Log.Info().Msgf("calculateSubAccountRollup for inst %d", institutionID)
	if institutionID < 1 {
		common.Context().Log.Info().Msgf("calculateSubAccountRollup: ignoring request for inst %d", institutionID)
		return nil, nil
	}
	inst, err := InstitutionByID(institutionID)
	if err != nil {
		common.Context().Log.Info().Msgf("calculateSubAccountRollup: error looking up inst %d: %v", institutionID, err)
		return nil, err
	}
	hasSubAccounts, err := inst.HasSubAccounts()
	if err != nil {
		common.Context().Log.Error().Msgf("calculateSubAccountRollup: inst %d: error checking for subaccounts: %v", institutionID, err)
		return nil, err
	}
	if !hasSubAccounts {
		common.Context().Log.Info().Msgf("calculateSubAccountRollup: ignoring inst %d: institution has no subaccounts", institutionID)
		return nil, nil
	}
	var stats []*DepositStats
	rollupQuery := getSubAccountRollupQuery(institutionID, storageOption, endDate)
	common.Context().Log.Info().Msgf("calculateSubAccountRollup: gathering rollup data with the following query: %s", rollupQuery)
	_, err = common.Context().DB.Query(&stats, rollupQuery,
		institutionID, institutionID,
		storageOption, storageOption,
		endDate, endDate)

	// Empty result set is OK, because sub accounts often
	// exist for weeks or months before making a deposit.
	// So don't consider empty result set an error.
	if IsNoRowError(err) {
		common.Context().Log.Info().Msg("calculateSubAccountRollup: query returned no rows")
		return stats, nil
	}
	if err != nil {
		common.Context().Log.Error().Msgf("calculateSubAccountRollup: query returned error: %v", err)
	}
	common.Context().Log.Info().Msgf("calculateSubAccountRollup: got stats: %v", stats)
	return stats, err
}

func getSubAccountRollupQuery(institutionID int64, storageOption string, endDate time.Time) string {
	q := `select 0 as institution_id, 
	'All Institutions' as institution_name, 
	storage_option, 
	sum(file_count) as file_count, 
	sum(object_count) as object_count, 
	sum(total_bytes) as total_bytes, 
	sum(total_gb) as total_gb, 
	sum(total_tb) as total_tb, 
	sum(cost_gb_per_month) as cost_gb_per_month,
	sum(monthly_cost) as monthly_cost, 
	end_date 
	from %s 
	where (institution_id = ? or member_institution_id = ?)
	and (? = '' or storage_option = ?) `

	tableName, dateClause := getTableNameAndDateClause(endDate)
	q += dateClause
	q += " group by storage_option, end_date, secondary_sort order by secondary_sort "
	return fmt.Sprintf(q, tableName)
}

func getTableNameAndDateClause(endDate time.Time) (string, string) {
	tableName := "historical_deposit_stats"
	dateClause := ""

	// Current stats report, which displays on dashboard, passes in
	// time.Now() as end date. In this case, we want to query the
	// current stats table, not historical stats.
	now := time.Now().UTC()
	firstOfThisMonth := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if endDate.After(firstOfThisMonth) || endDate == firstOfThisMonth {
		// current stats view does not need end_date
		tableName = "current_deposit_stats"
	} else {
		// historical stats has exact cache dates
		dateClause = " and (? = '0001-01-01 00:00:00+00:00:00' or end_date = ?) "
	}
	return tableName, dateClause
}

func calculateMontlyTotals(stats []*DepositStats) {
	instTotal := 0.0
	for _, stat := range stats {
		instTotal += stat.MonthlyCost
		if stat.StorageOption == "Total" {
			stat.MonthlyCost = instTotal
			instTotal = 0.0
		}
	}
}

package pgmodels

import (
	"fmt"
	"time"

	"github.com/APTrust/registry/common"
)

// Note: chart_metric is ignored by backend. Used only in front-end.
var DepositStatsFilters = []string{
	"chart_metric",
	"end_date",
	"institution_id",
	"storage_option",
}

// DepositStats contains info about member deposits and the costs
// of those deposits. This struct does not implement the usual pgmodel
// interface, nor does it map to a single underlying table or view.
// This struct merely represents to the output of a reporting query.
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
}

// DepositStatsSelect returns info about materials a depositor updated
// in our system before a given date. This breaks down deposits by
// storage option and institution. To report on all institutions, use
// zero for institutionID. To report on all storage options, pass an
// empty string for storageOption.
func DepositStatsSelect(institutionID int64, storageOption string, endDate time.Time) ([]*DepositStats, error) {
	var stats []*DepositStats
	statsQuery := getDepositStatsQuery(institutionID, storageOption, endDate)
	// fmt.Println(statsQuery, "INST", institutionID, "STOR", storageOption, "END", endDate)
	// fmt.Println(statsQuery)
	_, err := common.Context().DB.Query(&stats, statsQuery,
		institutionID, institutionID, institutionID,
		storageOption, storageOption,
		endDate, endDate)

	// If we happen to get a query for a date before 2014,
	// we'll get no results. We don't want to return nil, because
	// the caller is likely expected something that can be serialized
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

	return stats, err
}

// TODO: Change the query below to use a union, so primary depositor
//       is on top. We'll also need to add a rollup. This may best be
//       done in a stored procedure or function, or in the stats tables
//       themselves.

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
				end_date from %s 
				where (? = 0 or institution_id = ? or member_institution_id = ?)
				and (? = '' or storage_option = ?) `
	tableName := "historical_deposit_stats"

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
		q += "and (? = '0001-01-01 00:00:00+00:00:00' or end_date = ?)"
	}
	q += " order by primary_sort, secondary_sort"
	return fmt.Sprintf(q, tableName)
}

func calculateSubAccountRollup(institutionID int64, storageOption string, endDate time.Time) ([]*DepositStats, error) {
	inst, err := InstitutionByID(institutionID)
	if err != nil {
		return nil, err
	}
	hasSubAccounts, err := inst.HasSubAccounts()
	if err != nil {
		return nil, err
	}
	if !hasSubAccounts {
		return nil, nil
	}
	var stats []*DepositStats
	rollupQuery := getSubAccountRollupQuery(institutionID, storageOption, endDate)
	_, err = common.Context().DB.Query(&stats, rollupQuery,
		institutionID, institutionID,
		storageOption, storageOption,
		endDate, endDate)
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

	tableName := "historical_deposit_stats"

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
		q += "and (? = '0001-01-01 00:00:00+00:00:00' or end_date = ?) "
	}
	q += " group by storage_option, end_date, secondary_sort order by secondary_sort "
	return fmt.Sprintf(q, tableName)
}

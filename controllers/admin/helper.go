package admin

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// Converts query params to SQL-style SP params for AdminUsers list
func BuildAdminUserListSPParams(query url.Values, siteUsersId int) map[string]interface{} {
	// Parse pagination
	pageNumber := 1
	pageSize := 10
	if v := query.Get("pageNumber"); v != "" {
		fmt.Sscanf(v, "%d", &pageNumber)
	}
	if v := query.Get("pageSize"); v != "" {
		fmt.Sscanf(v, "%d", &pageSize)
	}

	// Parse sort
	sortBy := query.Get("sortBy")
	rawSortString := query.Get("rawSortString")
	if sortBy == "" && rawSortString != "" {
		sortBy = rawSortString
	}

	// Parse filters
	filters := query.Get("filters")
	rawFilterString := query.Get("rawFilterString")
	sqlFilters := ""
	if filters != "" {
		sqlFilters = ConvertFilterStringToSQL(filters)
	}
	if rawFilterString == "" && sqlFilters != "" {
		rawFilterString = sqlFilters
	}

	// Parse search
	search := query.Get("search")
	rawSearchString := query.Get("rawSearchString")
	sqlSearch := ""
	if search != "" {
		sqlSearch = ConvertSearchStringToSQL(search)
	}
	if rawSearchString == "" && sqlSearch != "" {
		rawSearchString = sqlSearch
	}

	// Build param map
	spParams := map[string]interface{}{
		"User_SiteUsersID": siteUsersId,
		"PageNumber":       pageNumber,
		"PageSize":         pageSize,
		"ListKey":          "AdminUsers",
		"TrackingID":       "Dashboard-AdminList",
	}
	// Only add non-empty string params
	if sortBy != "" {
		spParams["SortBy"] = sortBy
	}
	if rawSortString != "" {
		spParams["RawSortString"] = rawSortString
	}
	if sqlFilters != "" {
		spParams["Filters"] = sqlFilters
	}
	if rawFilterString != "" {
		spParams["RawFilterString"] = rawFilterString
	}
	if sqlSearch != "" {
		spParams["SearchString"] = sqlSearch
	}
	if rawSearchString != "" {
		spParams["RawSearchString"] = rawSearchString
	}
	return spParams
}

// Converts filter string like 'Field CONTAINS (val) | Field2 EQ (val2)' to SQL WHERE clause
func ConvertFilterStringToSQL(filter string) string {
	if filter == "" {
		return ""
	}
	parts := strings.Split(filter, "|")
	var sqlParts []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		// CONTAINS
		if strings.Contains(part, "CONTAINS") {
			re := regexp.MustCompile(`([\w.]+) CONTAINS \(([^)]+)\)`)
			sqlParts = append(sqlParts, re.ReplaceAllStringFunc(part, func(m string) string {
				sub := re.FindStringSubmatch(m)
				if len(sub) == 3 {
					field := strings.ReplaceAll(sub[1], "__", ".")
					val := strings.ReplaceAll(sub[2], "'", "''")
					return field + " LIKE '%" + val + "%'"
				}
				return m
			}))
		}
		if strings.Contains(part, "EQ") {
			re := regexp.MustCompile(`([\w.]+) EQ \(([^)]+)\)`)
			sqlParts = append(sqlParts, re.ReplaceAllStringFunc(part, func(m string) string {
				sub := re.FindStringSubmatch(m)
				if len(sub) == 3 {
					field := strings.ReplaceAll(sub[1], "__", ".")
					val := strings.ReplaceAll(sub[2], "'", "''")
					return field + " = '" + val + "'"
				}
				return m
			}))
		}
		if strings.Contains(part, "BETWEEN") {
			re := regexp.MustCompile(`([\w.]+) BETWEEN \(([^)]+)TO([^)]+)\)`)
			sqlParts = append(sqlParts, re.ReplaceAllStringFunc(part, func(m string) string {
				sub := re.FindStringSubmatch(m)
				if len(sub) == 4 {
					field := strings.ReplaceAll(sub[1], "__", ".")
					from := strings.TrimSpace(strings.ReplaceAll(sub[2], "'", "''"))
					to := strings.TrimSpace(strings.ReplaceAll(sub[3], "'", "''"))
					if from != "" && to != "" {
						return field + " BETWEEN '" + from + "' AND '" + to + "'"
					} else if from != "" {
						return field + " >= '" + from + "'"
					} else if to != "" {
						return field + " <= '" + to + "'"
					}
				}
				return m
			}))
		}
	}
	if len(sqlParts) > 0 {
		return " AND " + strings.Join(sqlParts, " AND ")
	}
	return ""
}

// Converts search string to SQL WHERE clause
func ConvertSearchStringToSQL(search string) string {
	if search == "" {
		return ""
	}
	// For demo: just wrap in WHERE and LIKE for all fields (customize as needed)
	fields := []string{"AdminUsers.FirstName", "AdminUsers.LastName", "SiteUsers.EmailAddress", "AdminUsers.AdminUsersCode"}
	var orParts []string
	for _, f := range fields {
		orParts = append(orParts, f+" LIKE '%"+search+"%'")
	}
	return "WHERE (" + strings.Join(orParts, " OR ") + ")"
}

func SendAuthError(c *gin.Context, errorType int) {
	errorMsg := []gin.H{}
	if errorType == 0 {

		errorMsg = []gin.H{
			{"fieldName": "Username", "messageCode": "Username_Or_Password_Incorrect"},
			{"fieldName": "Password", "messageCode": "Username_Or_Password_Incorrect"},
		}
	} else if errorType == 1 {
		errorMsg = []gin.H{
			{"fieldName": "TfaCode", "messageCode": "Invalid"},
		}
	} else if errorType == 2 {
		errorMsg = []gin.H{
			{"fieldName": "TfaType", "messageCode": "TfaType_Invalid"},
		}

	}
	c.JSON(http.StatusUnauthorized, gin.H{
		"id":      0,
		"details": nil,
		"status":  "0",
		"errors":  errorMsg,
	})
}

// getChartSPName maps chart endpoints to their respective stored procedure names
func getChartSPName(path string) string {
	switch {
	case strings.Contains(path, "/depositschart"):
		return "v1_AdminRole_DashboardModule_GetDeposits"
	case strings.Contains(path, "/withdrawalschart"):
		return "v1_AdminRole_DashboardModule_GetWithdrawals"
	case strings.Contains(path, "/cardsissuedchart"):
		return "v1_AdminRole_DashboardModule_GetNumberCardsIssued"
	case strings.Contains(path, "/cardspendchart"):
		return "v1_AdminRole_DashboardModule_GetCardSpend"
	case strings.Contains(path, "/turnoverchart"):
		return "v1_AdminRole_DashboardModule_GetTurnover"
	case strings.Contains(path, "/revenuechart"):
		return "v1_AdminRole_DashboardModule_GetRevenue"
	case strings.Contains(path, "/usercounts"):
		return "v1_AdminRole_DashboardModule_GetUserCounts"
	default:
		return ""
	}
}

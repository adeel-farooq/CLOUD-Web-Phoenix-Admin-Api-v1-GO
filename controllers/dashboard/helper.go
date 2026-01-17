// Helper to get columns for adminrolesmodule

package dashboard

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

func BuildSPParams(query url.Values, siteUsersId int, path string) map[string]interface{} {
	pageNumber := 1
	pageSize := 10
	if v := query.Get("pageNumber"); v != "" {
		fmt.Sscanf(v, "%d", &pageNumber)
	}
	if v := query.Get("pageSize"); v != "" {
		fmt.Sscanf(v, "%d", &pageSize)
	}

	sortBy := query.Get("sortBy")
	filters := query.Get("filters")
	search := query.Get("search")

	sqlSort := ConvertSortStringToSQL(sortBy)
	sqlFilters := ConvertFilterStringToSQL(filters)
	sqlSearch := ConvertSearchStringToSQL(search, path)

	spParams := map[string]interface{}{
		"User_SiteUsersID": siteUsersId,
		"PageNumber":       pageNumber,
		"PageSize":         pageSize,
		"ListKey":          "AdminUsers",
		"TrackingID":       "DefaultTrackingID",
		"RawSortString":    sortBy,
		"RawFilterString":  filters,
		"RawSearchString":  search,
	}

	if sqlSort != "" {
		spParams["SortBy"] = sqlSort
	}
	if sqlFilters != "" {
		spParams["Filters"] = sqlFilters
	}
	if sqlSearch != "" {
		spParams["SearchString"] = sqlSearch
	}

	return spParams
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
func getSPName(path string) string {
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
	// case strings.Contains(path, "/usercounts"):
	// 	return "v1_AdminRole_DashboardModule_GetUserCounts"
	// case strings.Contains(path, "/licenseeadminusersmodule/list"):
	// 	return "v1_AdminRole_LicenseeAdminUsersModule_List"
	// case strings.Contains(path, "/adminusersmodule/list"):
	// 	return "v1_AdminRole_AdminUsersModule_List"
	// case strings.Contains(path, "/adminrolesmodule/list"):
	// 	return "v1_AdminRole_AdminRolesModule_List"
	default:
		return ""
	}
}

// Helper to get listData mapping based on module type (snake_case)
func GetAdminUsersListData(path string, row map[string]interface{}) map[string]interface{} {
	if strings.Contains(path, "licenseeadminusersmodule") {
		return map[string]interface{}{
			"AdminUsers__AddDate":           row["AdminUsers__AddDate"],
			"AdminUsers__AdminUsersCode":    row["AdminUsers__AdminUsersCode"],
			"SiteUsers__EmailAddress":       row["SiteUsers__EmailAddress"],
			"Licensees__LicenseeName":       row["Licensees__LicenseeName"],
			"LicenseesBrands__InternalName": row["LicenseesBrands__InternalName"],
			"AdminUsers__FirstName":         row["AdminUsers__FirstName"],
			"AdminUsers__LastName":          row["AdminUsers__LastName"],
			"SiteUsers__bSuppressed":        row["SiteUsers__bSuppressed"],
			"AdminUsers__Id":                row["AdminUsers__Id"],
		}
	}
	// Default: adminusersmodule
	return map[string]interface{}{
		"AdminUsers__AddDate":        row["AdminUsers__AddDate"],
		"AdminUsers__AdminUsersCode": row["AdminUsers__AdminUsersCode"],
		"AdminUsers__FirstName":      row["AdminUsers__FirstName"],
		"AdminUsers__Id":             row["AdminUsers__Id"],
		"AdminUsers__LastName":       row["AdminUsers__LastName"],
		"SiteUsers__bSuppressed":     row["SiteUsers__bSuppressed"],
		"SiteUsers__EmailAddress":    row["SiteUsers__EmailAddress"],
	}
}

func ConvertSortStringToSQL(sort string) string {
	if sort == "" {
		return ""
	}
	parts := strings.Split(sort, "|")
	out := []string{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		// format: adminUsers__FirstName DESC
		chunks := strings.Fields(p)
		if len(chunks) != 2 {
			continue
		}
		col := strings.ReplaceAll(chunks[0], "__", ".")
		dir := strings.ToUpper(chunks[1])
		if dir != "ASC" && dir != "DESC" {
			continue
		}
		out = append(out, col+" "+dir)
	}
	if len(out) == 0 {
		return ""
	}
	return " ORDER BY " + strings.Join(out, ", ")
}

func escapeSQL(val string) string {
	return strings.ReplaceAll(val, "'", "''")
}

func ConvertFilterStringToSQL(filter string) string {
	if strings.TrimSpace(filter) == "" {
		return ""
	}

	parts := strings.Split(filter, "|")
	sqlParts := []string{}

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// columnKey OP (value)
		re := regexp.MustCompile(`^(.+?)\s+([A-Z]+)\s+\((.+)\)$`)
		m := re.FindStringSubmatch(part)
		if len(m) != 4 {
			continue
		}

		colKey := strings.TrimSpace(m[1])
		op := strings.TrimSpace(m[2])
		rawVal := strings.TrimSpace(m[3])

		col := strings.ReplaceAll(colKey, "__", ".")
		val := escapeSQL(rawVal)

		switch op {
		case "EQ":
			sqlParts = append(sqlParts, col+" = '"+val+"'")
		case "NEQ":
			sqlParts = append(sqlParts, col+" <> '"+val+"'")
		case "CONTAINS":
			sqlParts = append(sqlParts, col+" LIKE '%"+val+"%'")
		case "SW": // starts with
			sqlParts = append(sqlParts, col+" LIKE '"+val+"%'")
		case "EW": // ends with
			sqlParts = append(sqlParts, col+" LIKE '%"+val+"'")
		case "GT":
			sqlParts = append(sqlParts, col+" > '"+val+"'")
		case "GEQ":
			sqlParts = append(sqlParts, col+" >= '"+val+"'")
		case "LT":
			sqlParts = append(sqlParts, col+" < '"+val+"'")
		case "LEQ":
			sqlParts = append(sqlParts, col+" <= '"+val+"'")
		case "BETWEEN":
			// format: (fromTOto)
			chunks := strings.Split(rawVal, "TO")
			if len(chunks) == 2 {
				from := escapeSQL(strings.TrimSpace(chunks[0]))
				to := escapeSQL(strings.TrimSpace(chunks[1]))
				if from != "" && to != "" {
					sqlParts = append(sqlParts, col+" BETWEEN '"+from+"' AND '"+to+"'")
				} else if from != "" {
					sqlParts = append(sqlParts, col+" >= '"+from+"'")
				} else if to != "" {
					sqlParts = append(sqlParts, col+" <= '"+to+"'")
				}
			}
		case "INSTRINGARRAY":
			// value example: A,B,C
			arr := strings.Split(rawVal, ",")
			vals := []string{}
			for _, a := range arr {
				a = strings.TrimSpace(a)
				if a == "" {
					continue
				}
				vals = append(vals, "'"+escapeSQL(a)+"'")
			}
			if len(vals) > 0 {
				sqlParts = append(sqlParts, col+" IN ("+strings.Join(vals, ",")+")")
			}
		}
	}

	if len(sqlParts) == 0 {
		return ""
	}
	return " AND " + strings.Join(sqlParts, " AND ")
}
func ConvertSearchStringToSQL(search string, path string) string {
	search = strings.TrimSpace(search)
	if search == "" {
		return ""
	}
	s := escapeSQL(search)

	fields := []string{
		"AdminUsers.FirstName",
		"AdminUsers.LastName",
		"SiteUsers.EmailAddress",
		"AdminUsers.AdminUsersCode",
	}

	var orParts []string
	for _, f := range fields {
		orParts = append(orParts, f+" LIKE '%"+s+"%'")
	}
	return "WHERE (" + strings.Join(orParts, " OR ") + ")"
}
func DeserializeFilters(serialized string) []map[string]interface{} {
	serialized = strings.TrimSpace(serialized)
	if serialized == "" {
		return []map[string]interface{}{}
	}

	parts := strings.Split(serialized, "|")
	type filt struct {
		ColumnKey string
		Operator  string
		Value     interface{}
	}

	all := []filt{}
	re := regexp.MustCompile(`^(.+?)\s+([A-Z]+)\s+\((.+)\)$`)

	for _, p := range parts {
		p = strings.TrimSpace(p)
		m := re.FindStringSubmatch(p)
		if len(m) != 4 {
			continue
		}
		col := strings.TrimSpace(m[1])
		op := strings.TrimSpace(m[2])
		val := strings.TrimSpace(m[3])

		var vv interface{} = val
		if op == "BETWEEN" {
			vv = strings.Split(val, "TO")
		}
		all = append(all, filt{ColumnKey: col, Operator: op, Value: vv})
	}

	// group by columnKey
	group := map[string][]map[string]interface{}{}
	for _, f := range all {
		group[f.ColumnKey] = append(group[f.ColumnKey], map[string]interface{}{
			"columnKey": f.ColumnKey,
			"operator":  f.Operator,
			"value":     f.Value,
		})
	}

	out := []map[string]interface{}{}
	for col, arr := range group {
		out = append(out, map[string]interface{}{
			"columnKey": col,
			"filters":   arr,
		})
	}

	return out
}

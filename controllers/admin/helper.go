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
	fields := []string{
		"AdminUsers.FirstName",
		"AdminUsers.LastName",
		"SiteUsers.EmailAddress",
		"AdminUsers.AdminUsersCode",
		"Licensees.LicenseeName",
		"LicenseesBrands.InternalName",
	}
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

// Helper to get columns based on module type (snake_case)
func GetAdminUsersColumns(path string) []map[string]interface{} {
	if strings.Contains(path, "licenseeadminusersmodule") {
		return []map[string]interface{}{
			{"columnKey": "AdminUsers__Id", "labelKey": "Id", "labelValue": "Id", "orderNumber": 1, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "Integer", "filterMetadata": map[string]interface{}{"filterType": "Amount", "details": nil}},
			{"columnKey": "AdminUsers__AdminUsersCode", "labelKey": "AdminUsersCode", "labelValue": "Code", "orderNumber": 2, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "String", "filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil}},
			{"columnKey": "SiteUsers__EmailAddress", "labelKey": "EmailAddress", "labelValue": "Email Address", "orderNumber": 3, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "String", "filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil}},
			{"columnKey": "Licensees__LicenseeName", "labelKey": "LicenseeName", "labelValue": "Licensee Name", "orderNumber": 4, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "String", "filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil}},
			{"columnKey": "LicenseesBrands__InternalName", "labelKey": "InternalName", "labelValue": "Brand Name", "orderNumber": 5, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "String", "filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil}},
			{"columnKey": "AdminUsers__FirstName", "labelKey": "FirstName", "labelValue": "First Name", "orderNumber": 6, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "String", "filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil}},
			{"columnKey": "AdminUsers__LastName", "labelKey": "LastName", "labelValue": "Last Name", "orderNumber": 7, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "String", "filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil}},
			{"columnKey": "SiteUsers__bSuppressed", "labelKey": "bSuppressed", "labelValue": "Suppressed", "orderNumber": 8, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "Boolean", "filterMetadata": map[string]interface{}{"filterType": "SingleChoice", "details": map[string]interface{}{"PossibleValues": []map[string]string{{"value": "0", "label": "Active"}, {"value": "1", "label": "Inactive"}}}}},
			{"columnKey": "AdminUsers__AddDate", "labelKey": "AddDate", "labelValue": "Add Date", "orderNumber": 9, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "DateTime", "filterMetadata": map[string]interface{}{"filterType": "DateTime:Range", "details": map[string]interface{}{}}},
		}
	}
	// Default: adminusersmodule
	return []map[string]interface{}{
		{"columnKey": "AdminUsers__Id", "labelKey": "Id", "labelValue": "Id", "orderNumber": 1, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "Integer", "filterMetadata": map[string]interface{}{"filterType": "Amount", "details": nil}},
		{"columnKey": "AdminUsers__AdminUsersCode", "labelKey": "AdminUsersCode", "labelValue": "Code", "orderNumber": 2, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "String", "filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil}},
		{"columnKey": "SiteUsers__EmailAddress", "labelKey": "EmailAddress", "labelValue": "Email Address", "orderNumber": 3, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "String", "filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil}},
		{"columnKey": "AdminUsers__FirstName", "labelKey": "FirstName", "labelValue": "First Name", "orderNumber": 4, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "String", "filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil}},
		{"columnKey": "AdminUsers__LastName", "labelKey": "LastName", "labelValue": "Last Name", "orderNumber": 5, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "String", "filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil}},
		{"columnKey": "SiteUsers__bSuppressed", "labelKey": "bSuppressed", "labelValue": "Suppressed", "orderNumber": 6, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "Boolean", "filterMetadata": map[string]interface{}{"filterType": "SingleChoice", "details": map[string]interface{}{"PossibleValues": []map[string]string{{"value": "0", "label": "Active"}, {"value": "1", "label": "Inactive"}}}}},
		{"columnKey": "AdminUsers__AddDate", "labelKey": "AddDate", "labelValue": "Add Date", "orderNumber": 7, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "DateTime", "filterMetadata": map[string]interface{}{"filterType": "DateTime:Range", "details": map[string]interface{}{}}},
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

// helper.go
package business

import (
	"math"
	"strconv"
	"strings"

	"cloud-web-phoenix-customer-v1-go/controllers/admin"
)

// NET-like SP params:
// - RawSearchString: ” when empty (tumhari requirement)
// - RawSortString/RawFilterString: NULL when empty (better for SP + dynamic SQL)
// - SortBy/Filters/SearchString: NULL when empty (never "")
func BuildOperationalAssetAccountsSPParamsNetLike(q admin.QueryRecordList, siteUsersId int, cfg admin.ListSPConfig) map[string]interface{} {
	rawSort := strings.TrimSpace(q.SortBy)
	rawFilter := strings.TrimSpace(q.Filters)
	rawSearch := strings.TrimSpace(q.Search)

	// RawSort/RawFilter -> NULL when empty
	var rawSortParam interface{} = nil
	if rawSort != "" {
		rawSortParam = rawSort
	}
	var rawFilterParam interface{} = nil
	if rawFilter != "" {
		rawFilterParam = rawFilter
	}

	// Computed fragments -> NULL when empty
	var sortByParam interface{} = nil
	if rawSort != "" {
		if sqlSort := admin.ConvertSortToSQL(rawSort, cfg.ColumnMap); strings.TrimSpace(sqlSort) != "" {
			sortByParam = sqlSort
		}
	}

	var filtersParam interface{} = nil
	if rawFilter != "" {
		if sqlFilters := admin.ConvertFiltersToSQL(rawFilter, cfg.ColumnMap); strings.TrimSpace(sqlFilters) != "" {
			filtersParam = sqlFilters
		}
	}

	var searchParam interface{} = nil
	if rawSearch != "" {
		// ConvertSearchToSQL should return something like: "WHERE 1=1 AND (...LIKE...)"
		// If your admin.ConvertSearchToSQL returns " AND (...)" then we will wrap it.
		sqlSearch := strings.TrimSpace(admin.ConvertSearchToSQL(rawSearch, cfg.SearchFields))
		if sqlSearch != "" {
			if strings.HasPrefix(strings.ToUpper(sqlSearch), "AND ") {
				sqlSearch = "WHERE 1=1 " + sqlSearch
			}
			searchParam = sqlSearch
		}
	}

	return map[string]interface{}{
		"User_SiteUsersID": siteUsersId,
		"PageNumber":       q.PageNumber,
		"PageSize":         q.PageSize,
		"ListKey":          cfg.ListKey,
		"TrackingID":       cfg.TrackingID,

		"RawSortString":   rawSortParam,   // NULL if empty
		"RawFilterString": rawFilterParam, // NULL if empty
		"RawSearchString": rawSearch,      // '' if empty (as you asked)

		"SortBy":       sortByParam,  // NULL if empty
		"Filters":      filtersParam, // NULL if empty
		"SearchString": searchParam,  // NULL if empty (CRITICAL)
		// optional: some SPs also accept @Search (raw), keep if needed:
		"Search": nilIfEmpty(rawSearch),
	}
}

func nilIfEmpty(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// Row normalize:
// - Remove RowNum/HowManyResults
// - LowercaseFirstChar on key (NET listData keys)
// - Normalize decimal strings -> number
// - Remove trailing 'Z' from datetime if you want NET-like (optional)
func NormalizeRowNetLike(row map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(row))
	for k, v := range row {
		if k == "RowNum" || k == "HowManyResults" {
			continue
		}
		key := admin.LowercaseFirstChar(k)
		out[key] = normalizeValueNetLike(v)
	}
	return out
}

func normalizeValueNetLike(v interface{}) interface{} {
	// string values
	if s, ok := v.(string); ok {
		ss := strings.TrimSpace(s)

		// datetime: if SP returns ...Z and .NET doesn't, strip it (optional)
		if strings.HasSuffix(ss, "Z") && strings.Contains(ss, "T") {
			return strings.TrimSuffix(ss, "Z")
		}

		// decimal numeric string -> number
		if looksNumeric(ss) {
			if f, err := strconv.ParseFloat(ss, 64); err == nil {
				if math.Abs(f-math.Round(f)) < 1e-9 {
					return int64(math.Round(f))
				}
				return f
			}
		}
		return ss
	}
	return v
}

func looksNumeric(s string) bool {
	if s == "" {
		return false
	}
	// allow digits and one dot
	dot := false
	for _, ch := range s {
		if ch == '.' {
			if dot {
				return false
			}
			dot = true
			continue
		}
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

// Columns normalize (NET-like):
// - if bFilterable=false => filterMetadata MUST be nil
// - if bFilterable=true and details is empty map => set details=nil (NET often returns null)
func NormalizeColumnsNetLike(cols []map[string]interface{}) []map[string]interface{} {
	// for _, c := range cols {
	// 	bf, _ := c["bFilterable"].(bool)
	// 	if !bf {
	// 		c["filterMetadata"] = nil
	// 		continue
	// 	}
	// 	if fm, ok := c["filterMetadata"].(map[string]interface{}); ok {
	// 		if d, ok := fm["details"].(map[string]interface{}); ok && len(d) == 0 {
	// 			fm["details"] = nil
	// 		}
	// 	}
	// }
	return cols
}

func BuildListDetailsNetLike(
	columns []map[string]interface{},
	listData []map[string]interface{},
	q admin.QueryRecordList,
	resultsCount int,
	detailErrors []map[string]string,
) map[string]interface{} {

	var filters interface{} = nil
	if strings.TrimSpace(q.Filters) != "" {
		// If you have DeserializeFilters in admin package, you can use it.
		// filters = admin.DeserializeFilters(q.Filters)
		filters = nil
	}

	var sortBy interface{} = nil
	if strings.TrimSpace(q.SortBy) != "" {
		sortBy = q.SortBy
	}

	var searchString interface{} = nil
	if strings.TrimSpace(q.Search) != "" {
		searchString = q.Search
	}

	return map[string]interface{}{
		"bHasSearchField": true,
		"columns":         columns,
		"customColumns":   nil,
		"filters":         filters,
		"listData":        listData,
		"summaryRows":     []interface{}{},
		"pageNumber":      q.PageNumber,
		"pageSize":        q.PageSize,
		"resultsCount":    resultsCount,
		"sortBy":          sortBy,
		"searchString":    searchString,
		"errors":          detailErrors, // .NET: details.errors = [{message}]
		"metadata":        map[string]interface{}{},
	}
}

func BuildCustomerAssetAccountsSPParamsNetLike(q admin.QueryRecordList, siteUsersId int, cfg admin.ListSPConfig) map[string]interface{} {
	rawSort := strings.TrimSpace(q.SortBy)
	rawFilter := strings.TrimSpace(q.Filters)
	rawSearch := strings.TrimSpace(q.Search)

	var rawSortParam interface{} = nil
	if rawSort != "" {
		rawSortParam = rawSort
	}
	var rawFilterParam interface{} = nil
	if rawFilter != "" {
		rawFilterParam = rawFilter
	}

	var sortByParam interface{} = nil
	if rawSort != "" {
		if sqlSort := admin.ConvertSortToSQL(rawSort, cfg.ColumnMap); strings.TrimSpace(sqlSort) != "" {
			sortByParam = sqlSort
		}
	}

	var filtersParam interface{} = nil
	if rawFilter != "" {
		if sqlFilters := admin.ConvertFiltersToSQL(rawFilter, cfg.ColumnMap); strings.TrimSpace(sqlFilters) != "" {
			filtersParam = sqlFilters
		}
	}

	var searchParam interface{} = nil
	if rawSearch != "" {
		sqlSearch := strings.TrimSpace(admin.ConvertSearchToSQL(rawSearch, cfg.SearchFields))
		if sqlSearch != "" {
			// some builders return " AND (...)" – make it WHERE based
			if strings.HasPrefix(strings.ToUpper(sqlSearch), "AND ") {
				sqlSearch = "WHERE 1=1 " + sqlSearch
			}
			searchParam = sqlSearch
		}
	}

	return map[string]interface{}{
		"User_SiteUsersID": siteUsersId,
		"PageNumber":       q.PageNumber,
		"PageSize":         q.PageSize,
		"ListKey":          cfg.ListKey,
		"TrackingID":       cfg.TrackingID,

		"RawSortString":   rawSortParam,   // NULL if empty
		"RawFilterString": rawFilterParam, // NULL if empty
		"RawSearchString": rawSearch,      // '' if empty (as per your preference)

		"SortBy":       sortByParam,  // NULL if empty
		"Filters":      filtersParam, // NULL if empty
		"SearchString": searchParam,  // NULL if empty (IMPORTANT)
		// some SPs also accept @Search:
		"Search": nilIfEmpty(rawSearch),
	}
}

func toInt(v interface{}) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case string:
		i, _ := strconv.Atoi(strings.TrimSpace(x))
		return i
	default:
		return 0
	}
}

// .NET-like SP params for v1_AdminRole_BusinessModule_List
func BuildBusinessListSPParamsNetLike(q admin.QueryRecordList, siteUsersId int) map[string]interface{} {
	return map[string]interface{}{
		"User_SiteUsersID": siteUsersId,

		"PageNumber": q.PageNumber,
		"PageSize":   q.PageSize,

		// SP signature has these
		"Filters":      nilIfEmpty(q.Filters),
		"SortBy":       nilIfEmpty(q.SortBy),
		"SearchString": nil, // IMPORTANT: This SP builds its own SQL; keep NULL unless you really build it.

		"RawFilterString": nilIfEmpty(q.Filters),
		"RawSortString":   nilIfEmpty(q.SortBy),

		// RawSearchString: agar search empty hai to NULL bhejo
		"RawSearchString": nilIfEmpty(q.Search),

		"ListKey":    "Business",
		"TrackingID": "DefaultTrackingID",
	}
}

func NormalizeColumnsDetailsNull(cols []map[string]interface{}) []map[string]interface{} {
	// .NET me TextContains/Amount ke details aksar null hotay hain ({} nahi)
	for _, c := range cols {
		if fm, ok := c["filterMetadata"].(map[string]interface{}); ok && fm != nil {
			// agar details {} ho to null kar do (except DateTime:Range jahan {} aa sakta hai)
			ft, _ := fm["filterType"].(string)
			if ft != "DateTime:Range" {
				fm["details"] = nil
			}
		}
	}
	return cols
}
func BusinessColumns() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"columnKey":   "Customers__Id",
			"labelKey":    "Id",
			"labelValue":  "Id",
			"orderNumber": 1,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    false,
			"bLocked":     false,
			"type":        "Integer",
			"filterMetadata": map[string]interface{}{
				"filterType": "Amount",
				"details":    nil,
			},
			"tooltip": nil,
		},
		{
			"columnKey":   "Customers__CustomersCode",
			"labelKey":    "CustomersCode",
			"labelValue":  "Code",
			"orderNumber": 2,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    true,
			"bLocked":     false,
			"type":        "String",
			"filterMetadata": map[string]interface{}{
				"filterType": "TextContains",
				"details":    nil,
			},
			"tooltip": nil,
		},
		{
			"columnKey":   "LicenseesBrands__StatementDescriptor",
			"labelKey":    "StatementDescriptor",
			"labelValue":  "Brand",
			"orderNumber": 3,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    true,
			"bLocked":     false,
			"type":        "String",
			"filterMetadata": map[string]interface{}{
				"filterType": "TextContains",
				"details":    nil,
			},
			"tooltip": nil,
		},
		{
			"columnKey":   "Customers__CompanyName",
			"labelKey":    "CompanyName",
			"labelValue":  "Name",
			"orderNumber": 4,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    true,
			"bLocked":     false,
			"type":        "String",
			"filterMetadata": map[string]interface{}{
				"filterType": "TextContains",
				"details":    nil,
			},
			"tooltip": nil,
		},
		{
			"columnKey":   "Customers__CompanyEmailAddress",
			"labelKey":    "CompanyEmailAddress",
			"labelValue":  "Email",
			"orderNumber": 5,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    true,
			"bLocked":     false,
			"type":        "String",
			"filterMetadata": map[string]interface{}{
				"filterType": "TextContains",
				"details":    nil,
			},
			"tooltip": nil,
		},

		// Docs Verified (SingleChoice 0/1)
		{
			"columnKey":   "Customers__bBusinessDocumentsVerified",
			"labelKey":    "bBusinessDocumentsVerified",
			"labelValue":  "Docs. Verified",
			"orderNumber": 6,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    true,
			"bLocked":     false,
			"type":        "Boolean",
			"filterMetadata": map[string]interface{}{
				"filterType": "SingleChoice",
				"details": map[string]interface{}{
					"PossibleValues": []map[string]string{
						{"value": "0", "label": "Unverified"},
						{"value": "1", "label": "Verified"},
					},
				},
			},
			"tooltip": nil,
		},

		// Financial Institution (0/1)
		{
			"columnKey":   "Customers__bFinancialInstitution",
			"labelKey":    "bFinancialInstitution",
			"labelValue":  "Financial Institution",
			"orderNumber": 7,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    true,
			"bLocked":     false,
			"type":        "Boolean",
			"filterMetadata": map[string]interface{}{
				"filterType": "SingleChoice",
				"details": map[string]interface{}{
					"PossibleValues": []map[string]string{
						{"value": "0", "label": "No"},
						{"value": "1", "label": "Yes"},
					},
				},
			},
			"tooltip": nil,
		},

		// Docs Submitted (0/1)
		{
			"columnKey":   "Customers__bAllDocumentsSubmitted",
			"labelKey":    "bAllDocumentsSubmitted",
			"labelValue":  "Docs. Submitted",
			"orderNumber": 8,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    true,
			"bLocked":     false,
			"type":        "Boolean",
			"filterMetadata": map[string]interface{}{
				"filterType": "SingleChoice",
				"details": map[string]interface{}{
					"PossibleValues": []map[string]string{
						{"value": "0", "label": "Unsubmitted"},
						{"value": "1", "label": "Submitted"},
					},
				},
			},
			"tooltip": nil,
		},

		// Virtual (0/1)
		{
			"columnKey":   "Customers__bVirtual",
			"labelKey":    "bVirtual",
			"labelValue":  "Virtual",
			"orderNumber": 9,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    true,
			"bLocked":     false,
			"type":        "Boolean",
			"filterMetadata": map[string]interface{}{
				"filterType": "SingleChoice",
				"details": map[string]interface{}{
					"PossibleValues": []map[string]string{
						{"value": "0", "label": "No"},
						{"value": "1", "label": "Yes"},
					},
				},
			},
			"tooltip": nil,
		},

		// Business Verification Status
		{
			"columnKey":   "Customers__BusinessVerificationStatus",
			"labelKey":    "BusinessVerificationStatus",
			"labelValue":  "Ver. Status",
			"orderNumber": 10,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    true,
			"bLocked":     false,
			"type":        "String",
			"filterMetadata": map[string]interface{}{
				"filterType": "TextContains",
				"details":    nil,
			},
			"tooltip": nil,
		},

		// Submitted Form (no FilterMetadata in .NET)
		{
			"columnKey":      "Customers__bSubmittedForm",
			"labelKey":       "bSubmittedForm",
			"labelValue":     "Submitted Form",
			"orderNumber":    11,
			"bSortable":      true,
			"bFilterable":    false,
			"bVisible":       true,
			"bLocked":        false,
			"type":           "Boolean",
			"filterMetadata": nil,
			"tooltip":        nil,
		},

		// Frozen (0/1)
		{
			"columnKey":   "Customers__bFrozen",
			"labelKey":    "bFrozen",
			"labelValue":  "Frozen",
			"orderNumber": 12,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    true,
			"bLocked":     false,
			"type":        "Boolean",
			"filterMetadata": map[string]interface{}{
				"filterType": "SingleChoice",
				"details": map[string]interface{}{
					"PossibleValues": []map[string]string{
						{"value": "0", "label": "Unfrozen"},
						{"value": "1", "label": "Frozen"},
					},
				},
			},
			"tooltip": nil,
		},

		// Add Date (DateTime:Range)
		{
			"columnKey":   "Customers__AddDate",
			"labelKey":    "AddDate",
			"labelValue":  "Add Date",
			"orderNumber": 13,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    true,
			"bLocked":     false,
			"type":        "DateTime",
			"filterMetadata": map[string]interface{}{
				"filterType": "DateTime:Range",
				"details":    map[string]interface{}{},
			},
			"tooltip": nil,
		},

		// CustomerUsers Id (hidden)
		{
			"columnKey":   "CustomerUsers__Id",
			"labelKey":    "Id",
			"labelValue":  "Id",
			"orderNumber": 14,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    false,
			"bLocked":     false,
			"type":        "Integer",
			"filterMetadata": map[string]interface{}{
				"filterType": "Amount",
				"details":    nil,
			},
			"tooltip": nil,
		},

		// New Account Available (hidden, not filterable/sortable)
		{
			"columnKey":      "CustomersAvailableAccounts__bNewAccountAvailable",
			"labelKey":       "bNewAccountAvailable",
			"labelValue":     "bNewAccountAvailable",
			"orderNumber":    15,
			"bSortable":      false,
			"bFilterable":    false,
			"bVisible":       false,
			"bLocked":        false,
			"type":           "Boolean",
			"filterMetadata": nil,
			"tooltip":        nil,
		},
	}
}

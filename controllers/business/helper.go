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
func BuildBusinessListSPParamsNetLike(q admin.QueryRecordList, siteUsersId int, cfg admin.ListSPConfig) map[string]interface{} {
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
		if spSort := admin.ConvertSortToSP(rawSort, cfg.ColumnMap); strings.TrimSpace(spSort) != "" {
			sortByParam = spSort
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

		"RawSortString":   rawSortParam,
		"RawFilterString": rawFilterParam,
		"RawSearchString": rawSearch, // '' when empty

		"SortBy":       sortByParam,
		"Filters":      filtersParam,
		"SearchString": searchParam,
		// "Search":       nilIfEmpty(rawSearch),
	}
}

// .NET-like SP params for v1_AdminRole_CustomersModule_List
func BuildCustomerListSPParamsNetLike(q admin.QueryRecordList, siteUsersId int, cfg admin.ListSPConfig) map[string]interface{} {
	// same behavior as business list, but ListKey differs
	return BuildBusinessListSPParamsNetLike(q, siteUsersId, cfg)
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

func CustomerColumns() []map[string]interface{} {
	return []map[string]interface{}{
		// 1) CustomerUsers__Id
		{
			"columnKey":   "CustomerUsers__Id",
			"labelKey":    "Id",
			"labelValue":  "Id",
			"orderNumber": 1,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    true, // .NET me visible explicitly false nahi hai
			"bLocked":     false,
			"type":        "Integer",
			"filterMetadata": map[string]interface{}{
				"filterType": "Amount",
				"details":    nil,
			},
			"tooltip": nil,
		},

		// 2) CustomerUsersCustomers__Id (hidden)
		{
			"columnKey":   "CustomerUsersCustomers__Id",
			"labelKey":    "Id",
			"labelValue":  "Id",
			"orderNumber": 2,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    false, // bVisible=false
			"bLocked":     false,
			"type":        "Integer",
			"filterMetadata": map[string]interface{}{
				"filterType": "Amount",
				"details":    nil,
			},
			"tooltip": nil,
		},

		// 3) Customers__Id (hidden)
		{
			"columnKey":   "Customers__Id",
			"labelKey":    "Id",
			"labelValue":  "Id",
			"orderNumber": 3,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    false, // bVisible=false
			"bLocked":     false,
			"type":        "Integer",
			"filterMetadata": map[string]interface{}{
				"filterType": "Amount",
				"details":    nil,
			},
			"tooltip": nil,
		},

		// 4) CustomerUsers__CustomerUsersCode (LabelValue="Code", searchable/filterable/sortable)
		{
			"columnKey":   "CustomerUsers__CustomerUsersCode",
			"labelKey":    "CustomerUsersCode",
			"labelValue":  "Code",
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

		// 5) Licensees__LicenseeName (LabelValue="Licensee")
		{
			"columnKey":      "Licensees__LicenseeName",
			"labelKey":       "LicenseeName",
			"labelValue":     "Licensee",
			"orderNumber":    5,
			"bSortable":      false,
			"bFilterable":    false,
			"bVisible":       true,
			"bLocked":        false,
			"type":           "String",
			"filterMetadata": nil,
			"tooltip":        nil,
		},

		// 6) LicenseesBrands__SiteName (LabelValue="Licensee Brand")
		{
			"columnKey":      "LicenseesBrands__SiteName",
			"labelKey":       "SiteName",
			"labelValue":     "Licensee Brand",
			"orderNumber":    6,
			"bSortable":      false,
			"bFilterable":    false,
			"bVisible":       true,
			"bLocked":        false,
			"type":           "String",
			"filterMetadata": nil,
			"tooltip":        nil,
		},

		// 7) Customers__AccountType (SingleChoice: Business,Personal,VirtualBusiness,VirtualPersonal)
		{
			"columnKey":   "Customers__AccountType",
			"labelKey":    "AccountType",
			"labelValue":  "AccountType",
			"orderNumber": 7,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    true,
			"bLocked":     false,
			"type":        "String",
			"filterMetadata": map[string]interface{}{
				"filterType": "SingleChoice",
				"details": map[string]interface{}{
					"PossibleValues": []map[string]string{
						{"value": "Business", "label": "Business"},
						{"value": "Personal", "label": "Personal"},
						{"value": "VirtualBusiness", "label": "VirtualBusiness"},
						{"value": "VirtualPersonal", "label": "VirtualPersonal"},
					},
				},
			},
			"tooltip": nil,
		},

		// 8) Customers__bVirtual (LabelValue="Virtual", SingleChoice 0/1)
		{
			"columnKey":   "Customers__bVirtual",
			"labelKey":    "bVirtual",
			"labelValue":  "Virtual",
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
						{"value": "0", "label": "No"},
						{"value": "1", "label": "Yes"},
					},
				},
			},
			"tooltip": nil,
		},

		// 9) Customers__CompanyName (hidden, not filterable/sortable)
		{
			"columnKey":      "Customers__CompanyName",
			"labelKey":       "CompanyName",
			"labelValue":     "CompanyName",
			"orderNumber":    9,
			"bSortable":      false,
			"bFilterable":    false,
			"bVisible":       false, // bVisible=false
			"bLocked":        false,
			"type":           "String",
			"filterMetadata": nil,
			"tooltip":        nil,
		},

		// 10) Customers__CompanyEmailAddress
		{
			"columnKey":   "Customers__CompanyEmailAddress",
			"labelKey":    "CompanyEmailAddress",
			"labelValue":  "CompanyEmailAddress",
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

		// 11) CustomerUsers__FirstName
		{
			"columnKey":   "CustomerUsers__FirstName",
			"labelKey":    "FirstName",
			"labelValue":  "FirstName",
			"orderNumber": 11,
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

		// 12) CustomerUsers__LastName
		{
			"columnKey":   "CustomerUsers__LastName",
			"labelKey":    "LastName",
			"labelValue":  "LastName",
			"orderNumber": 12,
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

		// 13) SiteUsers__EmailAddress
		{
			"columnKey":   "SiteUsers__EmailAddress",
			"labelKey":    "EmailAddress",
			"labelValue":  "EmailAddress",
			"orderNumber": 13,
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

		// 14) CustomerUsers__DateOfBirth (DateTime Range)
		{
			"columnKey":   "CustomerUsers__DateOfBirth",
			"labelKey":    "DateOfBirth",
			"labelValue":  "DateOfBirth",
			"orderNumber": 14,
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

		// 15) CustomerUsers__bDocumentVerified (LabelValue="Document Verified", SingleChoice)
		{
			"columnKey":   "CustomerUsers__bDocumentVerified",
			"labelKey":    "bDocumentVerified",
			"labelValue":  "Document Verified",
			"orderNumber": 15,
			"bSortable":   false,
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

		// 16) CustomerUsers__bRequiresManualVerification (LabelValue="Requires Assistance", SingleChoice)
		{
			"columnKey":   "CustomerUsers__bRequiresManualVerification",
			"labelKey":    "bRequiresManualVerification",
			"labelValue":  "Requires Assistance",
			"orderNumber": 16,
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

		// 17) CustomerUsers__bEligibleForManualVerification (LabelValue="Manual Verification Allowed", SingleChoice)
		{
			"columnKey":   "CustomerUsers__bEligibleForManualVerification",
			"labelKey":    "bEligibleForManualVerification",
			"labelValue":  "Manual Verification Allowed",
			"orderNumber": 17,
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

		// 18) CustomerUsers__VerificationDate (DateTime Range)
		{
			"columnKey":   "CustomerUsers__VerificationDate",
			"labelKey":    "VerificationDate",
			"labelValue":  "VerificationDate",
			"orderNumber": 18,
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

		// 19) CustomerUsers__VerificationStatus
		{
			"columnKey":   "CustomerUsers__VerificationStatus",
			"labelKey":    "VerificationStatus",
			"labelValue":  "VerificationStatus",
			"orderNumber": 19,
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

		// 20) SiteUsers__bSuppressed (LabelValue="Suppressed", SingleChoice Active/Inactive)
		{
			"columnKey":   "SiteUsers__bSuppressed",
			"labelKey":    "bSuppressed",
			"labelValue":  "Suppressed",
			"orderNumber": 20,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    true,
			"bLocked":     false,
			"type":        "Boolean",
			"filterMetadata": map[string]interface{}{
				"filterType": "SingleChoice",
				"details": map[string]interface{}{
					"PossibleValues": []map[string]string{
						{"value": "0", "label": "Active"},
						{"value": "1", "label": "Inactive"},
					},
				},
			},
			"tooltip": nil,
		},

		// 21) Customers__bFrozen (LabelValue="Frozen", SingleChoice Unfrozen/Frozen)
		{
			"columnKey":   "Customers__bFrozen",
			"labelKey":    "bFrozen",
			"labelValue":  "Frozen",
			"orderNumber": 21,
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

		// 22) CustomerUsers__AddDate (DateTime Range)
		{
			"columnKey":   "CustomerUsers__AddDate",
			"labelKey":    "AddDate",
			"labelValue":  "Add Date",
			"orderNumber": 22,
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

		// 23) CustomerUsers__bApiConfigurable (hidden)
		{
			"columnKey":      "CustomerUsers__bApiConfigurable",
			"labelKey":       "bApiConfigurable",
			"labelValue":     "bApiConfigurable",
			"orderNumber":    23,
			"bSortable":      false,
			"bFilterable":    false,
			"bVisible":       false,
			"bLocked":        false,
			"type":           "Boolean",
			"filterMetadata": nil,
			"tooltip":        nil,
		},

		// 24) CustomerUsersAvailableAccounts__bNewAccountAvailable (hidden)
		{
			"columnKey":      "CustomerUsersAvailableAccounts__bNewAccountAvailable",
			"labelKey":       "bNewAccountAvailable",
			"labelValue":     "bNewAccountAvailable",
			"orderNumber":    24,
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

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

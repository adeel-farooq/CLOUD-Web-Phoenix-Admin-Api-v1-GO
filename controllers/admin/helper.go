package admin

import (
	"cloud-web-phoenix-customer-v1-go/controllers/auth"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"cloud-web-phoenix-customer-v1-go/db"
	"encoding/json"
	// "github.com/jmoiron/sqlx"
)

func ParseQueryRecordList(q url.Values) QueryRecordList {
	parseBool := func(v string) bool { return strings.ToLower(strings.TrimSpace(v)) == "true" }
	parseInt := func(v string) int {
		i, _ := strconv.Atoi(strings.TrimSpace(v))
		return i
	}

	out := QueryRecordList{
		Filters:       q.Get("filters"),
		Search:        q.Get("search"),
		SortBy:        q.Get("sortBy"),
		CustomColumns: q.Get("customColumns"),
		PageNumber:    parseInt(q.Get("pageNumber")),
		PageSize:      parseInt(q.Get("pageSize")),
		BClearFilters: parseBool(q.Get("bClearFilters")),
		BClearSearch:  parseBool(q.Get("bClearSearch")),
		BClearSortBy:  parseBool(q.Get("bClearSortBy")),
		BResetColumns: parseBool(q.Get("bResetColumns")),
	}

	// defaults (.NET idea)
	if out.PageNumber == 0 {
		out.PageNumber = 1
	}
	if out.PageSize == 0 {
		out.PageSize = 10
	}
	return out
}

// --------------------
// .NET: OverrideWithExistingListSelections
// --------------------
func OverrideWithSelections(q *QueryRecordList, ex *ListSelections) {
	if ex == nil {
		return
	}

	// pagination defaults from DB selections (if missing)
	if q.PageNumber == 0 && ex.PageNumber > 0 {
		q.PageNumber = ex.PageNumber
	}
	if q.PageSize == 0 && ex.PageSize > 0 {
		q.PageSize = ex.PageSize
	}
	if q.PageNumber == 0 {
		q.PageNumber = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 10
	}

	// filters
	if q.BClearFilters {
		q.Filters = ""
	} else {
		if strings.TrimSpace(q.Filters) == "" && ex.FilterString != "" {
			q.Filters = ex.FilterString
		} else if ex.FilterString != "" && strings.TrimSpace(q.Filters) != "" && q.Filters != ex.FilterString {
			q.PageNumber = 1
		}
	}

	// search
	if q.BClearSearch {
		q.Search = ""
	} else {
		if strings.TrimSpace(q.Search) == "" && ex.FullTextSearchString != "" {
			q.Search = ex.FullTextSearchString
		} else if ex.FullTextSearchString != "" && strings.TrimSpace(q.Search) != "" && q.Search != ex.FullTextSearchString {
			q.PageNumber = 1
		}
	}

	// sort
	if q.BClearSortBy {
		q.SortBy = ""
	} else {
		if strings.TrimSpace(q.SortBy) == "" && ex.SortString != "" {
			q.SortBy = ex.SortString
		}
	}

	// custom columns
	if q.BResetColumns {
		q.CustomColumns = ""
	} else {
		if strings.TrimSpace(q.CustomColumns) == "" && ex.CustomColumnsString != "" {
			q.CustomColumns = ex.CustomColumnsString
		}
	}
}

// --------------------
// Selections row -> struct (ExecSP output row mapper)
// --------------------
func SelectionsFromRow(row map[string]interface{}) *ListSelections {
	if row == nil {
		return nil
	}
	getStr := func(k string) string {
		v := row[k]
		if v == nil {
			return ""
		}
		switch t := v.(type) {
		case string:
			return t
		case []byte:
			return string(t)
		default:
			return fmt.Sprint(t)
		}
	}
	getInt := func(k string) int {
		v := row[k]
		if v == nil {
			return 0
		}
		switch t := v.(type) {
		case int:
			return t
		case int64:
			return int(t)
		case float64:
			return int(t)
		default:
			i, _ := strconv.Atoi(fmt.Sprint(t))
			return i
		}
	}

	return &ListSelections{
		FilterString:         getStr("FilterString"),
		SortString:           getStr("SortString"),
		FullTextSearchString: getStr("FullTextSearchString"),
		CustomColumnsString:  getStr("CustomColumnsString"),
		PageNumber:           getInt("PageNumber"),
		PageSize:             getInt("PageSize"),
	}
}

// --------------------
// SQL builders (filters/sort/search) — whitelist based
// --------------------
func escapeSQL(val string) string { return strings.ReplaceAll(val, "'", "''") }

func resolveColumn(colMap map[string]string, key string) (string, bool) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", false
	}
	if col, ok := colMap[key]; ok {
		return col, true
	}
	// Many clients send PascalCase keys like "AdminUsers__Id" while our whitelist
	// uses lower-first-letter keys like "adminUsers__Id".
	alt := LowercaseFirstChar(key)
	if alt != key {
		if col, ok := colMap[alt]; ok {
			return col, true
		}
	}
	return "", false
}

func ConvertSortToSQL(sortStr string, colMap map[string]string) string {
	sortStr = strings.TrimSpace(sortStr)
	if sortStr == "" {
		return ""
	}

	parts := strings.Split(sortStr, "|")
	out := []string{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		chunks := strings.Fields(p)
		if len(chunks) != 2 {
			continue
		}
		key := chunks[0]
		dir := strings.ToUpper(chunks[1])
		if dir != "ASC" && dir != "DESC" {
			continue
		}
		col, ok := resolveColumn(colMap, key)
		if !ok {
			continue
		}
		out = append(out, col+" "+dir)
	}
	if len(out) == 0 {
		return ""
	}
	return " ORDER BY " + strings.Join(out, ", ")
}

func ConvertFiltersToSQL(filterStr string, colMap map[string]string) string {
	filterStr = strings.TrimSpace(filterStr)
	if filterStr == "" {
		return ""
	}

	re := regexp.MustCompile(`^(.+?)\s+([A-Z]+)\s+\((.+)\)$`)
	parts := strings.Split(filterStr, "|")
	sqlParts := []string{}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		m := re.FindStringSubmatch(part)
		if len(m) != 4 {
			continue
		}

		key := strings.TrimSpace(m[1])
		op := strings.TrimSpace(m[2])
		raw := strings.TrimSpace(m[3])

		col, ok := resolveColumn(colMap, key)
		if !ok {
			continue
		}

		val := escapeSQL(raw)

		switch op {
		case "EQ":
			sqlParts = append(sqlParts, col+" = '"+val+"'")
		case "CONTAINS":
			sqlParts = append(sqlParts, col+" LIKE '%"+val+"%'")
		case "BETWEEN":
			chunks := strings.Split(raw, "TO")
			if len(chunks) == 2 {
				from := escapeSQL(strings.TrimSpace(chunks[0]))
				to := escapeSQL(strings.TrimSpace(chunks[1]))
				if from != "" && to != "" {
					sqlParts = append(sqlParts, col+" BETWEEN '"+from+"' AND '"+to+"'")
				}
			}
		}
	}

	if len(sqlParts) == 0 {
		return ""
	}
	return " AND " + strings.Join(sqlParts, " AND ")
}

func ConvertSearchToSQL(search string, fields []string) string {
	search = strings.TrimSpace(search)
	if search == "" || len(fields) == 0 {
		return ""
	}
	s := escapeSQL(search)

	orParts := []string{}
	for _, f := range fields {
		orParts = append(orParts, f+" LIKE '%"+s+"%'")
	}
	return "WHERE (" + strings.Join(orParts, " OR ") + ")"
}

// --------------------
// SP Params builder (generic)
// --------------------
func BuildListSPParams(q QueryRecordList, siteUsersId int, cfg ListSPConfig) map[string]interface{} {
	params := map[string]interface{}{
		"User_SiteUsersID": siteUsersId,
		"PageNumber":       q.PageNumber,
		"PageSize":         q.PageSize,
		"ListKey":          cfg.ListKey,
		"TrackingID":       cfg.TrackingID,
		"RawSortString":    q.SortBy,
		"RawFilterString":  q.Filters,
		"RawSearchString":  q.Search,
	}

	if strings.TrimSpace(q.Filters) != "" {
		if sqlFilters := ConvertFiltersToSQL(q.Filters, cfg.ColumnMap); sqlFilters != "" {
			params["Filters"] = sqlFilters
		}
	}
	if strings.TrimSpace(q.SortBy) != "" {
		if sqlSort := ConvertSortToSQL(q.SortBy, cfg.ColumnMap); sqlSort != "" {
			params["SortBy"] = sqlSort
		}
	}
	if strings.TrimSpace(q.Search) != "" {
		if sqlSearch := ConvertSearchToSQL(q.Search, cfg.SearchFields); sqlSearch != "" {
			params["SearchString"] = sqlSearch
		}
	}

	return params
}

// --------------------
// Response details builder (generic)
// --------------------
func BuildListDetails(columns []map[string]interface{}, listData []map[string]interface{}, q QueryRecordList, resultsCount int) map[string]interface{} {
	var filters interface{}
	if strings.TrimSpace(q.Filters) != "" {
		filters = DeserializeFilters(q.Filters)
	}

	var sortBy interface{}
	if strings.TrimSpace(q.SortBy) != "" {
		sortBy = q.SortBy
	}

	var searchString interface{}
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
		"errors":          []interface{}{},
		"metadata":        map[string]interface{}{},
	}
}

// frontend compatible filters array
func DeserializeFilters(filterStr string) []map[string]interface{} {
	filterStr = strings.TrimSpace(filterStr)
	if filterStr == "" {
		return []map[string]interface{}{}
	}

	re := regexp.MustCompile(`^(.+?)\s+([A-Z]+)\s+\((.+)\)$`)
	parts := strings.Split(filterStr, "|")

	group := map[string][]map[string]interface{}{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		m := re.FindStringSubmatch(p)
		if len(m) != 4 {
			continue
		}
		col := strings.TrimSpace(m[1])
		op := strings.TrimSpace(m[2])
		val := strings.TrimSpace(m[3])

		group[col] = append(group[col], map[string]interface{}{
			"columnKey": col,
			"operator":  op,
			"value":     val,
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

func GetAdminUsersModuleListColumns() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"columnKey": "AdminUsers__Id", "labelKey": "Id", "labelValue": "Id",
			"orderNumber": 1, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false,
			"type":           "Integer",
			"filterMetadata": map[string]interface{}{"filterType": "Amount", "details": nil},
			"tooltip":        nil,
		},
		{
			"columnKey": "AdminUsers__AdminUsersCode", "labelKey": "AdminUsersCode", "labelValue": "Code",
			"orderNumber": 2, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false,
			"type":           "String",
			"filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil},
			"tooltip":        nil,
		},
		{
			"columnKey": "SiteUsers__EmailAddress", "labelKey": "EmailAddress", "labelValue": "Email Address",
			"orderNumber": 3, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false,
			"type":           "String",
			"filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil},
			"tooltip":        nil,
		},
		{
			"columnKey": "AdminUsers__FirstName", "labelKey": "FirstName", "labelValue": "First Name",
			"orderNumber": 4, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false,
			"type":           "String",
			"filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil},
			"tooltip":        nil,
		},
		{
			"columnKey": "AdminUsers__LastName", "labelKey": "LastName", "labelValue": "Last Name",
			"orderNumber": 5, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false,
			"type":           "String",
			"filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil},
			"tooltip":        nil,
		},
		{
			"columnKey": "SiteUsers__bSuppressed", "labelKey": "bSuppressed", "labelValue": "Suppressed",
			"orderNumber": 6, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false,
			"type": "Boolean",
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
		{
			"columnKey": "AdminUsers__AddDate", "labelKey": "AddDate", "labelValue": "Add Date",
			"orderNumber": 7, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false,
			"type":           "DateTime",
			"filterMetadata": map[string]interface{}{"filterType": "DateTime:Range", "details": map[string]interface{}{}},
			"tooltip":        nil,
		},
	}
}
func GetLicenseeAdminUsersModuleListColumns() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"columnKey": "AdminUsers__Id", "labelKey": "Id", "labelValue": "Id",
			"orderNumber": 1, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false,
			"type":           "Integer",
			"filterMetadata": map[string]interface{}{"filterType": "Amount", "details": nil},
			"tooltip":        nil,
		},
		{
			"columnKey": "AdminUsers__AdminUsersCode", "labelKey": "AdminUsersCode", "labelValue": "Code",
			"orderNumber": 2, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false,
			"type":           "String",
			"filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil},
			"tooltip":        nil,
		},
		{
			"columnKey": "SiteUsers__EmailAddress", "labelKey": "EmailAddress", "labelValue": "Email Address",
			"orderNumber": 3, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false,
			"type":           "String",
			"filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil},
			"tooltip":        nil,
		},
		{
			"columnKey": "AdminUsers__FirstName", "labelKey": "FirstName", "labelValue": "First Name",
			"orderNumber": 4, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false,
			"type":           "String",
			"filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil},
			"tooltip":        nil,
		},
		{
			"columnKey": "AdminUsers__LastName", "labelKey": "LastName", "labelValue": "Last Name",
			"orderNumber": 5, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false,
			"type":           "String",
			"filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil},
			"tooltip":        nil,
		},
		{
			"columnKey": "AdminUsers__JobTitle", "labelKey": "JobTitle", "labelValue": "Job Title",
			"orderNumber": 6, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false,
			"type":           "String",
			"filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil},
			"tooltip":        nil,
		},
		{
			"columnKey": "Licensees__LicenseeName", "labelKey": "LicenseeName", "labelValue": "Licensee Name",
			"orderNumber": 7, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false,
			"type":           "String",
			"filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil},
			"tooltip":        nil,
		},
		{
			"columnKey": "SiteUsers__bSuppressed", "labelKey": "bSuppressed", "labelValue": "Suppressed",
			"orderNumber": 8, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false,
			"type": "Boolean",
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
		{
			"columnKey": "AdminUsers__AddDate", "labelKey": "AddDate", "labelValue": "Add Date",
			"orderNumber": 9, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false,
			"type":           "DateTime",
			"filterMetadata": map[string]interface{}{"filterType": "DateTime:Range", "details": map[string]interface{}{}},
			"tooltip":        nil,
		},
	}
}
func GetAdminRolesModuleListColumns() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"columnKey":      "AdminRoles__Id",
			"labelKey":       "Id",
			"labelValue":     "Id",
			"orderNumber":    1,
			"bSortable":      false,
			"bFilterable":    false,
			"bVisible":       false,
			"bLocked":        false,
			"type":           "Integer",
			"filterMetadata": nil,
			"tooltip":        nil,
		},
		{
			"columnKey":   "AdminRoles__Name",
			"labelKey":    "Name",
			"labelValue":  "Name",
			"orderNumber": 2,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    true,
			"bLocked":     false,
			"type":        "String",
			"filterMetadata": map[string]interface{}{
				"filterType": "TextContains",
				"details":    map[string]interface{}{},
			},
			"tooltip": nil,
		},
		{
			"columnKey":   "AdminRoles__Level",
			"labelKey":    "Level",
			"labelValue":  "Level",
			"orderNumber": 3,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    true,
			"bLocked":     false,
			"type":        "String",
			"filterMetadata": map[string]interface{}{
				"filterType": "MultipleChoice",
				"details": map[string]interface{}{
					"PossibleValues": []map[string]string{
						{"value": "Admin", "label": "Admin"},
						{"value": "Licensee", "label": "Licensee"},
						{"value": "LicenseeBrand", "label": "Licensee Brand"},
					},
				},
			},
			"tooltip": nil,
		},
		{
			"columnKey":   "AdminRoles__bSuppressed",
			"labelKey":    "bSuppressed",
			"labelValue":  "Suppressed",
			"orderNumber": 4,
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
		{
			"columnKey":   "AdminRoles__AddDate",
			"labelKey":    "AddDate",
			"labelValue":  "Add Date",
			"orderNumber": 5,
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
	}
}

func LowercaseFirstChar(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func NormalizeRowKeys(row map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(row))

	for k, v := range row {
		// system fields skip (same as .NET DTO ignoring)
		if k == "HowManyResults" || k == "RowNum" {
			continue
		}
		out[LowercaseFirstChar(k)] = v
	}
	return out
}

func BuildColumnsFromSPRow(row map[string]interface{}, orderStart int) []map[string]interface{} {
	ignore := map[string]bool{
		"RowNum":         true,
		"HowManyResults": true,
	}

	typeGuess := func(v interface{}) (colType string, filterType string) {
		switch v.(type) {
		case int, int32, int64, float32, float64:
			return "Integer", "Amount"
		case bool:
			return "Boolean", "SingleChoice"
		default:
			// datetime guess
			s := fmt.Sprint(v)
			if strings.Contains(s, "T") || strings.Contains(s, ":") {
				// rough datetime detection
				return "DateTime", "DateTime:Range"
			}
			return "String", "TextContains"
		}
	}

	cols := []map[string]interface{}{}
	order := orderStart

	for k, v := range row {
		if ignore[k] {
			continue
		}

		colType, filterType := typeGuess(v)

		col := map[string]interface{}{
			"columnKey":   k, // NOTE: yahan same key rahegi (later normalize)
			"labelKey":    strings.ReplaceAll(strings.ReplaceAll(k, "__", " "), "_", " "),
			"labelValue":  strings.ReplaceAll(strings.ReplaceAll(k, "__", " "), "_", " "),
			"orderNumber": order,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    true,
			"bLocked":     false,
			"type":        colType,
			"tooltip":     nil,
		}

		// filterMetadata
		if filterType == "SingleChoice" && colType == "Boolean" {
			col["filterMetadata"] = map[string]interface{}{
				"filterType": "SingleChoice",
				"details": map[string]interface{}{
					"PossibleValues": []map[string]string{
						{"value": "0", "label": "Active"},
						{"value": "1", "label": "Inactive"},
					},
				},
			}
		} else {
			col["filterMetadata"] = map[string]interface{}{
				"filterType": filterType,
				"details":    map[string]interface{}{},
			}
		}

		cols = append(cols, col)
		order++
	}

	// IMPORTANT: map iteration random hoti hai, ordering fix karo
	sort.Slice(cols, func(i, j int) bool {
		return cols[i]["columnKey"].(string) < cols[j]["columnKey"].(string)
	})

	// Ab orderNumber ko final stable order me reset kar do
	for i := range cols {
		cols[i]["orderNumber"] = orderStart + i
	}

	return cols
}

func GetAdminRoleLevels() []map[string]interface{} {
	return []map[string]interface{}{
		{"label": "Admin", "value": "Admin"},
		{"label": "Licensee", "value": "Licensee"},
		{"label": "LicenseeBrand", "value": "LicenseeBrand"},
	}
}

func GetAdminRolesCreateMetadata() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name": "Name", "type": "String", "customType": nil, "label": "Name",
			"bRequired": true, "bRemoteDataSource": false, "dataSource": nil,
			"header": nil, "orderNumber": 0, "bVisible": false, "bSortable": false,
			"editable": false, "bFilterable": false,
		},
		{
			"name": "bSuppressed", "type": "Boolean", "customType": nil, "label": "Suppressed",
			"bRequired": true, "bRemoteDataSource": false, "dataSource": nil,
			"header": nil, "orderNumber": 0, "bVisible": false, "bSortable": false,
			"editable": false, "bFilterable": false,
		},
		{
			"name": "Level", "type": "SingleSelect", "customType": nil, "label": "Level",
			"bRequired": false, "bRemoteDataSource": false, "dataSource": "listAdminRoleLevels",
			"header": nil, "orderNumber": 0, "bVisible": false, "bSortable": false,
			"editable": false, "bFilterable": false,
		},
		{
			"name": "AccessRights", "type": "AccessRights", "customType": nil, "label": "Access Rights",
			"bRequired": true, "bRemoteDataSource": false, "dataSource": nil,
			"header": nil, "orderNumber": 0, "bVisible": false, "bSortable": false,
			"editable": false, "bFilterable": false,
		},
	}
}

// -------------- SP Call --------------

func LoadAdminRolesCreateDetails(level string) (map[string]interface{}, map[string]interface{}, error) {
	spName := "v1_AdminRole_AdminRolesModule_GetCreateDetails"
	params := map[string]interface{}{}
	if strings.TrimSpace(level) != "" {
		params["AdminRoleLevel"] = level
	}

	// SP returns single row: Status, Id, Details (json), Errors
	res, err := auth.ExecSP(db.DB, spName, params, 1)
	if err != nil {
		return nil, nil, err
	}

	row, ok := res.(map[string]interface{})
	if !ok {
		return nil, nil, err
	}

	// Parse Details JSON (SP JSON PATH nested strings issue)
	detailsStr, _ := row["Details"].(string)
	parsed, err := ParseAndFixNestedJSON(detailsStr)
	if err != nil {
		return row, nil, err
	}

	// Convert SP PascalCase keys -> frontend expected camelCase keys (for details/accessRights)
	details := BuildCreateDetailsForFrontend(parsed)

	return row, details, nil
}

// -------------- JSON Helpers --------------

// SP ke Details me nested JSON aksar string form me hota hai (AccessRights, ChildElements)
// Ye function usko recursively real array/map me convert kar deta hai.
func ParseAndFixNestedJSON(raw string) (map[string]interface{}, error) {
	out := map[string]interface{}{}

	if strings.TrimSpace(raw) == "" {
		return out, nil
	}

	// SQL output is already valid JSON; no need for replace, but safe handling:
	clean := strings.TrimSpace(raw)

	if err := json.Unmarshal([]byte(clean), &out); err != nil {
		return nil, err
	}

	fixed := fixNestedJSON(out)
	asMap, _ := fixed.(map[string]interface{})
	if asMap == nil {
		return map[string]interface{}{}, nil
	}
	return asMap, nil
}

func fixNestedJSON(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		for k, val := range t {
			t[k] = fixNestedJSON(val)
		}
		return t
	case []interface{}:
		for i := range t {
			t[i] = fixNestedJSON(t[i])
		}
		return t
	case string:
		s := strings.TrimSpace(t)
		if len(s) > 0 && (s[0] == '{' || s[0] == '[') {
			var x interface{}
			if err := json.Unmarshal([]byte(s), &x); err == nil {
				return fixNestedJSON(x)
			}
		}
		return t
	default:
		return v
	}
}

// -------------- Mapping Helpers --------------

func BuildCreateDetailsForFrontend(spDetails map[string]interface{}) map[string]interface{} {
	// SP keys: Id, Name, bSuppressed, Level, AccessRights (often decoded to []interface{})
	// Frontend keys: id, name, bSuppressed, level, listAdminRoleLevels, accessRights

	out := map[string]interface{}{}
	out["id"] = toInt(spDetails["Id"])
	out["name"] = spDetails["Name"]

	// bSuppressed SQL gives 0/1 (float64) sometimes
	out["bSuppressed"] = toBool(spDetails["bSuppressed"])

	out["level"] = spDetails["Level"]
	out["listAdminRoleLevels"] = GetAdminRoleLevels()

	// accessRights: rename keys inside tree
	out["accessRights"] = renameAccessTree(spDetails["AccessRights"])

	return out
}

func renameAccessTree(v interface{}) interface{} {
	switch t := v.(type) {
	case []interface{}:
		for i := range t {
			t[i] = renameAccessTree(t[i])
		}
		return t
	case map[string]interface{}:
		out := map[string]interface{}{}
		// SP keys: Id, DisplayName, Path, bHasAccess, ChildElements
		if val, ok := t["Id"]; ok {
			out["id"] = toInt(val)
		}
		if val, ok := t["DisplayName"]; ok {
			out["displayName"] = val
		}
		if val, ok := t["Path"]; ok {
			out["path"] = val
		}
		if val, ok := t["bHasAccess"]; ok {
			out["bHasAccess"] = toBool(val)
		}
		// ChildElements can be null or [] or string-json already fixed
		if val, ok := t["ChildElements"]; ok {
			out["childElements"] = renameAccessTree(val)
		} else {
			out["childElements"] = nil
		}
		return out
	default:
		return v
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
		// best-effort
		var i int
		_ = json.Unmarshal([]byte(x), &i)
		return i
	default:
		return 0
	}
}

func toBool(v interface{}) bool {
	switch x := v.(type) {
	case bool:
		return x
	case int:
		return x != 0
	case int64:
		return x != 0
	case float64:
		return x != 0
	case string:
		s := strings.ToLower(strings.TrimSpace(x))
		return s == "1" || s == "true" || s == "yes"
	default:
		return false
	}
}

func isValidAdminRoleLevel(level string) bool {
	switch level {
	case "Admin", "Licensee", "LicenseeBrand":
		return true
	default:
		return false
	}
}

func getAddedByFromToken(user map[string]interface{}) string {
	// JWT me tumhare sample ke mutabiq FirstName/LastName/UserCode hota hai
	fn := fmt.Sprint(user["FirstName"])
	ln := fmt.Sprint(user["LastName"])
	code := fmt.Sprint(user["UserCode"])

	full := strings.TrimSpace(strings.TrimSpace(fn) + " " + strings.TrimSpace(ln))
	if full == "" {
		full = code
	}
	// AddedBy field free text hai, is format se trace easy hota hai
	if code != "" && full != code {
		return full + " (" + code + ")"
	}
	return full
}

func buildAdminRoleCreateMetadata() []FormMeta {
	return []FormMeta{
		{
			Name: "Name", Type: "String", Label: "Name",
			BRequired: true, BRemoteDataSource: false, DataSource: nil,
			Header: nil, OrderNumber: 0, BVisible: false, BSortable: false,
			Editable: false, BFilterable: false,
		},
		{
			Name: "bSuppressed", Type: "Boolean", Label: "Suppressed",
			BRequired: true, BRemoteDataSource: false, DataSource: nil,
			Header: nil, OrderNumber: 0, BVisible: false, BSortable: false,
			Editable: false, BFilterable: false,
		},
		{
			Name: "Level", Type: "SingleSelect", Label: "Level",
			BRequired: false, BRemoteDataSource: false, DataSource: "listAdminRoleLevels",
			Header: nil, OrderNumber: 0, BVisible: false, BSortable: false,
			Editable: false, BFilterable: false,
		},
		{
			Name: "AccessRights", Type: "AccessRights", Label: "Access Rights",
			BRequired: true, BRemoteDataSource: false, DataSource: nil,
			Header: nil, OrderNumber: 0, BVisible: false, BSortable: false,
			Editable: false, BFilterable: false,
		},
	}
}

func parseDbResultRow(row map[string]interface{}) (DbResultRow, error) {
	out := DbResultRow{
		Id:      0,
		Status:  "0",
		Details: map[string]interface{}{},
		Errors:  []string{},
	}

	// Id
	if v, ok := row["Id"]; ok && v != nil {
		// ExecSP aksar int64 ya float64 de deta hai
		switch t := v.(type) {
		case int:
			out.Id = t
		case int64:
			out.Id = int(t)
		case float64:
			out.Id = int(t)
		default:
			// ignore
		}
	}

	// Status
	if v, ok := row["Status"]; ok && v != nil {
		out.Status = fmt.Sprint(v)
	}

	// Details JSON
	if v, ok := row["Details"]; ok && v != nil {
		s := fmt.Sprint(v)
		if strings.TrimSpace(s) != "" {
			var obj interface{}
			if err := json.Unmarshal([]byte(s), &obj); err == nil {
				out.Details = obj
			} else {
				// fallback: raw string
				out.Details = s
			}
		}
	}

	// Errors JSON
	if v, ok := row["Errors"]; ok && v != nil {
		s := fmt.Sprint(v)
		if strings.TrimSpace(s) != "" {
			var arr []string
			if err := json.Unmarshal([]byte(s), &arr); err == nil {
				out.Errors = arr
			} else {
				out.Errors = []string{s}
			}
		}
	}

	return out, nil
}

func spGetCreateDetails(level string) (DbResultRow, error) {
	sp := "v1_AdminRole_AdminRolesModule_GetCreateDetails"
	params := map[string]interface{}{
		"AdminRoleLevel": level,
	}
	res, err := auth.ExecSP(db.DB, sp, params, 1)
	if err != nil {
		return DbResultRow{}, err
	}
	rows, ok := res.([]map[string]interface{})
	if !ok || len(rows) == 0 {
		return DbResultRow{}, fmt.Errorf("empty SP response")
	}
	return parseDbResultRow(rows[0])
}

func spCreateAdminRole(siteUsersId int, addedBy string, req AdminRoleCreateRequest) (DbResultRow, error) {

	// SP expects comma-separated CDL in @AccessRightsCdl
	accessCdl := BuildAccessRightsCDL(req.AccessRights)

	sp := "v1_AdminRole_AdminRolesModule_Create"
	params := map[string]interface{}{
		"Name":             req.Name,
		"AdminRoleLevel":   req.Level,
		"bSuppressed":      req.BSuppressed,
		"AccessRightsCdl":  accessCdl,
		"User_SiteUsersID": siteUsersId,
		"User_AddedBy":     addedBy,
	}

	res, err := auth.ExecSP(db.DB, sp, params, 2)
	if err != nil {
		return DbResultRow{}, err
	}
	rows, ok := res.([]map[string]interface{})
	if !ok || len(rows) == 0 {
		return DbResultRow{}, fmt.Errorf("empty SP response")
	}
	return parseDbResultRow(rows[0])
}

func BuildAccessRightsCDL(nodes []AccessRightNode) string {
	seen := map[int]bool{}
	ids := make([]int, 0, 128)

	var walk func(list []AccessRightNode)
	walk = func(list []AccessRightNode) {
		for _, n := range list {
			if n.BHasAccess {
				if !seen[n.Id] {
					seen[n.Id] = true
					ids = append(ids, n.Id)
				}
			}
			if len(n.ChildElements) > 0 {
				walk(n.ChildElements)
			}
		}
	}
	walk(nodes)

	// join with comma
	if len(ids) == 0 {
		return ""
	}
	sb := strings.Builder{}
	for i, id := range ids {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(strconv.Itoa(id))
	}
	return sb.String()
}

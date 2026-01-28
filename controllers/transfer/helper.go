// helper.go
package transfer

import (
	"cloud-web-phoenix-customer-v1-go/controllers/admin"
	"encoding/json"
	"math"
	"strconv"

	"fmt"
	"strings"
)

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

func OutboundTransfersColumns() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"columnKey":   "OperationalAssetAccountsExternalTransfers__Id",
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
			"columnKey":   "OperationalAssetAccountsExternalTransfers__OperationalAssetAccountsExternalTransfersCode",
			"labelKey":    "Code",
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
			"columnKey":   "Assets__Code",
			"labelKey":    "Asset",
			"labelValue":  "Asset",
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
			"columnKey":   "OperationalAssetAccountsExternalTransfers__Reference",
			"labelKey":    "Reference",
			"labelValue":  "Reference",
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
			"columnKey":   "OperationalAssetAccountsExternalTransfers__Amount",
			"labelKey":    "Amount",
			"labelValue":  "Amount",
			"orderNumber": 5,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    true,
			"bLocked":     false,
			"type":        "Decimal",
			"filterMetadata": map[string]interface{}{
				"filterType": "Amount",
				"details":    nil,
			},
			"tooltip": nil,
		},
		{
			"columnKey":   "TransferStatus__Status",
			"labelKey":    "Status",
			"labelValue":  "Status",
			"orderNumber": 6,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    true,
			"bLocked":     false,
			"type":        "String",
			"filterMetadata": map[string]interface{}{
				"filterType": "MultipleChoice",
				"details": map[string]interface{}{
					"PossibleValues": []map[string]string{
						{"value": "Pending", "label": "Pending"},
						{"value": "Completed", "label": "Completed"},
						{"value": "Declined", "label": "Declined"},
					},
				},
			},
			"tooltip": nil,
		},
		{
			"columnKey":   "OperationalAssetAccountsExternalTransfers__AddDate",
			"labelKey":    "Date",
			"labelValue":  "Date",
			"orderNumber": 7,
			"bSortable":   true,
			"bFilterable": true,
			"bVisible":    true,
			"bLocked":     false,
			"type":        "DateTime",
			"filterMetadata": map[string]interface{}{
				"filterType": "DateTime:Range",
				"details":    nil,
			},
			"tooltip": nil,
		},
		{
			"columnKey":   "Products__ProductName",
			"labelKey":    "ProductName",
			"labelValue":  "Product",
			"orderNumber": 8,
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
	}
}

// NormalizeColumnsEmptyDetailsToNil converts `filterMetadata.details: {}` to `null`.
// Keeps non-empty objects as-is (e.g. PossibleValues for choice filters).
func NormalizeColumnsEmptyDetailsToNil(columns []map[string]interface{}) []map[string]interface{} {
	for _, col := range columns {
		fm, ok := col["filterMetadata"].(map[string]interface{})
		if !ok || fm == nil {
			continue
		}

		details, exists := fm["details"]
		if !exists || details == nil {
			// If details is null, whole filterMetadata should be null.
			col["filterMetadata"] = nil
			continue
		}

		switch d := details.(type) {
		case map[string]interface{}:
			if len(d) == 0 {
				fm["details"] = nil
			}
		case []interface{}:
			if len(d) == 0 {
				fm["details"] = nil
			}
		}

		// If details ended up nil, null out the entire filterMetadata.
		if fm["details"] == nil {
			col["filterMetadata"] = nil
		}
	}
	return columns
}

func StripSystemFields(row map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(row))
	for k, v := range row {
		if k == "HowManyResults" || k == "RowNum" {
			continue
		}
		key := UppercaseFirstChar(k)
		out[key] = normalizeOutboundTransfersValue(key, v)
	}
	return out
}

// NormalizeRowKeysNetLike lowercases only the first character of each key.
// This matches some existing list endpoints where SP returns PascalCase but JSON returns lower-first-letter keys.
func NormalizeRowKeysNetLike(row map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(row))
	for k, v := range row {
		if k == "HowManyResults" || k == "RowNum" {
			continue
		}
		out[admin.LowercaseFirstChar(k)] = v
	}
	return out
}

func BuildOutboundTransferParamsNetLike(q admin.QueryRecordList, siteUsersId int, cfg admin.ListSPConfig) map[string]interface{} {
	rawSort := strings.TrimSpace(q.SortBy)
	rawFilter := strings.TrimSpace(q.Filters)
	rawSearch := strings.TrimSpace(q.Search)

	// NOTE:
	// - RawSearchString: '' when empty
	// - RawSortString/RawFilterString: NULL when empty
	var rawSortParam interface{} = nil
	if rawSort != "" {
		rawSortParam = rawSort
	}
	var rawFilterParam interface{} = nil
	if rawFilter != "" {
		rawFilterParam = rawFilter
	}

	// Generated fragments should be NULL if not applicable (never empty string)
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
		if sqlSearch := admin.ConvertSearchToSQL(rawSearch, cfg.SearchFields); strings.TrimSpace(sqlSearch) != "" {
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
		"RawSearchString": rawSearch,

		"SortBy":       sortByParam,
		"Filters":      filtersParam,
		"SearchString": searchParam,
	}
}

func NormalizeOutboundRowNetLike(row map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(row))

	for k, v := range row {
		if k == "HowManyResults" || k == "RowNum" {
			continue
		}

		key := NormalizeListDataKeyCamelCaseNetLike(k)
		out[key] = NormalizeValueNetLike(v)
	}
	return out
}

func NormalizeListDataKeyCamelCaseNetLike(k string) string {
	// Match existing .NET behavior used across this API:
	// only lowercase the first character of the FULL key, leaving __ segments intact.
	// Example: InvalidPooledAccountTransfers__AddDate -> invalidPooledAccountTransfers__AddDate
	return admin.LowercaseFirstChar(k)
}

func normalizeKeySegmentCamelCase(s string) string {
	if s == "" {
		return s
	}

	out := admin.LowercaseFirstChar(s)

	// Common acronym normalization used by .NET JSON conventions.
	// e.g. PaymentID -> paymentId, ExternalID -> externalId
	if strings.HasSuffix(out, "ID") {
		out = strings.TrimSuffix(out, "ID") + "Id"
	} else if strings.HasSuffix(out, "IDs") {
		out = strings.TrimSuffix(out, "IDs") + "Ids"
	}

	return out
}

func NormalizeValueNetLike(v interface{}) interface{} {
	if s, ok := v.(string); ok {
		ss := strings.TrimSpace(s)

		// strip Z if present (.NET response me Z nahi aa raha)
		if strings.HasSuffix(ss, "Z") && strings.Contains(ss, "T") {
			ss = strings.TrimSuffix(ss, "Z")
			return ss
		}

		// decimal numeric string
		if looksLikeDecimal(ss) {
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

func looksLikeDecimal(s string) bool {
	if s == "" {
		return false
	}

	dot := false
	hasDigit := false
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
		hasDigit = true
	}
	return hasDigit
}

func FixColumnsNetLike(cols []map[string]interface{}) []map[string]interface{} {
	for _, c := range cols {
		bf, _ := c["bFilterable"].(bool)
		if !bf {
			c["filterMetadata"] = nil
			continue
		}

		if fm, ok := c["filterMetadata"].(map[string]interface{}); ok {
			if d, ok := fm["details"].(map[string]interface{}); ok && len(d) == 0 {
				fm["details"] = nil
			}
		}
	}
	return cols
}

func UppercaseFirstChar(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = []rune(strings.ToUpper(string(r[0])))[0]
	return string(r)
}

func normalizeOutboundTransfersValue(key string, v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch key {
	case "OperationalAssetAccountsExternalTransfers__Amount":
		if s, ok := asFixedDecimalString(v, 18); ok {
			return s
		}
		return v
	case "OperationalAssetAccountsExternalTransfers__AddDate",
		"OperationalAssetAccountsExternalTransfers__CompletionDate":
		if s, ok := v.(string); ok {
			return ensureZuluSuffix(strings.TrimSpace(s))
		}
		return v
	default:
		return v
	}
}

func ensureZuluSuffix(s string) string {
	if s == "" {
		return s
	}
	// Already has timezone info
	if strings.HasSuffix(s, "Z") || hasTimeZoneOffsetSuffix(s) {
		return s
	}
	// Looks like ISO-ish datetime but without timezone
	if strings.Contains(s, "T") {
		return s + "Z"
	}
	return s
}

func hasTimeZoneOffsetSuffix(s string) bool {
	// Accept trailing "+HH:MM" or "-HH:MM".
	if len(s) < 6 {
		return false
	}
	tail := s[len(s)-6:]
	if (tail[0] != '+' && tail[0] != '-') || tail[3] != ':' {
		return false
	}
	for _, idx := range []int{1, 2, 4, 5} {
		c := tail[idx]
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func asFixedDecimalString(v interface{}, scale int) (string, bool) {
	if scale < 0 {
		scale = 0
	}

	// If already string, pad/truncate fractional part without rounding.
	if s, ok := v.(string); ok {
		s = strings.TrimSpace(s)
		if s == "" {
			return "", true
		}
		// keep sign, ignore thousands separators if any
		s = strings.ReplaceAll(s, ",", "")
		if strings.ContainsAny(s, "eE") {
			// fallback: don't try to reformat scientific safely
			return s, true
		}
		parts := strings.SplitN(s, ".", 2)
		if len(parts) == 1 {
			return parts[0] + "." + strings.Repeat("0", scale), true
		}
		intPart := parts[0]
		frac := parts[1]
		if len(frac) == scale {
			return intPart + "." + frac, true
		}
		if len(frac) > scale {
			return intPart + "." + frac[:scale], true
		}
		return intPart + "." + frac + strings.Repeat("0", scale-len(frac)), true
	}

	switch n := v.(type) {
	case int:
		return fmt.Sprintf("%d.%s", n, strings.Repeat("0", scale)), true
	case int64:
		return fmt.Sprintf("%d.%s", n, strings.Repeat("0", scale)), true
	case float64:
		return fmt.Sprintf("%.*f", scale, n), true
	case float32:
		return fmt.Sprintf("%.*f", scale, float64(n)), true
	default:
		return "", false
	}
}
func BuildListSPParams(q admin.QueryRecordList, siteUsersId int, cfg admin.ListSPConfig) map[string]interface{} {
	params := map[string]interface{}{
		"User_SiteUsersID": siteUsersId,
		"PageNumber":       q.PageNumber,
		"PageSize":         q.PageSize,
		"ListKey":          cfg.ListKey,
		"TrackingID":       cfg.TrackingID,

		// always send Raw strings (never nil)
		"RawSortString":   strings.TrimSpace(q.SortBy),
		"RawFilterString": strings.TrimSpace(q.Filters),
		"RawSearchString": strings.TrimSpace(q.Search),
	}

	// ✅ For .NET-like SPs, stop here.
	// if cfg.NetLikeRawOnly {
	// 	return params
	// }

	// (old behavior for other endpoints)
	if strings.TrimSpace(q.Filters) != "" {
		if sqlFilters := admin.ConvertFiltersToSQL(q.Filters, cfg.ColumnMap); sqlFilters != "" {
			params["Filters"] = sqlFilters
		}
	}
	if strings.TrimSpace(q.SortBy) != "" {
		if sqlSort := admin.ConvertSortToSQL(q.SortBy, cfg.ColumnMap); sqlSort != "" {
			params["SortBy"] = sqlSort
		}
	}
	if strings.TrimSpace(q.Search) != "" {
		if sqlSearch := admin.ConvertSearchToSQL(q.Search, cfg.SearchFields); sqlSearch != "" {
			params["SearchString"] = sqlSearch
		}
	}

	return params
}

// BuildListDetailsNetLike: same output shape as .NET details block,
// without touching admin package.
func BuildListDetailsNetLike(
	columns []map[string]interface{},
	listData []map[string]interface{},
	q admin.QueryRecordList,
	resultsCount int,
	detailErrors []map[string]string,
) map[string]interface{} {

	// match .NET: null when empty
	var filters interface{}
	if strings.TrimSpace(q.Filters) != "" {
		filters = admin.DeserializeFilters(q.Filters)
	} else {
		filters = nil
	}

	var sortBy interface{}
	if strings.TrimSpace(q.SortBy) != "" {
		sortBy = q.SortBy
	} else {
		sortBy = nil
	}

	var searchString interface{}
	if strings.TrimSpace(q.Search) != "" {
		searchString = q.Search
	} else {
		searchString = nil
	}

	if detailErrors == nil {
		detailErrors = []map[string]string{}
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
		"errors":          detailErrors, // ✅ [{message:"..."}]
		"metadata":        map[string]interface{}{},
	}
}
func BuildListSPParamsNetLike(q admin.QueryRecordList, siteUsersId int, cfg admin.ListSPConfig) map[string]interface{} {
	rawSort := strings.TrimSpace(q.SortBy)
	rawFilter := strings.TrimSpace(q.Filters)
	rawSearch := strings.TrimSpace(q.Search)

	var rawSortVal interface{} = nil
	if rawSort != "" {
		rawSortVal = rawSort
	}
	var rawFilterVal interface{} = nil
	if rawFilter != "" {
		rawFilterVal = rawFilter
	}
	var rawSearchVal interface{} = nil
	if rawSearch != "" {
		rawSearchVal = rawSearch
	}

	params := map[string]interface{}{
		"PageSize":         q.PageSize,
		"ListKey":          cfg.ListKey,
		"TrackingID":       cfg.TrackingID,
		"RawSortString":    rawSortVal,
		"RawFilterString":  rawFilterVal,
		"RawSearchString":  rawSearchVal,
		"User_SiteUsersID": siteUsersId,
		"PageNumber":       q.PageNumber,
	}

	// This SP expects "missing" optional params as NULL (not empty string).
	// It uses Raw* strings for parsing internally, so keep these derived fields NULL.
	params["SearchString"] = nil
	params["Filters"] = nil
	params["SortBy"] = nil

	return params
}

func BuildOutboundTransfersSearchString(s string) string {
	s = strings.ReplaceAll(s, "'", "''") // basic escape

	return " AND (" +
		"OperationalAssetAccountsExternalTransfers.OperationalAssetAccountsExternalTransfersCode LIKE '%" + s + "%' " +
		"OR OperationalAssetAccountsExternalTransfers.Description LIKE '%" + s + "%' " +
		"OR OperationalAssetAccountsExternalTransfers.Reference LIKE '%" + s + "%' " +
		"OR Assets.Code LIKE '%" + s + "%' " +
		"OR Assets.Name LIKE '%" + s + "%' " +
		"OR TransferStatus.Status LIKE '%" + s + "%' " +
		"OR Products.ProductName LIKE '%" + s + "%'" +
		")"
}
func ApplyFilterMetaNetLike(col map[string]interface{}) map[string]interface{} {
	bFilterable, _ := col["bFilterable"].(bool)
	if !bFilterable {
		col["filterMetadata"] = nil
		return col
	}

	// filterable => filterMetadata must exist
	ft, _ := col["filterType"].(string) // hum temp key use karenge
	if ft == "" {
		ft = "TextContains"
	}

	// .NET me details aksar null hota hai (not {})
	col["filterMetadata"] = map[string]interface{}{
		"filterType": ft,
		"details":    nil,
	}
	return col
}
func NormalizeDecimalNetLike(v interface{}) interface{} {
	switch x := v.(type) {
	case float64, float32, int, int64:
		return v
	case []byte:
		return NormalizeDecimalNetLike(string(x))
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return nil
		}
		// try decimal -> float
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return x
		}
		// if whole number, return int (like 18)
		if math.Abs(f-math.Round(f)) < 0.0000001 {
			return int(math.Round(f))
		}
		return f
	default:
		return v
	}
}

func BuildListSPParamsUnAppliedTransfersNetLike(q admin.QueryRecordList, siteUsersId int, cfg admin.ListSPConfig) map[string]interface{} {
	rawSort := strings.TrimSpace(q.SortBy)
	rawFilter := strings.TrimSpace(q.Filters)
	rawSearch := strings.TrimSpace(q.Search)

	// Raw* always exist (SP persist merge uses these)
	params := map[string]interface{}{
		"PageNumber":       q.PageNumber,
		"PageSize":         q.PageSize,
		"ListKey":          cfg.ListKey,
		"TrackingID":       cfg.TrackingID,
		"User_SiteUsersID": siteUsersId,

		"RawSortString":   rawSort,
		"RawFilterString": rawFilter,
		"RawSearchString": rawSearch,

		// VERY IMPORTANT: these 3 should be NULL when empty
		"Filters":      nil,
		"SortBy":       nil,
		"SearchString": nil,
	}

	// SortBy: SP default sort uses COALESCE(@SortBy, '... DESC')
	// so only send when actual sort exists (otherwise keep NULL)
	if rawSort != "" {
		// yahan tum apna existing admin.ConvertSortToSQL(...) use kar sakte ho
		// lekin is SP me @SortBy expects something like: "InvalidPooledAccountTransfers__AddDate DESC"
		// agar tum RawSortString ko SQL me convert kar rahe ho to:
		// params["SortBy"] = admin.ConvertSortToSQL(rawSort, <columnMap>)
		// filhal simplest: NULL rakho jab tak mapping ready na ho
	}

	// Filters: SP expects " AND ...."
	if rawFilter != "" {
		// yahan tum admin.ConvertFiltersToSQL(...) use kar sakte ho
		// params["Filters"] = admin.ConvertFiltersToSQL(rawFilter, <columnMap>)
	}

	// Search: SP expects: "WHERE 1=1 AND (... LIKE ...)"
	if rawSearch != "" {
		params["SearchString"] = BuildUnAppliedTransfersSearchSQL(rawSearch)
	}

	return params
}

func BuildUnAppliedTransfersSearchSQL(term string) string {
	t := strings.TrimSpace(term)
	if t == "" {
		return ""
	}
	// SQL LIKE safe basic escaping for % and _
	esc := strings.ReplaceAll(t, "%", "[%]")
	esc = strings.ReplaceAll(esc, "_", "[_]")
	like := "%" + esc + "%"

	// NOTE: SP joins TransferTypes and selects InvalidPooledAccountTransfers.Product/Asset etc
	// So columns are exactly these table columns:
	return "WHERE 1=1 AND (" +
		"InvalidPooledAccountTransfers.PaymentID LIKE '" + like + "' OR " +
		"TransferTypes.[Type] LIKE '" + like + "' OR " +
		"InvalidPooledAccountTransfers.Product LIKE '" + like + "' OR " +
		"InvalidPooledAccountTransfers.Asset LIKE '" + like + "' OR " +
		"InvalidPooledAccountTransfers.ReferenceGiven LIKE '" + like + "' OR " +
		"InvalidPooledAccountTransfers.CorrectedReference LIKE '" + like + "'" +
		")"
}

func UnAppliedTransfersColumns() []map[string]interface{} {
	// NOTE:
	// - jahan bFilterable=false => filterMetadata MUST be null (.NET)
	// - bool filter metadata single choice (0|No,1|Yes) .NET me hai

	return []map[string]interface{}{
		{
			"columnKey":      "InvalidPooledAccountTransfers__Id",
			"labelKey":       "InvalidPooledAccountTransfers__Id",
			"labelValue":     "InvalidPooledAccountTransfers__Id",
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
			"columnKey":      "InvalidPooledAccountTransfers__PaymentID",
			"labelKey":       "PaymentID",
			"labelValue":     "Payment ID",
			"orderNumber":    2,
			"bSortable":      true,
			"bFilterable":    true,
			"bVisible":       true,
			"bLocked":        false,
			"type":           "String",
			"filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil},
			"tooltip":        nil,
		},
		{
			"columnKey":      "InvalidPooledAccountTransfers__ExternalID",
			"labelKey":       "ExternalID",
			"labelValue":     "External ID",
			"orderNumber":    3,
			"bSortable":      true,
			"bFilterable":    false,
			"bVisible":       false,
			"bLocked":        false,
			"type":           "String",
			"filterMetadata": nil,
			"tooltip":        nil,
		},
		{
			"columnKey":      "TransferTypes__Type",
			"labelKey":       "TransferType",
			"labelValue":     "Transfer Type",
			"orderNumber":    4,
			"bSortable":      true,
			"bFilterable":    true,
			"bVisible":       true,
			"bLocked":        false,
			"type":           "String",
			"filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil},
			"tooltip":        nil,
		},
		{
			"columnKey":      "InvalidPooledAccountTransfers__Product",
			"labelKey":       "Product",
			"labelValue":     "Product",
			"orderNumber":    5,
			"bSortable":      true,
			"bFilterable":    true,
			"bVisible":       true,
			"bLocked":        false,
			"type":           "String",
			"filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil},
			"tooltip":        nil,
		},
		{
			"columnKey":      "InvalidPooledAccountTransfers__Asset",
			"labelKey":       "Asset",
			"labelValue":     "Asset",
			"orderNumber":    6,
			"bSortable":      true,
			"bFilterable":    true,
			"bVisible":       true,
			"bLocked":        false,
			"type":           "String",
			"filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil},
			"tooltip":        nil,
		},
		{
			"columnKey":      "Assets__Symbol",
			"labelKey":       "Assets__Symbol",
			"labelValue":     "Assets__Symbol",
			"orderNumber":    7,
			"bSortable":      false,
			"bFilterable":    false,
			"bVisible":       false,
			"bLocked":        false,
			"type":           "String",
			"filterMetadata": nil,
			"tooltip":        nil,
		},
		{
			"columnKey":      "InvalidPooledAccountTransfers__Amount",
			"labelKey":       "Amount",
			"labelValue":     "Amount",
			"orderNumber":    8,
			"bSortable":      true,
			"bFilterable":    false,
			"bVisible":       true,
			"bLocked":        false,
			"type":           "Decimal",
			"filterMetadata": nil,
			"tooltip":        nil,
		},
		{
			"columnKey":      "InvalidPooledAccountTransfers__ReferenceGiven",
			"labelKey":       "ReferenceGiven",
			"labelValue":     "Given Reference",
			"orderNumber":    9,
			"bSortable":      true,
			"bFilterable":    true,
			"bVisible":       true,
			"bLocked":        false,
			"type":           "String",
			"filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil},
			"tooltip":        nil,
		},
		{
			"columnKey":      "InvalidPooledAccountTransfers__CorrectedReference",
			"labelKey":       "CorrectedReference",
			"labelValue":     "Corrected Reference",
			"orderNumber":    10,
			"bSortable":      true,
			"bFilterable":    true,
			"bVisible":       true,
			"bLocked":        false,
			"type":           "String",
			"filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil},
			"tooltip":        nil,
		},
		{
			"columnKey":   "InvalidPooledAccountTransfers__bTransactionFailed",
			"labelKey":    "Failed",
			"labelValue":  "Failed",
			"orderNumber": 11,
			"bSortable":   false,
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
		{
			"columnKey":   "InvalidPooledAccountTransfers__bReversed",
			"labelKey":    "Reversed",
			"labelValue":  "Reversed",
			"orderNumber": 12,
			"bSortable":   false,
			"bFilterable": true,
			"bVisible":    false,
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
		{
			"columnKey":      "InvalidPooledAccountTransfers__AddDate",
			"labelKey":       "Add Date",
			"labelValue":     "Add Date",
			"orderNumber":    13,
			"bSortable":      true,
			"bFilterable":    false,
			"bVisible":       true,
			"bLocked":        false,
			"type":           "DateTime",
			"filterMetadata": nil,
			"tooltip":        nil,
		},
		{
			"columnKey":      "InvalidPooledAccountTransfers__PendingApproval",
			"labelKey":       "Pending Approval",
			"labelValue":     "Pending Approval",
			"orderNumber":    14,
			"bSortable":      false,
			"bFilterable":    false,
			"bVisible":       false,
			"bLocked":        false,
			"type":           "Boolean",
			"filterMetadata": nil,
			"tooltip":        nil,
		},
		{
			"columnKey":      "InvalidPooledAccountTransfers__PendingProcessing",
			"labelKey":       "Pending Processing",
			"labelValue":     "Pending Processing",
			"orderNumber":    15,
			"bSortable":      false,
			"bFilterable":    false,
			"bVisible":       false,
			"bLocked":        false,
			"type":           "Boolean",
			"filterMetadata": nil,
			"tooltip":        nil,
		},
		{
			"columnKey":      "InvalidPooledAccountTransfers__Status",
			"labelKey":       "Status",
			"labelValue":     "Status",
			"orderNumber":    16,
			"bSortable":      false,
			"bFilterable":    false,
			"bVisible":       true,
			"bLocked":        false,
			"type":           "String",
			"filterMetadata": nil,
			"tooltip":        nil,
		},
	}
}

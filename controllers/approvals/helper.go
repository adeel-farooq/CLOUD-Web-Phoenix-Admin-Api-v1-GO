package approvals

import (
	"cloud-web-phoenix-customer-v1-go/controllers/auth"
	"cloud-web-phoenix-customer-v1-go/db"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/golang-jwt/jwt/v5"
)

// --------------------
// Query parsing
// --------------------
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

    if out.PageNumber == 0 {
        out.PageNumber = 1
    }
    if out.PageSize == 0 {
        out.PageSize = 10
    }
    return out
}

// --------------------
// Selections override
// --------------------
func OverrideWithSelections(q *QueryRecordList, ex *ListSelections) {
    if ex == nil {
        return
    }
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
    if q.BClearFilters {
        q.Filters = ""
    } else if strings.TrimSpace(q.Filters) == "" && ex.FilterString != "" {
        q.Filters = ex.FilterString
    }
    if q.BClearSearch {
        q.Search = ""
    } else if strings.TrimSpace(q.Search) == "" && ex.FullTextSearchString != "" {
        q.Search = ex.FullTextSearchString
    }
    if q.BClearSortBy {
        q.SortBy = ""
    } else if strings.TrimSpace(q.SortBy) == "" && ex.SortString != "" {
        q.SortBy = ex.SortString
    }
    if q.BResetColumns {
        q.CustomColumns = ""
    } else if strings.TrimSpace(q.CustomColumns) == "" && ex.CustomColumnsString != "" {
        q.CustomColumns = ex.CustomColumnsString
    }
}

// --------------------
// Load selections from SP
// --------------------
func LoadSelections(siteUsersId int, trackingId string, listKey string) *ListSelections {
    sp := "v1_General_ListSelectionsModule_GetSiteUsersListSelections"
    params := map[string]interface{}{
        "SiteUsersId": siteUsersId,
        "TrackingId":  trackingId,
        "ListKey":     listKey,
    }
    res, err := auth.ExecSP(db.DB, sp, params, 1)
    if err != nil {
        return nil
    }
    row, ok := auth.AsSingleRow(res)
    if !ok {
        return nil
    }
    return SelectionsFromRow(row)
}

// --------------------
// Selections row mapper
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
// SP Params builder
// --------------------
func BuildListSPParams(q QueryRecordList, siteUsersId int, cfg ListSPConfig) map[string]interface{} {
    return map[string]interface{}{
        "User_SiteUsersID": siteUsersId,
        "PageNumber":       q.PageNumber,
        "PageSize":         q.PageSize,
        "ListKey":          cfg.ListKey,
        "TrackingID":       cfg.TrackingID,
        "RawSortString":    q.SortBy,
        "RawFilterString":  q.Filters,
        "RawSearchString":  q.Search,
    }
}


// --------------------
// Row normalization
// --------------------
func NormalizeRowKeys(row map[string]interface{}) map[string]interface{} {
    out := make(map[string]interface{}, len(row))
    for k, v := range row {
        if k == "HowManyResults" || k == "RowNum" {
            continue
        }
        out[LowercaseFirstChar(k)] = v
    }
    return out
}

func LowercaseFirstChar(s string) string {
    if s == "" {
        return s
    }
    r := []rune(s)
    r[0] = unicode.ToLower(r[0])
    return string(r)
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
        i, _ := strconv.Atoi(x)
        return i
    default:
        return 0
    }
}

// --------------------
// Response builder
// --------------------
func BuildListDetails(columns []map[string]interface{}, listData []map[string]interface{}, q QueryRecordList, resultsCount int) map[string]interface{} {
    return map[string]interface{}{
        "bHasSearchField": true,
        "columns":         columns,
        "customColumns":   nil,
        "filters":         q.Filters,
        "listData":        listData,
        "summaryRows":     []interface{}{},
        "pageNumber":      q.PageNumber,
        "pageSize":        q.PageSize,
        "resultsCount":    resultsCount,
        "sortBy":          q.SortBy,
        "searchString":    q.Search,
        "errors":          []interface{}{},
        "metadata":        map[string]interface{}{},
    }
}

// --------------------
// Approval specific columns
// --------------------
func approvalsColumnMap() map[string]string {
    return map[string]string{
        "approvals__Id":                "Approvals.Id",
        "approvals__Code":              "Approvals.Code",
        "approvals__Type":              "Approvals.Type",
        "approvals__ApprovalsRequired": "Approvals.ApprovalsRequired",
        "approvals__ApprovalsReceived": "Approvals.ApprovalsReceived",
        "approvals__Description":       "Approvals.Description",
        "approvals__bApproved":         "Approvals.bApproved",
        "approvals__AddDate":           "Approvals.AddDate",
    }
}


// --------------------
// Helper: extract user ID from JWT claims
// --------------------
func getUserIdFromClaims(claims jwt.MapClaims) (int, bool) {
    // First try "id"
    if idVal, ok := claims["id"]; ok && idVal != nil {
        switch v := idVal.(type) {
        case float64:
            return int(v), true
        case int:
            return v, true
        case string:
            if idInt, err := strconv.Atoi(v); err == nil {
                return idInt, true
            }
        }
    }

    // Fallback: try "sub"
    if subVal, ok := claims["sub"]; ok && subVal != nil {
        switch v := subVal.(type) {
        case float64:
            return int(v), true
        case int:
            return v, true
        case string:
            if idInt, err := strconv.Atoi(v); err == nil {
                return idInt, true
            }
        }
    }

    return 0, false
}



func GetApprovalsModuleListColumns() []map[string]interface{} {
    return []map[string]interface{}{
        {
            "columnKey": "Approvals__Id", "labelKey": "Id", "labelValue": "Id", "orderNumber": 1,
            "bSortable": true, "bFilterable": true, "bVisible": true, "type": "Integer",
            "filterMetadata": map[string]interface{}{"filterType": "Amount", "details": nil},
        },
        {
            "columnKey": "Approvals__Code", "labelKey": "Code", "labelValue": "Code", "orderNumber": 2,
            "bSortable": true, "bFilterable": true, "bVisible": true, "type": "String",
            "filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil},
        },
        {
            "columnKey": "Approvals__Type", "labelKey": "Type", "labelValue": "Type", "orderNumber": 3,
            "bSortable": true, "bFilterable": true, "bVisible": true, "type": "String",
            "filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil},
        },
        {
            "columnKey": "Approvals__ApprovalsRequired", "labelKey": "ApprovalsRequired", "labelValue": "Approvals Required", "orderNumber": 4,
            "bSortable": true, "bFilterable": true, "bVisible": false, "type": "Integer",
            "filterMetadata": map[string]interface{}{"filterType": "Amount", "details": nil},
        },
        {
            "columnKey": "Approvals__ApprovalsReceived", "labelKey": "ApprovalsReceived", "labelValue": "Approvals Received", "orderNumber": 5,
            "bSortable": true, "bFilterable": true, "bVisible": true, "type": "Integer",
            "filterMetadata": map[string]interface{}{"filterType": "Amount", "details": nil},
        },
        {
            "columnKey": "Approvals__Description", "labelKey": "Description", "labelValue": "Approval Description", "orderNumber": 6,
            "bSortable": true, "bFilterable": true, "bVisible": true, "type": "String",
            "filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil},
        },
        {
            "columnKey": "Approvals__bApproved", "labelKey": "bApproved", "labelValue": "Status", "orderNumber": 7,
            "bSortable": true, "bFilterable": true, "bVisible": true, "type": "Boolean",
            "filterMetadata": map[string]interface{}{
                "filterType": "SingleChoice",
                "details": map[string]interface{}{
                    "PossibleValues": []map[string]string{
                        {"value": "0", "label": "Not Approved"},
                        {"value": "1", "label": "Approved"},
                    },
                },
            },
        },
        {
            "columnKey": "Approvals__AddDate", "labelKey": "AddDate", "labelValue": "Add Date", "orderNumber": 8,
            "bSortable": true, "bFilterable": true, "bVisible": true, "type": "DateTime",
            "filterMetadata": map[string]interface{}{"filterType": "DateTime:Range", "details": map[string]interface{}{}},
        },
    }
}

// --------------------
// Approval view metadata
// --------------------
func GetApprovalsViewMetadata() []map[string]interface{} {
    return []map[string]interface{}{
        {"name": "Code", "type": "String", "customType": nil, "label": "Code",
            "bRequired": false, "bRemoteDataSource": false, "bVisible": false},
        {"name": "Type", "type": "String", "customType": nil, "label": "Type",
            "bRequired": false, "bRemoteDataSource": false, "bVisible": false},
        {"name": "Description", "type": "Custom", "customType": "TextArea", "label": "Description",
            "bRequired": false, "bRemoteDataSource": false, "bVisible": false},
        {"name": "DetailsJson", "type": "Custom", "customType": "JsonTable", "label": "Details",
            "bRequired": false, "bRemoteDataSource": false, "bVisible": false},
        {"name": "bApproved", "type": "Boolean", "customType": nil, "label": "Approved",
            "bRequired": false, "bRemoteDataSource": false, "bVisible": false},
        {"name": "DateProcessed", "type": "DateTime", "customType": nil, "label": "Date Processed",
            "bRequired": false, "bRemoteDataSource": false, "bVisible": false},
        {"name": "AddDate", "type": "DateTime", "customType": nil, "label": "Add Date",
            "bRequired": false, "bRemoteDataSource": false, "bVisible": false},
        {"name": "AddedBy", "type": "String", "customType": nil, "label": "Added By",
            "bRequired": false, "bRemoteDataSource": false, "bVisible": false},
        {"name": "Approvers", "type": "Custom", "customType": "Approvers", "label": "Approvers",
            "bRequired": false, "bRemoteDataSource": false, "bVisible": false},
    }
}

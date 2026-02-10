package approvals

// Query + selections base types
type QueryRecordList struct {
    Filters       string
    Search        string
    SortBy        string
    CustomColumns string
    PageNumber    int
    PageSize      int
    BClearFilters bool
    BClearSearch  bool
    BClearSortBy  bool
    BResetColumns bool
}

type ListSelections struct {
    FilterString         string
    SortString           string
    FullTextSearchString string
    CustomColumnsString  string
    PageNumber           int
    PageSize             int
}

type ListSPConfig struct {
    ListKey      string
    TrackingID   string
    ColumnMap    map[string]string
    SearchFields []string
}

// Approval specific DTOs
type ProcessApprovalRequest struct {
    ApprovalsId int  `json:"approvalsId"`
    BApprove    bool `json:"bApprove"`
}

type DbResultRow struct {
    Id      int         `json:"id"`
    Status  string      `json:"status"`
    Details interface{} `json:"details"`
    Errors  []string    `json:"errors"`
}

type FormMetadataField struct {
    Name              string      `json:"name"`
    Type              string      `json:"type"`
    CustomType        interface{} `json:"customType"`
    Label             string      `json:"label"`
    BRequired         bool        `json:"bRequired"`
    BRemoteDataSource bool        `json:"bRemoteDataSource"`
    DataSource        interface{} `json:"dataSource"`
    Header            interface{} `json:"header"`
    OrderNumber       int         `json:"orderNumber"`
    BVisible          bool        `json:"bVisible"`
    BSortable         bool        `json:"bSortable"`
    Editable          bool        `json:"editable"`
    BFilterable       bool        `json:"bFilterable"`
}

package admin

// .NET: QueryRecordListDto ka equivalent
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

// .NET: SiteUsersListSelections row ka equivalent (GetSiteUsersListSelections SP)
type ListSelections struct {
	FilterString         string
	SortString           string
	FullTextSearchString string
	CustomColumnsString  string
	PageNumber           int
	PageSize             int
}

// List endpoints ke liye config (har list endpoint me different hoga)
type ListSPConfig struct {
	ListKey      string
	TrackingID   string
	ColumnMap    map[string]string // whitelist
	SearchFields []string          // whitelist
}

type DropDownItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
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

type AccessRightNode struct {
	Id            int               `json:"id"`
	DisplayName   string            `json:"displayName,omitempty"`
	Path          string            `json:"path,omitempty"`
	BHasAccess    bool              `json:"bHasAccess"`
	ChildElements []AccessRightNode `json:"childElements"`
}

type AdminRoleCreateRequest struct {
	Id                  int                 `json:"id"`
	Name                string              `json:"name"`
	BSuppressed         bool                `json:"bSuppressed"`
	Level               string              `json:"level"`
	ListAdminRoleLevels []map[string]string `json:"listAdminRoleLevels,omitempty"`
	AccessRights        []AccessRightNode   `json:"accessRights"`
}

type AdminRoleEditRequest struct {
	Id           int               `json:"id"`
	Name         string            `json:"name"`
	BSuppressed  bool              `json:"bSuppressed"`
	Level        string            `json:"level"`
	AccessRights []AccessRightNode `json:"accessRights"`
}

// SP DbResultDto style row
type DbResultRow struct {
	Id      int         `json:"id"`
	Status  string      `json:"status"`
	Details interface{} `json:"details"`
	Errors  []string    `json:"errors"`
}

type FormMeta struct {
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
type AdminRoleDeleteRequest struct {
	Ids []int `json:"idsToDelete"` // frontend payload: {"ids":[1030,1031,1032]}
}

type DeleteDetails struct {
	SuccessfulDeletions []int `json:"successfulDeletions"`
	FailedDeletions     []int `json:"failedDeletions"`
}

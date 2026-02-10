package dashboard

import "time"

type QueryRecordListDto struct {
	PageNumber int
	PageSize   int
	Filters    map[string]string
	SortBy     string
	Search     string
}

type SiteUsersListSelectionsDto struct {
	Filters map[string]string
	SortBy  string
	Search  string
}

type AdminUserListDto struct {
	Users []AdminUserListDataRow
	Total int
}

type AdminUserListDataRow struct {
	ID       int
	Username string
	Email    string
	// add other fields
}
type AdminUser struct {
	ID         int       `json:"adminUsers__Id"`
	Code       string    `json:"adminUsers__AdminUsersCode"`
	Email      string    `json:"siteUsers__EmailAddress"`
	FirstName  string    `json:"adminUsers__FirstName"`
	LastName   string    `json:"adminUsers__LastName"`
	Suppressed bool      `json:"siteUsers__bSuppressed"`
	AddDate    time.Time `json:"adminUsers__AddDate"`
}

type ColumnMetadata struct {
	ColumnKey      string      `json:"columnKey"`
	LabelKey       string      `json:"labelKey"`
	LabelValue     string      `json:"labelValue"`
	OrderNumber    int         `json:"orderNumber"`
	BSortable      bool        `json:"bSortable"`
	BFilterable    bool        `json:"bFilterable"`
	BVisible       bool        `json:"bVisible"`
	BLocked        bool        `json:"bLocked"`
	Type           string      `json:"type"`
	FilterMetadata interface{} `json:"filterMetadata"`
	Tooltip        *string     `json:"tooltip"`
}

type AdminUsersListDetails struct {
	ListData        []AdminUser            `json:"listData"`
	SummaryRows     []interface{}          `json:"summaryRows"`
	Columns         []ColumnMetadata       `json:"columns"`
	PageNumber      int                    `json:"pageNumber"`
	PageSize        int                    `json:"pageSize"`
	Filters         interface{}            `json:"filters"`
	SortBy          interface{}            `json:"sortBy"`
	SearchString    interface{}            `json:"searchString"`
	BHasSearchField bool                   `json:"bHasSearchField"`
	CustomColumns   interface{}            `json:"customColumns"`
	ResultsCount    int                    `json:"resultsCount"`
	Errors          []string               `json:"errors"`
	Metadata        map[string]interface{} `json:"metadata"`
}

type AdminUsersListResponse struct {
	ID      int                   `json:"id"`
	Details AdminUsersListDetails `json:"details"`
	Status  string                `json:"status"`
	Errors  []string              `json:"errors"`
}

type listProductsAccountRaw struct {
	OperationalAssetAccountsID interface{} `json:"OperationalAssetAccountsID"`
	AccountName                interface{} `json:"AccountName"`
	AssetCode                  interface{} `json:"AssetCode"`
	Balance                    interface{} `json:"Balance"`
	TopUpThreshold             interface{} `json:"TopUpThreshold"`
}

type listProductsProductRaw struct {
	ProductId   interface{}              `json:"ProductId"`
	ProductName interface{}              `json:"ProductName"`
	Accounts    []listProductsAccountRaw `json:"Accounts"`
}

type listProductsAccountOut struct {
	OperationalAssetAccountsID interface{} `json:"operationalAssetAccountsID"`
	AccountName                interface{} `json:"accountName"`
	AssetCode                  interface{} `json:"assetCode"`
	Balance                    interface{} `json:"balance"`
	TopUpThreshold             interface{} `json:"topUpThreshold"`
}

type listProductsProductOut struct {
	ProductId   interface{}              `json:"productId"`
	ProductName interface{}              `json:"productName"`
	Accounts    []listProductsAccountOut `json:"accounts"`
}

type listProductsResponse struct {
	Id      interface{}              `json:"id"`
	Details []listProductsProductOut `json:"details"`
	Status  interface{}              `json:"status"`
	Errors  []interface{}            `json:"errors"`
}

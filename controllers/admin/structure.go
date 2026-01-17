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

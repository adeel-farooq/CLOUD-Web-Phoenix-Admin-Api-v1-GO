// structure.go
package business

// ColumnMap: front filters/sort keys -> SQL columns
func OperationalAssetAccountsColumnMap() map[string]string {
	return map[string]string{
		"OperationalAssetAccounts__Id":      "OperationalAssetAccounts.Id",
		"OperationalAssetAccounts__Code":    "OperationalAssetAccounts.OperationalAssetAccountsCode",
		"OperationalAssetAccounts__Name":    "OperationalAssetAccounts.Name",
		"OperationalAssetAccounts__AddDate": "OperationalAssetAccounts.AddDate",
		"OperationalAssetAccounts__bActive": "OperationalAssetAccounts.bActive",

		"Assets__Id":     "Assets.Id",
		"Assets__Code":   "Assets.Code",
		"Assets__Name":   "Assets.Name",
		"Assets__Symbol": "Assets.Symbol",

		"Products__ProductName": "Products.ProductName",

		// add more if your SP returns these
		"Licensees__LicenseeName": "Licensees.LicenseeName",
	}
}

// Search fields used to build SearchString (LIKE %term%)
func OperationalAssetAccountsSearchFields() []string {
	return []string{
		"OperationalAssetAccounts.OperationalAssetAccountsCode",
		"OperationalAssetAccounts.Name",
		"Assets.Code",
		"Assets.Name",
		"Products.ProductName",
		"Licensees.LicenseeName",
	}
}

// Fixed columns (NET-like). Adjust keys/labels to match your UI requirements.
// IMPORTANT: bFilterable=false -> filterMetadata must be nil (handled by NormalizeColumnsNetLike too)
func OperationalAssetAccountsColumns() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"columnKey":      "OperationalAssetAccounts__Id",
			"labelKey":       "Id",
			"labelValue":     "Id",
			"orderNumber":    1,
			"bSortable":      true,
			"bFilterable":    false,
			"bVisible":       false,
			"bLocked":        false,
			"type":           "Integer",
			"filterMetadata": nil,
			"tooltip":        nil,
		},
		{
			"columnKey":   "OperationalAssetAccounts__Code",
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
			"columnKey":   "Products__ProductName",
			"labelKey":    "Product",
			"labelValue":  "Product",
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
			"columnKey":      "OperationalAssetAccounts__AddDate",
			"labelKey":       "Add Date",
			"labelValue":     "Add Date",
			"orderNumber":    5,
			"bSortable":      true,
			"bFilterable":    false,
			"bVisible":       true,
			"bLocked":        false,
			"type":           "DateTime",
			"filterMetadata": nil,
			"tooltip":        nil,
		},
	}
}

func CustomerAssetAccountsColumnMap() map[string]string {
	return map[string]string{
		"CustomerAssetAccounts__Id":                        "CustomerAssetAccounts.Id",
		"CustomerAssetAccounts__CustomerAssetAccountsCode": "CustomerAssetAccounts.CustomerAssetAccountsCode",
		"CustomerAssetAccounts__AccountNumber":             "CustomerAssetAccounts.AccountNumber",
		"CustomerAssetAccounts__AddDate":                   "CustomerAssetAccounts.AddDate",
		"CustomerAssetAccounts__Status":                    "CustomerAssetAccounts.Status",

		"Customers__Id":            "Customers.Id",
		"Customers__CustomersCode": "Customers.CustomersCode",

		"CustomerDetails__CustomerName": "CustomerDetails.CustomerName",

		"Assets__Code":   "Assets.Code",
		"Assets__Name":   "Assets.Name",
		"Assets__Symbol": "Assets.Symbol",

		"Products__ProductName":   "Products.ProductName",
		"Licensees__LicenseeName": "Licensees.LicenseeName",
	}
}

func CustomerAssetAccountsSearchFields() []string {
	return []string{
		"CustomerAssetAccounts.CustomerAssetAccountsCode",
		"CustomerAssetAccounts.AccountNumber",
		"Customers.CustomersCode",
		"CustomerDetails.CustomerName",
		"Assets.Code",
		"Assets.Name",
		"Products.ProductName",
		"Licensees.LicenseeName",
	}
}

// Fixed columns: .NET parity (bFilterable false => filterMetadata nil)
func CustomerAssetAccountsColumns() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"columnKey":      "CustomerAssetAccounts__Id",
			"labelKey":       "Id",
			"labelValue":     "Id",
			"orderNumber":    1,
			"bSortable":      true,
			"bFilterable":    false,
			"bVisible":       false,
			"bLocked":        false,
			"type":           "Integer",
			"filterMetadata": nil,
			"tooltip":        nil,
		},
		{
			"columnKey":      "CustomerAssetAccounts__CustomerAssetAccountsCode",
			"labelKey":       "Code",
			"labelValue":     "Code",
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
			"columnKey":      "Customers__CustomersCode",
			"labelKey":       "Customer",
			"labelValue":     "Customer",
			"orderNumber":    3,
			"bSortable":      true,
			"bFilterable":    true,
			"bVisible":       true,
			"bLocked":        false,
			"type":           "String",
			"filterMetadata": map[string]interface{}{"filterType": "TextContains", "details": nil},
			"tooltip":        nil,
		},
		{
			"columnKey":      "CustomerDetails__CustomerName",
			"labelKey":       "CustomerName",
			"labelValue":     "Customer Name",
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
			"columnKey":      "Assets__Code",
			"labelKey":       "Asset",
			"labelValue":     "Asset",
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
			"columnKey":      "Products__ProductName",
			"labelKey":       "Product",
			"labelValue":     "Product",
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
			"columnKey":      "CustomerAssetAccounts__AddDate",
			"labelKey":       "Add Date",
			"labelValue":     "Add Date",
			"orderNumber":    7,
			"bSortable":      true,
			"bFilterable":    false,
			"bVisible":       true,
			"bLocked":        false,
			"type":           "DateTime",
			"filterMetadata": nil,
			"tooltip":        nil,
		},
	}
}

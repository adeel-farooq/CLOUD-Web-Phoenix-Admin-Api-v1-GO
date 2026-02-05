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

// Business module (v1_AdminRole_BusinessModule_List)
func BusinessColumnMap() map[string]string {
	return map[string]string{
		"Customers__Id":                                    "Customers.Id",
		"Customers__CustomersCode":                         "Customers.CustomersCode",
		"LicenseesBrands__StatementDescriptor":             "LicenseesBrands.StatementDescriptor",
		"Customers__CompanyName":                           "Customers.CompanyName",
		"Customers__CompanyEmailAddress":                   "Customers.CompanyEmailAddress",
		"Customers__bBusinessDocumentsVerified":            "Customers.bBusinessDocumentsVerified",
		"Customers__bFinancialInstitution":                 "Customers.bFinancialInstitution",
		"Customers__bAllDocumentsSubmitted":                "Customers.bAllDocumentsSubmitted",
		"Customers__bVirtual":                              "Customers.bVirtual",
		"Customers__BusinessVerificationStatus":            "Customers.BusinessVerificationStatus",
		"Customers__bSubmittedForm":                        "Customers.bSubmittedForm",
		"Customers__bFrozen":                               "Customers.bFrozen",
		"Customers__AddDate":                               "Customers.AddDate",
		"CustomerUsers__Id":                                "CustomerUsers.Id",
		"CustomersAvailableAccounts__bNewAccountAvailable": "CustomersAvailableAccounts.bNewAccountAvailable",
	}
}

func BusinessSearchFields() []string {
	return []string{
		"Customers.CustomersCode",
		"LicenseesBrands.StatementDescriptor",
		"Customers.CompanyName",
		"Customers.CompanyEmailAddress",
		"Customers.BusinessVerificationStatus",
	}
}

// Customer users module (v1_AdminRole_CustomersModule_List)
func CustomerUsersColumnMap() map[string]string {
	return map[string]string{
		"CustomerUsers__Id":                                    "CustomerUsers.Id",
		"CustomerUsersCustomers__Id":                           "CustomerUsersCustomers.Id",
		"Customers__Id":                                        "Customers.Id",
		"CustomerUsers__CustomerUsersCode":                     "CustomerUsers.CustomerUsersCode",
		"Licensees__LicenseeName":                              "Licensees.LicenseeName",
		"LicenseesBrands__SiteName":                            "LicenseesBrands.SiteName",
		"Customers__AccountType":                               "Customers.AccountType",
		"Customers__bVirtual":                                  "Customers.bVirtual",
		"Customers__CompanyName":                               "Customers.CompanyName",
		"Customers__CompanyEmailAddress":                       "Customers.CompanyEmailAddress",
		"CustomerUsers__FirstName":                             "CustomerUsers.FirstName",
		"CustomerUsers__LastName":                              "CustomerUsers.LastName",
		"SiteUsers__EmailAddress":                              "SiteUsers.EmailAddress",
		"CustomerUsers__DateOfBirth":                           "CustomerUsers.DateOfBirth",
		"CustomerUsers__bDocumentVerified":                     "CustomerUsers.bDocumentVerified",
		"CustomerUsers__bRequiresManualVerification":           "CustomerUsers.bRequiresManualVerification",
		"CustomerUsers__bEligibleForManualVerification":        "CustomerUsers.bEligibleForManualVerification",
		"CustomerUsers__VerificationDate":                      "CustomerUsers.VerificationDate",
		"CustomerUsers__VerificationStatus":                    "CustomerUsers.VerificationStatus",
		"SiteUsers__bSuppressed":                               "SiteUsers.bSuppressed",
		"Customers__bFrozen":                                   "Customers.bFrozen",
		"CustomerUsers__AddDate":                               "CustomerUsers.AddDate",
		"CustomerUsers__bApiConfigurable":                      "CustomerUsers.bApiConfigurable",
		"CustomerUsersAvailableAccounts__bNewAccountAvailable": "CustomerUsersAvailableAccounts.bNewAccountAvailable",
	}
}

func CustomerUsersSearchFields() []string {
	return []string{
		"CustomerUsers.CustomerUsersCode",
		"Customers.CompanyName",
		"CustomerUsers.FirstName",
		"CustomerUsers.LastName",
		"SiteUsers.EmailAddress",
		"Licensees.LicenseeName",
		"LicenseesBrands.SiteName",
		"Customers.CustomersCode",
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




// Product Listings module (v1_AdminRole_CustomersModule_GetAllProducts)
func ProductListingsColumnMap() map[string]string {
    return map[string]string{
        "sourceProductsID":    "SourceProducts.Id",
        "sourceProductName":   "SourceProducts.ProductName",
        "sourceAssetsID":      "SourceAssets.Id",
        "sourceAssetName":     "SourceAssets.AssetName",
        "countriesID":         "Countries.Id",
        "countryName":         "Countries.CountryName",
        "payeeAssetsID":       "PayeeAssets.Id",
        "payeeAssetName":      "PayeeAssets.AssetName",
        "transferProductsID":  "TransferProducts.Id",
        "transferProductName": "TransferProducts.ProductName",
        "bPersonalPayment":    "ProductListings.bPersonalPayment",
        "bCompanyPayment":     "ProductListings.bCompanyPayment",
    }
}

func ProductListingsSearchFields() []string {
    return []string{
        "SourceProducts.ProductName",
        "SourceAssets.AssetName",
        "Countries.CountryName",
        "PayeeAssets.AssetName",
        "TransferProducts.ProductName",
    }
}



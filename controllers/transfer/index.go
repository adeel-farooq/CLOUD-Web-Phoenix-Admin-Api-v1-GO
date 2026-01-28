// index.go
package transfer

import (
	"cloud-web-phoenix-customer-v1-go/controllers/admin"
	"cloud-web-phoenix-customer-v1-go/controllers/auth"

	"cloud-web-phoenix-customer-v1-go/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetOutboundTransferList(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	siteUsersId := user["id"].(int)

	// query params (page, size, sort, filter)
	q := admin.ParseQueryRecordList(c.Request.URL.Query())

	// Load saved selections (same as .NET)
	ex := admin.LoadListSelections(siteUsersId, "Admin_OutboundTransfers")
	admin.OverrideWithSelections(&q, ex)

	// CALL LIST SP
	spName := "v1_AdminRole_TransfersModule_GetOutboundTransferList"

	// This stored procedure appears to build its own dynamic SQL from Raw* inputs.
	// Passing generated SQL fragments via SearchString/Filters has been causing syntax errors.
	spParams := BuildOutboundTransferParamsNetLike(q, siteUsersId, admin.ListSPConfig{
		ListKey:    "Admin_OutboundTransfers",
		TrackingID: "DefaultTrackingID",
		ColumnMap: map[string]string{
			"operationalAssetAccountsExternalTransfers__OperationalAssetAccountsExternalTransfersCode": "OperationalAssetAccountsExternalTransfers.OperationalAssetAccountsExternalTransfersCode",
			"operationalAssetAccountsExternalTransfers__Description":                                   "OperationalAssetAccountsExternalTransfers.Description",
			"operationalAssetAccountsExternalTransfers__Reference":                                     "OperationalAssetAccountsExternalTransfers.Reference",
			"operationalAssetAccountsExternalTransfers__Amount":                                        "OperationalAssetAccountsExternalTransfers.Amount",
			"operationalAssetAccountsExternalTransfers__AddDate":                                       "OperationalAssetAccountsExternalTransfers.AddDate",
			"operationalAssetAccountsExternalTransfers__CompletionDate":                                "OperationalAssetAccountsExternalTransfers.CompletionDate",
			"assets__Code":           "Assets.Code",
			"transferStatus__Status": "TransferStatus.Status",
			"products__ProductName":  "Products.ProductName",
		},
		SearchFields: []string{
			"OperationalAssetAccountsExternalTransfers.OperationalAssetAccountsExternalTransfersCode",
			"OperationalAssetAccountsExternalTransfers.Description",
			"OperationalAssetAccountsExternalTransfers.Reference",
			"Assets.Code",
			"Assets.Name",
			// "TransferStatus.Status",
			"Products.ProductName",
		},
	})

	res, err := auth.ExecSP(db.DB, spName, spParams, 2)

	// ✅ .NET style: list endpoint pe error -> 200 with details.errors
	if err != nil {
		columns := FixColumnsNetLike(OutboundTransfersColumns())
		c.JSON(200, gin.H{
			"id":     siteUsersId,
			"status": "1",
			"errors": []string{},
			"details": BuildListDetailsNetLike(
				columns,
				[]map[string]interface{}{},
				q,
				0,
				[]map[string]string{{"message": err.Error()}},
			),
		})
		return
	}

	rows := admin.AsRows(res)

	total := 0
	if len(rows) > 0 {
		if v, ok := rows[0]["HowManyResults"]; ok {
			total = toInt(v)
		}
	}

	listData := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		listData = append(listData, NormalizeOutboundRowNetLike(row))
	}

	columns := FixColumnsNetLike(OutboundTransfersColumns())

	c.JSON(200, gin.H{
		"id":     siteUsersId,
		"status": "1",
		"errors": []string{},
		"details": BuildListDetailsNetLike(
			columns,
			listData,
			q,
			total,
			[]map[string]string{},
		),
	})
}

func GetUnAppliedTransfersList(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	siteUsersId := user["id"].(int)

	// Query params (page/size/sort/filter/search + clear flags)
	q := admin.ParseQueryRecordList(c.Request.URL.Query())

	// .NET flow: selections load + override
	// SP default listkey: 'Admin_UnAppliedTransfers'
	ex := admin.LoadListSelections(siteUsersId, "Admin_UnAppliedTransfers")
	admin.OverrideWithSelections(&q, ex)

	spName := "v1_AdminRole_UnAppliedTransfersModule_GetUnAppliedTransferList"

	cfg := admin.ListSPConfig{
		ListKey:    "Admin_UnAppliedTransfers",
		TrackingID: "DefaultTrackingID",
	}

	// IMPORTANT: optional fragments -> NULL, Raw* -> can be ""
	spParams := BuildListSPParamsUnAppliedTransfersNetLike(q, siteUsersId, cfg)

	res, err := auth.ExecSP(db.DB, spName, spParams, 2)

	// .NET style: list endpoint pe error -> 200 + details.errors
	if err != nil {
		c.JSON(200, gin.H{
			"id":     siteUsersId,
			"status": "1",
			"errors": []string{},
			"details": BuildListDetailsNetLike(
				UnAppliedTransfersColumns(),
				[]map[string]interface{}{},
				q,
				0,
				[]map[string]string{{"message": err.Error()}},
			),
		})
		return
	}

	rows := admin.AsRows(res)

	total := 0
	if len(rows) > 0 {
		if v, ok := rows[0]["HowManyResults"]; ok {
			total = toInt(v)
		}
	}

	listData := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		listData = append(listData, NormalizeOutboundRowNetLike(row))
	}

	c.JSON(200, gin.H{
		"id":     siteUsersId,
		"status": "1",
		"errors": []string{},
		"details": BuildListDetailsNetLike(
			UnAppliedTransfersColumns(),
			listData,
			q,
			total,
			[]map[string]string{},
		),
	})

}

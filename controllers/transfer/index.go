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

	// 🔹 Load saved selections (same as .NET)
	ex := admin.LoadListSelections(siteUsersId, "OutboundTransfers")
	admin.OverrideWithSelections(&q, ex)

	// 🔹 Endpoint config
	cfg := admin.ListSPConfig{
		ListKey:    "OutboundTransfers",
		TrackingID: "DefaultTrackingID",
		ColumnMap: map[string]string{
			"transfers__Id":           "Transfers.Id",
			"transfers__Reference":    "Transfers.Reference",
			"transfers__Amount":       "Transfers.Amount",
			"transfers__Currency":     "Transfers.Currency",
			"transfers__Status":       "Transfers.Status",
			"transfers__CreatedDate":  "Transfers.CreatedDate",
			"customers__EmailAddress": "Customers.EmailAddress",
		},
		SearchFields: []string{
			"Transfers.Reference",
			"Customers.EmailAddress",
		},
	}

	// 🔹 Build SP params
	spParams := admin.BuildListSPParams(q, siteUsersId, cfg)

	// 🔹 CALL LIST SP
	spName := "v1_TransfersModule_OutboundTransfer_List"
	res, err := auth.ExecSP(db.DB, spName, spParams, 2)
	if err != nil {
		c.JSON(500, gin.H{"status": "0", "error": err.Error()})
		return
	}

	rows := admin.AsRows(res)

	listData := []map[string]interface{}{}
	total := 0

	for _, row := range rows {
		item := admin.NormalizeRowKeys(row)
		listData = append(listData, item)

		if v, ok := row["HowManyResults"]; ok {
			total = toInt(v)
		}
	}

	columns := []map[string]interface{}{}
	if len(rows) > 0 {
		columns = admin.BuildColumnsFromSPRow(rows[0], 1)
	}

	c.JSON(200, gin.H{
		"status": "1",
		"details": admin.BuildListDetails(
			columns,
			listData,
			q,
			total,
		),
		"errors": []string{},
	})
}

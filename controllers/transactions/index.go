package transactions

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cloud-web-phoenix-customer-v1-go/controllers/admin"
	"cloud-web-phoenix-customer-v1-go/controllers/auth"
	"cloud-web-phoenix-customer-v1-go/db"

	"github.com/gin-gonic/gin"
)

// List handles GET /transactions/list
// Uses generic handler for ~90% reduction in code
func List(c *gin.Context) {
	HandleList(c, ListConfigTransactions())
}

// ListAll handles GET /transactions/listall
func ListAll(c *gin.Context) {
	HandleList(c, ListConfigTransactionsAll())
}

// DownloadAll handles POST /transactionsmodule/downloadall
// Returns CSV file of all transactions
func DownloadAll(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)

	// Parse query params
	pageNumber := 1
	pageSize := 10
	ignorePagination := false

	if pn := c.Query("pageNumber"); pn != "" {
		if v, err := strconv.Atoi(pn); err == nil && v > 0 {
			pageNumber = v
		}
	}
	if ps := c.Query("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = v
		}
	}
	if ip := c.Query("ignorePagination"); ip == "true" {
		ignorePagination = true
		pageNumber = 1
		pageSize = 100000 // Max rows for Excel
	}

	filters := strings.ReplaceAll(c.Query("filters"), "+", " ")
	filters = strings.ReplaceAll(filters, ",", "TO")
	sortBy := c.Query("sortBy")
	search := c.Query("search")

	// Call SP to get transaction list
	res, err := auth.ExecSP(db.DB, "v1_AdminRole_TransactionsModule_GetAllTransactionsList",
		map[string]interface{}{
			"User_SiteUsersID": siteUsersId,
			"PageNumber":       pageNumber,
			"PageSize":         pageSize,
			"ListKey":          "Admin_DownloadAllTransactions",
			"TrackingID":       "DefaultTrackingID",
			"RawFilterString":  filters,
			"RawSortString":    sortBy,
			"RawSearchString":  search,
		}, 0) // 0 = return all rows
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	rows, ok := res.([]map[string]interface{})
	if !ok || len(rows) == 0 {
		rows = []map[string]interface{}{}
	}

	// Generate CSV
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write custom headers (metadata)
	writer.Write([]string{"Filters string:", filters})
	writer.Write([]string{"Sort string:", sortBy})
	writer.Write([]string{"Search string:", search})
	writer.Write([]string{"Page number:", strconv.Itoa(pageNumber)})
	writer.Write([]string{"Page size:", strconv.Itoa(pageSize)})
	writer.Write([]string{"Ignore Pagination:", strconv.FormatBool(ignorePagination)})
	writer.Write([]string{}) // Empty row separator

	// Get column headers from first row
	if len(rows) > 0 {
		headers := make([]string, 0)
		for key := range rows[0] {
			if key != "HowManyResults" && key != "RowNum" {
				headers = append(headers, key)
			}
		}
		writer.Write(headers)

		// Write data rows
		for _, row := range rows {
			values := make([]string, len(headers))
			for i, header := range headers {
				if v, ok := row[header]; ok && v != nil {
					values[i] = fmt.Sprint(v)
				}
			}
			writer.Write(values)
		}
	}

	writer.Flush()

	// Set response headers for CSV download
	filename := fmt.Sprintf("CustomerAssetAccountsTransactions_%s.csv", time.Now().Format("2006-01-02_15-04-05"))
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "text/csv", buf.Bytes())
}

// ListPending handles GET /transactions/listpending
func ListPending(c *gin.Context) {
	// Check for bFilterToOwnAssignments query param
	bFilterToOwnAssignments := c.Query("bFilterToOwnAssignments") == "true"

	cfg := ListConfigPendingTransactions(bFilterToOwnAssignments)

	// Custom response builder to include filters in response
	cfg.CustomResponseBuilder = func(ctx *gin.Context, listData []map[string]interface{}, total int, q admin.QueryRecordList, siteUsersId int) {
		columns := ColumnsPendingTransactionsListRefactored()
		details := admin.BuildListDetails(columns, listData, q, total)

		// Include raw filters string in response (matches .NET behavior)
		if raw := strings.TrimSpace(ctx.Query("filters")); raw != "" {
			details["filters"] = raw
		}

		ctx.JSON(http.StatusOK, gin.H{
			"id":      siteUsersId,
			"details": details,
			"status":  "1",
			"errors":  []interface{}{},
		})
	}

	HandleList(c, cfg)
}

// ListPendingTreasury handles GET /transactions/listpendingtreasury
func ListPendingTreasury(c *gin.Context) {
	HandleList(c, ListConfigPendingTransactionsTreasury())
}

// ListFrozen handles GET /transactions/listfrozen
func ListFrozen(c *gin.Context) {
	HandleList(c, ListConfigFrozenTransactions())
}

func TransactionDetails(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"id":      0,
			"details": nil,
			"status":  "0",
			"errors": []gin.H{
				{"fieldName": "Authorization", "messageCode": "Unauthorized"},
			},
		})
		return
	}

	// .NET expects: ?transactionsId=123
	transactionsIdStr := c.Query("transactionsId")
	if transactionsIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"id":      0,
			"details": nil,
			"status":  "0",
			"errors": []gin.H{
				{"fieldName": "transactionsId", "messageCode": "Required"},
			},
		})
		return
	}

	transactionsId, err := strconv.Atoi(transactionsIdStr)
	if err != nil || transactionsId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"id":      0,
			"details": nil,
			"status":  "0",
			"errors": []gin.H{
				{"fieldName": "transactionsId", "messageCode": "Invalid"},
			},
		})
		return
	}

	// Call the same SP as .NET DB client:
	// v2_AdminRole_TransactionsModule_GetTransactionDetails
	res, err := auth.ExecSP(
		db.DB,
		"v2_AdminRole_TransactionsModule_GetTransactionDetails",
		map[string]interface{}{
			"SiteUsersId":                         user["id"],
			"CustomerAssetAccountsTransactionsId": transactionsId,
		},
		1,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"id":      0,
			"details": nil,
			"status":  "0",
			"errors": []gin.H{
				{"fieldName": "transactionsId", "messageCode": "SP_Error"},
			},
		})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"id":      0,
			"details": nil,
			"status":  "0",
			"errors": []gin.H{
				{"fieldName": "", "messageCode": "Invalid_SP_Response"},
			},
		})
		return
	}

	// Details is JSON string in SQL result (same pattern as your other endpoints)
	var details map[string]interface{}
	if detailsStr, ok := row["Details"].(string); ok && detailsStr != "" {
		_ = json.Unmarshal([]byte(detailsStr), &details)
	}
	if details != nil {
		if normalized, ok := normalizeJSONKeysLowerCamel(details).(map[string]interface{}); ok {
			details = normalized
		}
		FormatMoneyFields2dp(details)
	}

	// .NET returns:
	// { "Id": <int>, "Details": <object>, "Metadata": <array>, "Status": "1|0", "Errors": [...] }
	// (But your project generally uses lowercase keys. You can switch keys if your frontend expects lowercase.)
	c.JSON(http.StatusOK, gin.H{
		"id":       row["Id"],
		"details":  details,
		"metadata": TransactionDetailsMetadata,
		"status":   row["Status"],
		"errors":   []interface{}{},
	})
}

func ListProducts(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	res, err := auth.ExecSP(
		db.DB,
		"v1_AdminRole_PendingTransactionsModule_ListProducts",
		map[string]interface{}{
			"SiteUsersId": user["id"],
		},
		1,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid SP response"})
		return
	}

	// Details is JSON string in DBResultDto style
	var raw []pendingProductsRaw
	if detailsStr, ok := row["Details"].(string); ok && strings.TrimSpace(detailsStr) != "" {
		// some SPs return single-quoted JSON, same cleanup used in dashboard module
		cleanJSON := strings.ReplaceAll(detailsStr, "'", "\"")
		_ = json.Unmarshal([]byte(cleanJSON), &raw)
	}

	out := make([]pendingProductsOut, 0, len(raw))
	for _, r := range raw {
		out = append(out, pendingProductsOut{
			ProductId:               r.ProductId,
			ProductName:             r.ProductName,
			Asset:                   r.Asset,
			DisplayName:             r.DisplayName,
			Amount:                  r.Amount,
			PendingTransactionCount: r.PendingTransactionCount,
		})
	}

	c.JSON(http.StatusOK, pendingProductsResponse{
		Id:      row["Id"],
		Details: out,
		Status:  row["Status"],
		Errors:  []interface{}{},
	})
}

func ListPendingTreasuryProducts(c *gin.Context) {
	// .NET response shows id: 0 always
	const responseId = 0

	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	siteUsersId := user["id"].(int)

	// ---------------------------
	// CHANGE THIS IF YOUR DB USES DIFFERENT SP NAME
	// ---------------------------
	const spName = "v1_AdminRole_PendingTransactionsTreasuryModule_ListProducts"
	// ---------------------------

	// ✅ FIX: SP expects @SiteUsersId
	params := map[string]interface{}{
		"SiteUsersId": siteUsersId,
	}

	res, err := auth.ExecSP(db.DB, spName, params, 1)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			c.JSON(http.StatusOK, pendingProductsResponse{
				Id:      responseId,
				Details: []pendingProductsOut{},
				Status:  "1",
				Errors:  []interface{}{},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "0",
			"error":  err.Error(),
		})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid SP response"})
		return
	}

	// Details is JSON string in DBResultDto style (same pattern as ListProducts)
	detailsStr := ""
	switch t := row["Details"].(type) {
	case string:
		detailsStr = t
	case []byte:
		detailsStr = string(t)
	}

	var raw []pendingProductsRaw
	if strings.TrimSpace(detailsStr) != "" {
		// some SPs return single-quoted JSON
		cleanJSON := strings.ReplaceAll(detailsStr, "'", "\"")
		_ = json.Unmarshal([]byte(cleanJSON), &raw)
	}

	out := make([]pendingProductsOut, 0, len(raw))
	for _, r := range raw {
		out = append(out, pendingProductsOut{
			ProductId:               r.ProductId,
			ProductName:             r.ProductName,
			Asset:                   r.Asset,
			DisplayName:             r.DisplayName,
			Amount:                  r.Amount,
			PendingTransactionCount: r.PendingTransactionCount,
		})
	}

	status := row["Status"]
	if status == nil {
		status = "1"
	}

	c.JSON(http.StatusOK, pendingProductsResponse{
		Id:      responseId,
		Details: out,
		Status:  status,
		Errors:  []interface{}{},
	})
}

func TransactionJSON(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	// Query params: Id + transactionsId (both appear in your URL)
	idStr := c.Query("Id")
	trxIdStr := c.Query("transactionsId")

	// Decide "requested id" for the "no result" response
	// In your sample, when no result: id = 249074 (the request value)
	requestedID := parseInt64Prefer(trxIdStr, idStr)

	if requestedID == 0 {
		// If caller didn't send a valid number, follow your existing API style
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"error":  "Missing or invalid Id/transactionsId",
		})
		return
	}

	siteUsersId := user["id"].(int)

	// ✅ SP NAME (adjust if your DB uses a slightly different name)
	// Based on your endpoint: /transactionsmodule/transaction-json

	const spName = "v1_AdminRole_TransactionsModule_GetTransactionJson"

	// ✅ Params: include both ids because your endpoint includes both.
	// Also include SiteUsersId because many AdminRole SPs require it.
	params := map[string]interface{}{
		"SiteUsersId": siteUsersId,
		// "Id":             requestedID,
		"CustomerAssetAccountsTransactionsId": requestedID,
	}

	res, err := auth.ExecSP(db.DB, spName, params, 1)
	if err != nil {
		// If SP returns no rows, your API still returns status=1 with details=null
		if err.Error() == "sql: no rows in result set" {
			c.JSON(http.StatusOK, transactionJSONResponse{
				Id:      requestedID,
				Details: nil,
				Status:  "1",
				Errors:  []interface{}{},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "0",
			"error":  err.Error(),
		})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok || row == nil {
		// Treat as "no result" to match your sample response
		c.JSON(http.StatusOK, transactionJSONResponse{
			Id:      requestedID,
			Details: nil,
			Status:  "1",
			Errors:  []interface{}{},
		})
		return
	}

	// Expected columns from SP (common patterns):
	// - "Id" or "id"
	// - "Details" or "details"
	respID := firstNonZeroInt(row, "Id", "id", "TransactionsId", "transactionsId")
	if respID == 0 {
		respID = requestedID
	}

	details := firstStringPtr(row, "Details", "details") // keep it as string JSON

	if details == nil || *details == "" {
		c.JSON(http.StatusOK, transactionJSONResponse{
			Id:      respID,
			Details: nil,
			Status:  "1",
			Errors:  []interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, transactionJSONResponse{
		Id:      respID,
		Details: *details, // IMPORTANT: details is a JSON string
		Status:  "1",
		Errors:  []interface{}{},
	})
}

func ListNotes(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)

	transactionsId, ok := mustInt64FromQueryOrForm(c.Query("transactionsId"))
	if !ok || transactionsId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"error":  "Missing or invalid transactionsId",
		})
		return
	}

	// ✅ Adjust SP name if your DB uses a different one
	const spName = "v1_AdminRole_TransactionsModule_GetNotesList"

	params := map[string]interface{}{
		"SiteUsersId":    siteUsersId,
		"TransactionsId": transactionsId,
	}

	res, err := auth.ExecSP(db.DB, spName, params, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok || row == nil {
		// no result => match your standard response
		c.JSON(http.StatusOK, gin.H{
			"id":      0,
			"details": gin.H{"notes": []interface{}{}},
			"status":  "1",
			"errors":  []interface{}{},
		})
		return
	}

	// ✅ The SP returns JSON in a column; your sample shows it's the 3rd column value.
	// We'll robustly find the FIRST string column that looks like JSON.
	jsonStr := ""
	for _, v := range row {
		s, ok := v.(string)
		if !ok {
			continue
		}
		ss := strings.TrimSpace(s)
		if strings.HasPrefix(ss, "{") || strings.HasPrefix(ss, "[") {
			jsonStr = ss
			break
		}
	}

	if jsonStr == "" {
		// If SP returned no JSON, respond empty
		c.JSON(http.StatusOK, gin.H{
			"id":      0,
			"details": gin.H{"notes": []interface{}{}},
			"status":  "1",
			"errors":  []interface{}{},
		})
		return
	}

	var env spNotesEnvelope
	if err := json.Unmarshal([]byte(jsonStr), &env); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "0",
			"error":  "Failed to parse notes JSON: " + err.Error(),
		})
		return
	}

	// Convert SP notes -> API notes (camelCase keys + id as string)
	out := make([]apiNoteItem, 0, len(env.Notes))
	for _, n := range env.Notes {
		out = append(out, apiNoteItem{
			Id:        strconv.Itoa(n.Id),
			Text:      n.Text,
			AddDate:   n.AddDate,
			AddedBy:   n.AddedBy,
			BEditable: n.BEditable,
			BPinned:   n.BPinned,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"id": 0,
		"details": gin.H{
			"notes": out,
		},
		"status": "1",
		"errors": []interface{}{},
	})

}

// helper for picking first existing key from a row
func firstNonNil(row map[string]interface{}, keys ...string) interface{} {
	for _, k := range keys {
		if v, ok := row[k]; ok && v != nil {
			return v
		}
	}
	return nil
}

func AddNote(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)

	// form-data fields
	transactionsIdStr := c.PostForm("transactionsId")
	text := c.PostForm("text")

	transactionsId, ok := mustInt64FromQueryOrForm(transactionsIdStr)
	if !ok || transactionsId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"error":  "Missing or invalid transactionsId",
		})
		return
	}
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"error":  "Missing text",
		})
		return
	}

	fullName := ""
	userCode := ""

	if v, ok := user["FullName"].(string); ok {
		fullName = v
	}
	if v, ok := user["UserCode"].(string); ok {
		userCode = v
	}

	addedBy := fullName
	if userCode != "" {
		addedBy = fullName + " (Admin - " + userCode + ")"
	}

	// ✅ Adjust SP name if your DB uses a different one
	const spName = "v1_AdminRole_TransactionsModule_AddNote"

	params := map[string]interface{}{
		"SiteUsersId":    siteUsersId,
		"TransactionsId": transactionsId,
		"Text":           text,
		"AddedBy":        addedBy,
	}

	res, err := auth.ExecSP(db.DB, spName, params, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "0",
			"error":  err.Error(),
		})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok || row == nil {
		// Treat as failure (SP should return id)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "0",
			"error":  "Invalid SP response",
		})
		return
	}

	// SP should return the new note id; key names may differ
	newId := asString(firstNonNil(row, "Id", "id", "NoteId", "noteId"))
	if newId == "" {
		newId = "0"
	}

	c.JSON(http.StatusOK, addNoteResponse{
		Id: newId,
		Details: addNoteDetails{
			TransactionsId: transactionsId,
			Text:           text,
		},
		Status: "1",
		Errors: []interface{}{},
	})
}

func EditNote(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	siteUsersId := user["id"].(int)
	// Form-data
	noteId := strings.TrimSpace(c.PostForm("id"))
	text := strings.TrimSpace(c.PostForm("text"))

	if noteId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"error":  "id is required",
		})
		return
	}
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"error":  "text is required",
		})
		return
	}

	// Optional: build EditedBy if your SP requires it
	editedBy := ""
	if v, ok := user["firstName"]; ok {
		editedBy = strings.TrimSpace(fmt.Sprint(v))
	}
	if v, ok := user["lastName"]; ok {
		ln := strings.TrimSpace(fmt.Sprint(v))
		if ln != "" {
			if editedBy != "" {
				editedBy += " "
			}
			editedBy += ln
		}
	}
	userCode := ""
	if v, ok := user["userCode"]; ok {
		userCode = strings.TrimSpace(fmt.Sprint(v))
	}
	if userCode != "" {
		if editedBy == "" {
			editedBy = fmt.Sprintf("(Admin - %s)", userCode)
		} else {
			editedBy = fmt.Sprintf("%s (Admin - %s)", editedBy, userCode)
		}
	}

	// ✅ Use the correct SP name from your DB (replace this with the exact one from .NET)
	// Example:
	spName := "v1_AdminRole_TransactionsModule_EditNote"

	params := map[string]interface{}{
		"SiteUsersId": siteUsersId,
		"NotesID":     noteId,
		"Text":        text,
		// If SP expects AddedBy, keep this:
		"EditedBy": editedBy,
	}

	res, err := auth.ExecSP(db.DB, spName, params, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "0",
			"error":  err.Error(),
		})
		return
	}

	// Many of your SPs return either:
	// 1) one row with edited note fields
	// 2) or nothing but success
	row, ok := auth.AsSingleRow(res)
	if !ok || row == nil {
		// still return expected response
		c.JSON(http.StatusOK, gin.H{
			"id": 0,
			"details": editNoteDetails{
				Id:        noteId,
				Text:      text,
				BEditable: false,
			},
			"status": "1",
			"errors": []interface{}{},
		})
		return
	}

	// bEditable might come from SP, otherwise default false
	bEditable := false
	if v, ok := row["bEditable"]; ok && v != nil {
		switch t := v.(type) {
		case bool:
			bEditable = t
		case int:
			bEditable = t != 0
		case int64:
			bEditable = t != 0
		case float64:
			bEditable = t != 0
		default:
			s := strings.TrimSpace(fmt.Sprint(t))
			bEditable = strings.EqualFold(s, "true") || s == "1"
		}
	}

	// id/text might also come from SP
	outId := noteId
	if v, ok := row["Id"]; ok && v != nil {
		outId = strings.TrimSpace(fmt.Sprint(v))
	}
	outText := text
	if v, ok := row["Text"]; ok && v != nil {
		outText = fmt.Sprint(v)
	}

	c.JSON(http.StatusOK, gin.H{
		"id": 0,
		"details": editNoteDetails{
			Id:        outId,
			Text:      outText,
			BEditable: bEditable,
		},
		"status": "1",
		"errors": []interface{}{},
	})
}

func PinNote(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	siteUsersId := user["id"].(int)

	var req pinNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"error":  "Invalid JSON body",
		})
		return
	}

	req.Id = strings.TrimSpace(req.Id)
	if req.Id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"error":  "id is required",
		})
		return
	}

	// Optional: build EditedBy if your SP requires it
	editedBy := ""
	if v, ok := user["firstName"]; ok {
		editedBy = strings.TrimSpace(fmt.Sprint(v))
	}
	if v, ok := user["lastName"]; ok {
		ln := strings.TrimSpace(fmt.Sprint(v))
		if ln != "" {
			if editedBy != "" {
				editedBy += " "
			}
			editedBy += ln
		}
	}
	userCode := ""
	if v, ok := user["userCode"]; ok {
		userCode = strings.TrimSpace(fmt.Sprint(v))
	}
	if userCode != "" {
		if editedBy == "" {
			editedBy = fmt.Sprintf("(Admin - %s)", userCode)
		} else {
			editedBy = fmt.Sprintf("%s (Admin - %s)", editedBy, userCode)
		}
	}

	// ✅ Replace with exact SP name from your .NET mapping if different
	spName := "v1_AdminRole_TransactionsModule_PinNote"

	params := map[string]interface{}{
		"SiteUsersId": siteUsersId,
		"NotesID":     req.Id,
		"bPinned":     req.BPinned,
		"EditedBy":    editedBy,
	}

	_, err := auth.ExecSP(db.DB, spName, params, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "0",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": 0,
		"details": pinNoteDetails{
			Id:      req.Id,
			BPinned: req.BPinned,
		},
		"status": "1",
		"errors": []interface{}{},
	})
}

func ViewAlerts(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	// .NET uses query param "Id"
	idStr := c.Query("Id")
	if idStr == "" {
		idStr = c.Query("id")
	}

	transactionsId := 0
	if idStr != "" {
		n, err := strconv.Atoi(idStr)
		if err != nil || n < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"id":      0,
				"details": nil,
				"status":  "0",
				"errors": []gin.H{
					{"fieldName": "Id", "messageCode": "Invalid"},
				},
			})
			return
		}
		transactionsId = n
	}

	res, err := auth.ExecSP(
		db.DB,
		"v1_AdminRole_TransactionsModule_ViewAlerts",
		map[string]interface{}{
			"TransactionsId": transactionsId,
			"SiteUsersID":    user["id"],
		},
		1,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"id":      0,
			"details": nil,
			"status":  "0",
			"errors": []gin.H{
				{"fieldName": "Id", "messageCode": "SP_Error"},
			},
		})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"id":      0,
			"details": nil,
			"status":  "0",
			"errors": []gin.H{
				{"fieldName": "", "messageCode": "Invalid_SP_Response"},
			},
		})
		return
	}

	// Details from SP is usually JSON string
	var details map[string]interface{}
	if s, ok := row["Details"].(string); ok && s != "" {
		_ = json.Unmarshal([]byte(s), &details)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       row["Id"],
		"details":  details,
		"metadata": ViewAlertsMetadata,
		"status":   row["Status"],
		"errors":   []interface{}{},
	})
}

func Release(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)
	addedBy := addedByFromToken(user)

	var req ReleaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// keep it simple; .NET validation helper would return structured errors,
		// but for invalid JSON this repo usually returns status 0 + string error.
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"errors": []string{"Invalid JSON body"},
		})
		return
	}

	// .NET validation: CustomerAssetAccountsTransactionsId required
	if req.CustomerAssetAccountsTransactionsId <= 0 {
		c.JSON(http.StatusOK, requiredField("customerAssetAccountsTransactionsId"))
		return
	}

	// .NET DB client:
	// v1_AdminRole_PendingTransactionsTreasuryModule_Release
	// @CustomerAssetAccountsTransactionsId, @SiteUsersID, @AddedBy
	spName := "v1_AdminRole_PendingTransactionsTreasuryModule_Release"
	params := map[string]interface{}{
		"CustomerAssetAccountsTransactionsId": req.CustomerAssetAccountsTransactionsId,
		"SiteUsersID":                         siteUsersId,
		"AddedBy":                             addedBy,
	}

	res, err := auth.ExecSP(db.DB, spName, params, 1) // single row
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "0",
			"error":  err.Error(),
		})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok || row == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "0",
			"error":  "Invalid SP response",
		})
		return
	}

	// If SP returns status != 1, return it as-is (same as .NET: TransformToObjectResultDto)
	status := asString(row["Status"])
	if status != "1" {
		// Some SPs return Errors as JSON / string; safest is pass-through.
		c.JSON(http.StatusOK, gin.H{
			"id":      row["Id"],
			"details": row["Details"],
			"status":  row["Status"],
			"errors":  row["Errors"],
		})
		return
	}

	// .NET success response:
	// id = (int)command.CustomerAssetAccountsTransactionsId
	// details = { customerAssetAccountsTransactionsId = (int)command.CustomerAssetAccountsTransactionsId }
	c.JSON(http.StatusOK, gin.H{
		"id": req.CustomerAssetAccountsTransactionsId,
		"details": ReleaseDetails{
			CustomerAssetAccountsTransactionsId: req.CustomerAssetAccountsTransactionsId,
		},
		"status": "1",
		"errors": []interface{}{},
	})
}

// =============================================================================
// TransactionsModule - Cancellation Endpoints
// =============================================================================

// GetCancellationEligibility handles GET /transactionsmodule/cancel
// Gets real-time information about whether or not a transaction can be cancelled
func GetCancellationEligibility(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)

	transactionsIdStr := c.Query("customerAssetAccountsTransactionsId")
	if transactionsIdStr == "" {
		c.JSON(http.StatusBadRequest, requiredField("customerAssetAccountsTransactionsId"))
		return
	}

	transactionsId, err := strconv.Atoi(transactionsIdStr)
	if err != nil || transactionsId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"errors": []FieldError{{FieldName: "customerAssetAccountsTransactionsId", MessageCode: "Invalid"}},
		})
		return
	}

	res, err := auth.ExecSP(db.DB, "v1_AdminRole_TransactionsModule_GetCancellationEligibility",
		map[string]interface{}{
			"SiteUsersId":                         siteUsersId,
			"CustomerAssetAccountsTransactionsId": transactionsId,
		}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"id":      transactionsId,
			"details": CancellationEligibilityResponse{BCanBeCancelled: false},
			"status":  "1",
			"errors":  []interface{}{},
		})
		return
	}

	// Parse response from SP
	var details map[string]interface{}
	if detailsStr, ok := row["Details"].(string); ok && detailsStr != "" {
		_ = json.Unmarshal([]byte(detailsStr), &details)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      row["Id"],
		"details": details,
		"status":  row["Status"],
		"errors":  []interface{}{},
	})
}

// CancelTransaction handles POST /transactionsmodule/cancel
// Submits a transaction cancellation request
func CancelTransaction(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)
	addedBy := addedByFromToken(user)

	var req CancelTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "0", "error": "Invalid JSON body"})
		return
	}

	if req.CustomerAssetAccountsTransactionsId <= 0 {
		c.JSON(http.StatusOK, requiredField("customerAssetAccountsTransactionsId"))
		return
	}

	res, err := auth.ExecSP(db.DB, "v1_AdminRole_TransactionsModule_CancelTransaction",
		map[string]interface{}{
			"SiteUsersId":                         siteUsersId,
			"CustomerAssetAccountsTransactionsId": req.CustomerAssetAccountsTransactionsId,
			"AddedBy":                             addedBy,
		}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": "Invalid SP response"})
		return
	}

	status := asString(row["Status"])
	if status != "1" {
		c.JSON(http.StatusOK, gin.H{
			"id":      row["Id"],
			"details": row["Details"],
			"status":  status,
			"errors":  row["Errors"],
		})
		return
	}

	// Parse cancellation response
	var details map[string]interface{}
	if detailsStr, ok := row["Details"].(string); ok && detailsStr != "" {
		_ = json.Unmarshal([]byte(detailsStr), &details)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      req.CustomerAssetAccountsTransactionsId,
		"details": details,
		"status":  "1",
		"errors":  []interface{}{},
	})
}

// =============================================================================
// TransactionsModule - WaiveFee Endpoints
// =============================================================================

// GetWaiveFee handles GET /transactionsmodule/waive-fee
func GetWaiveFee(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)

	transactionsIdStr := c.Query("transactionsId")
	if transactionsIdStr == "" {
		c.JSON(http.StatusBadRequest, requiredField("transactionsId"))
		return
	}

	transactionsId, err := strconv.Atoi(transactionsIdStr)
	if err != nil || transactionsId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"errors": []FieldError{{FieldName: "transactionsId", MessageCode: "Invalid"}},
		})
		return
	}

	res, err := auth.ExecSP(db.DB, "v1_AdminRole_TransactionsModule_GetWaiveFee",
		map[string]interface{}{
			"SiteUsersId":                         siteUsersId,
			"CustomerAssetAccountsTransactionsId": transactionsId,
		}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"id":      transactionsId,
			"details": nil,
			"status":  "1",
			"errors":  []interface{}{},
		})
		return
	}

	var details map[string]interface{}
	if detailsStr, ok := row["Details"].(string); ok && detailsStr != "" {
		_ = json.Unmarshal([]byte(detailsStr), &details)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      row["Id"],
		"details": details,
		"status":  row["Status"],
		"errors":  []interface{}{},
	})
}

// WaiveFee handles POST /transactionsmodule/waive-fee
func WaiveFee(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)
	addedBy := addedByFromToken(user)

	var req WaiveFeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "0", "error": "Invalid JSON body"})
		return
	}

	if req.CustomerAssetAccountsTransactionsId <= 0 {
		c.JSON(http.StatusOK, requiredField("customerAssetAccountsTransactionsId"))
		return
	}
	if req.Amount <= 0 {
		c.JSON(http.StatusOK, requiredField("amount"))
		return
	}

	// Calculate amountToWaive based on .NET logic
	var amountToWaive float64
	if req.BFullRefund {
		amountToWaive = req.Amount
	} else if req.PercentageAmount != nil {
		amountToWaive = req.Amount * (*req.PercentageAmount) / 100
	} else if req.FixedAmount != nil {
		amountToWaive = *req.FixedAmount
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"error":  "Invalid WaiveFeeRequest - amountToWaive not set",
		})
		return
	}

	res, err := auth.ExecSP(db.DB, "v1_AdminRole_TransactionsModule_WaiveFee",
		map[string]interface{}{
			"SiteUsersId":                         siteUsersId,
			"CustomerAssetAccountsTransactionsId": req.CustomerAssetAccountsTransactionsId,
			"AmountToWaive":                       amountToWaive,
			"AddedBy":                             addedBy,
		}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": "Invalid SP response"})
		return
	}

	status := asString(row["Status"])
	if status != "1" {
		c.JSON(http.StatusOK, gin.H{
			"id":      row["Id"],
			"details": row["Details"],
			"status":  status,
			"errors":  row["Errors"],
		})
		return
	}

	var details map[string]interface{}
	if detailsStr, ok := row["Details"].(string); ok && detailsStr != "" {
		_ = json.Unmarshal([]byte(detailsStr), &details)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      req.CustomerAssetAccountsTransactionsId,
		"details": details,
		"status":  "1",
		"errors":  []interface{}{},
	})
}

// =============================================================================
// TransactionsModule - Document Endpoints
// =============================================================================

// ListDocuments handles GET /transactionsmodule/list-documents
func ListDocuments(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)

	transactionsIdStr := c.Query("transactionsId")
	if transactionsIdStr == "" {
		transactionsIdStr = c.Query("TransactionsID")
	}

	transactionsId, err := strconv.Atoi(transactionsIdStr)
	if err != nil || transactionsId <= 0 {
		c.JSON(http.StatusBadRequest, requiredField("transactionsId"))
		return
	}

	res, err := auth.ExecSP(db.DB, "v1_AdminRole_TransactionsModule_GetDocumentsList",
		map[string]interface{}{
			"SiteUsersId":    siteUsersId,
			"TransactionsId": transactionsId,
		}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"id":      0,
			"details": DocumentListResponse{Documents: []DocumentListItem{}},
			"status":  "1",
			"errors":  []interface{}{},
		})
		return
	}

	var details map[string]interface{}
	if detailsStr, ok := row["Details"].(string); ok && detailsStr != "" {
		_ = json.Unmarshal([]byte(detailsStr), &details)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      row["Id"],
		"details": details,
		"status":  row["Status"],
		"errors":  []interface{}{},
	})
}

// AddDocument handles POST /transactionsmodule/adddocument
func AddDocument(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)
	addedBy := addedByFromToken(user)

	// Form data
	transactionsIdStr := c.PostForm("transactionsId")
	text := c.PostForm("text")

	transactionsId, _ := strconv.Atoi(transactionsIdStr)

	// Get file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, requiredField("file"))
		return
	}

	// Read file content
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": "Failed to open file"})
		return
	}
	defer f.Close()

	fileBytes := make([]byte, file.Size)
	_, err = f.Read(fileBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": "Failed to read file"})
		return
	}

	res, err := auth.ExecSP(db.DB, "v1_AdminRole_TransactionsModule_AddDocument",
		map[string]interface{}{
			"SiteUsersId":    siteUsersId,
			"TransactionsId": transactionsId,
			"Text":           text,
			"FileName":       file.Filename,
			"FileContent":    fileBytes,
			"ContentType":    file.Header.Get("Content-Type"),
			"AddedBy":        addedBy,
		}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": "Invalid SP response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      row["Id"],
		"details": row["Details"],
		"status":  row["Status"],
		"errors":  []interface{}{},
	})
}

// GetEditDocument handles GET /transactionsmodule/edit-document
func GetEditDocument(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)

	idStr := c.Query("id")
	if idStr == "" {
		idStr = c.Query("Id")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, requiredField("id"))
		return
	}

	res, err := auth.ExecSP(db.DB, "v1_AdminRole_TransactionsModule_GetEditDocument",
		map[string]interface{}{
			"SiteUsersId": siteUsersId,
			"DocumentId":  id,
		}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"id":      0,
			"details": nil,
			"status":  "1",
			"errors":  []interface{}{},
		})
		return
	}

	var details map[string]interface{}
	if detailsStr, ok := row["Details"].(string); ok && detailsStr != "" {
		_ = json.Unmarshal([]byte(detailsStr), &details)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      row["Id"],
		"details": details,
		"status":  row["Status"],
		"errors":  []interface{}{},
	})
}

// EditDocument handles POST /transactionsmodule/edit-document
func EditDocument(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)
	editedBy := addedByFromToken(user)

	// Form data
	idStr := c.PostForm("id")
	text := c.PostForm("text")

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, requiredField("id"))
		return
	}

	res, err := auth.ExecSP(db.DB, "v1_AdminRole_TransactionsModule_EditDocument",
		map[string]interface{}{
			"SiteUsersId": siteUsersId,
			"DocumentId":  id,
			"Text":        text,
			"EditedBy":    editedBy,
		}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"id":      0,
			"details": EditDocumentRequest{Id: id, Text: text},
			"status":  "1",
			"errors":  []interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      row["Id"],
		"details": row["Details"],
		"status":  row["Status"],
		"errors":  []interface{}{},
	})
}

// PinDocument handles POST /transactionsmodule/pin-document
func PinDocument(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)
	editedBy := addedByFromToken(user)

	var req PinDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "0", "error": "Invalid JSON body"})
		return
	}

	if req.Id <= 0 {
		c.JSON(http.StatusBadRequest, requiredField("id"))
		return
	}

	res, err := auth.ExecSP(db.DB, "v1_AdminRole_TransactionsModule_PinDocument",
		map[string]interface{}{
			"SiteUsersId": siteUsersId,
			"DocumentId":  req.Id,
			"bPinned":     req.BPinned,
			"EditedBy":    editedBy,
		}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	row, _ := auth.AsSingleRow(res)

	c.JSON(http.StatusOK, gin.H{
		"id": 0,
		"details": gin.H{
			"id":      req.Id,
			"bPinned": req.BPinned,
		},
		"status": func() string {
			if row != nil {
				if s, ok := row["Status"].(string); ok {
					return s
				}
			}
			return "1"
		}(),
		"errors": []interface{}{},
	})
}

// GetEditNote handles GET /transactionsmodule/edit-note
func GetEditNote(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)

	idStr := c.Query("id")
	if idStr == "" {
		idStr = c.Query("Id")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, requiredField("id"))
		return
	}

	res, err := auth.ExecSP(db.DB, "v1_AdminRole_TransactionsModule_GetEditNote",
		map[string]interface{}{
			"SiteUsersId": siteUsersId,
			"NotesId":     id,
		}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"id":      0,
			"details": nil,
			"status":  "1",
			"errors":  []interface{}{},
		})
		return
	}

	var details map[string]interface{}
	if detailsStr, ok := row["Details"].(string); ok && detailsStr != "" {
		_ = json.Unmarshal([]byte(detailsStr), &details)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      row["Id"],
		"details": details,
		"status":  row["Status"],
		"errors":  []interface{}{},
	})
}

// =============================================================================
// PendingTransactionsModule - Endpoints
// =============================================================================

// GetAssignDetails handles GET /pendingtransactionsmodule/assign
func GetAssignDetails(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)

	transactionsIdStr := c.Query("transactionsId")
	if transactionsIdStr == "" {
		c.JSON(http.StatusBadRequest, requiredField("transactionsId"))
		return
	}

	transactionsId, err := strconv.ParseInt(transactionsIdStr, 10, 64)
	if err != nil || transactionsId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"errors": []FieldError{{FieldName: "transactionsId", MessageCode: "Invalid"}},
		})
		return
	}

	res, err := auth.ExecSP(db.DB, "v1_AdminRole_PendingTransactionsModule_GetAssignDetails",
		map[string]interface{}{
			"SiteUsersId":                         siteUsersId,
			"CustomerAssetAccountsTransactionsId": transactionsId,
		}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"id":       0,
			"details":  nil,
			"metadata": []interface{}{},
			"status":   "1",
			"errors":   []interface{}{},
		})
		return
	}

	var details map[string]interface{}
	if detailsStr, ok := row["Details"].(string); ok && detailsStr != "" {
		_ = json.Unmarshal([]byte(detailsStr), &details)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       row["Id"],
		"details":  details,
		"metadata": row["Metadata"],
		"status":   row["Status"],
		"errors":   []interface{}{},
	})
}

// Assign handles POST /pendingtransactionsmodule/assign
func Assign(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)

	var req AssignFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "0", "error": "Invalid JSON body"})
		return
	}

	if req.CustomerAssetAccountsTransactionsId <= 0 {
		c.JSON(http.StatusOK, requiredField("customerAssetAccountsTransactionsId"))
		return
	}

	// Convert assigned users to JSON for SP
	assignedUsersJSON, _ := json.Marshal(req.AssignedUsers)

	res, err := auth.ExecSP(db.DB, "v1_AdminRole_PendingTransactionsModule_Assign",
		map[string]interface{}{
			"SiteUsersId":                         siteUsersId,
			"CustomerAssetAccountsTransactionsId": req.CustomerAssetAccountsTransactionsId,
			"AssignedUsersJson":                   string(assignedUsersJSON),
		}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	row, _ := auth.AsSingleRow(res)

	status := "1"
	if row != nil {
		if s, ok := row["Status"].(string); ok {
			status = s
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      req.CustomerAssetAccountsTransactionsId,
		"details": req,
		"status":  status,
		"errors":  []interface{}{},
	})
}

// Complete handles POST /pendingtransactionsmodule/complete
func Complete(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)
	addedBy := addedByFromToken(user)

	var req CompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "0", "error": "Invalid JSON body"})
		return
	}

	if req.CustomerAssetAccountsTransactionsId <= 0 {
		c.JSON(http.StatusOK, requiredField("customerAssetAccountsTransactionsId"))
		return
	}

	res, err := auth.ExecSP(db.DB, "v1_AdminRole_PendingTransactionsModule_Complete",
		map[string]interface{}{
			"SiteUsersId":                         siteUsersId,
			"CustomerAssetAccountsTransactionsId": req.CustomerAssetAccountsTransactionsId,
			"TransactionDetailsJson":              req.TransactionDetailsJson,
			"AddedBy":                             addedBy,
		}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": "Invalid SP response"})
		return
	}

	status := asString(row["Status"])
	if status != "1" {
		c.JSON(http.StatusOK, gin.H{
			"id":      row["Id"],
			"details": row["Details"],
			"status":  status,
			"errors":  row["Errors"],
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": req.CustomerAssetAccountsTransactionsId,
		"details": CompleteResponse{
			CustomerAssetAccountsTransactionsId: req.CustomerAssetAccountsTransactionsId,
		},
		"status": "1",
		"errors": []interface{}{},
	})
}

// PendingCancel handles POST /pendingtransactionsmodule/cancel
func PendingCancel(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)
	addedBy := addedByFromToken(user)

	var req PendingCancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "0", "error": "Invalid JSON body"})
		return
	}

	if req.CustomerAssetAccountsTransactionsId <= 0 {
		c.JSON(http.StatusOK, requiredField("customerAssetAccountsTransactionsId"))
		return
	}

	res, err := auth.ExecSP(db.DB, "v1_AdminRole_PendingTransactionsModule_Cancel",
		map[string]interface{}{
			"SiteUsersId":                         siteUsersId,
			"CustomerAssetAccountsTransactionsId": req.CustomerAssetAccountsTransactionsId,
			"TransactionDetailsJson":              req.TransactionDetailsJson,
			"bRefundFee":                          req.BRefundFee,
			"bCancelled":                          req.BCancelled,
			"AddedBy":                             addedBy,
		}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": "Invalid SP response"})
		return
	}

	status := asString(row["Status"])
	if status != "1" {
		c.JSON(http.StatusOK, gin.H{
			"id":      row["Id"],
			"details": row["Details"],
			"status":  status,
			"errors":  row["Errors"],
		})
		return
	}

	// Parse response to get cancellation request ID
	var details map[string]interface{}
	if detailsStr, ok := row["Details"].(string); ok && detailsStr != "" {
		_ = json.Unmarshal([]byte(detailsStr), &details)
	}

	cancellationRequestId := 0
	if details != nil {
		if v, ok := details["customerTransactionCancellationRequestsId"]; ok {
			cancellationRequestId = asInt(v)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id": cancellationRequestId,
		"details": PendingCancelResponse{
			CustomerTransactionCancellationRequestsId: cancellationRequestId,
			CustomerAssetAccountsTransactionsId:       req.CustomerAssetAccountsTransactionsId,
		},
		"status": "1",
		"errors": []interface{}{},
	})
}

// MarkAsPending handles POST /pendingtransactionsmodule/markaspending
func MarkAsPending(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)

	var req MarkAsPendingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "0", "error": "Invalid JSON body"})
		return
	}

	if req.CustomerAssetAccountsTransactionsId <= 0 {
		c.JSON(http.StatusOK, requiredField("customerAssetAccountsTransactionsId"))
		return
	}

	res, err := auth.ExecSP(db.DB, "v1_AdminRole_PendingTransactionsModule_MarkAsPending",
		map[string]interface{}{
			"SiteUsersId":                         siteUsersId,
			"CustomerAssetAccountsTransactionsId": req.CustomerAssetAccountsTransactionsId,
			"ExternalID":                          req.ExternalID,
			"TransactionDetailsJson":              req.TransactionDetailsJson,
		}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	row, _ := auth.AsSingleRow(res)

	status := "1"
	if row != nil {
		if s, ok := row["Status"].(string); ok {
			status = s
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      req.CustomerAssetAccountsTransactionsId,
		"details": nil,
		"status":  status,
		"errors":  []interface{}{},
	})
}

// =============================================================================
// FrozenTransactionsModule - Endpoints
// =============================================================================

// FrozenView handles GET /frozentransactionsmodule/view
func FrozenView(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)

	idStr := c.Query("id")
	if idStr == "" {
		idStr = c.Query("Id")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, requiredField("id"))
		return
	}

	res, err := auth.ExecSP(db.DB, "v1_AdminRole_FrozenTransactionsModule_GetViewDetails",
		map[string]interface{}{
			"SiteUsersId":           siteUsersId,
			"CustomerTransactionId": id,
		}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"id":       0,
			"details":  nil,
			"metadata": FrozenViewMetadata,
			"status":   "1",
			"errors":   []interface{}{},
		})
		return
	}

	var details map[string]interface{}
	if detailsStr, ok := row["Details"].(string); ok && detailsStr != "" {
		_ = json.Unmarshal([]byte(detailsStr), &details)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       row["Id"],
		"details":  details,
		"metadata": FrozenViewMetadata,
		"status":   row["Status"],
		"errors":   []interface{}{},
	})
}

// Unfreeze handles POST /frozentransactionsmodule/unfreeze
func Unfreeze(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)

	var req UnfreezeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "0", "error": "Invalid JSON body"})
		return
	}

	if req.CustomerTransactionID <= 0 {
		c.JSON(http.StatusBadRequest, requiredField("customerTransactionID"))
		return
	}

	res, err := auth.ExecSP(db.DB, "v1_AdminRole_FrozenTransactionsModule_Unfreeze",
		map[string]interface{}{
			"SiteUsersId":           siteUsersId,
			"CustomerTransactionId": req.CustomerTransactionID,
		}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	row, _ := auth.AsSingleRow(res)

	status := "1"
	if row != nil {
		if s, ok := row["Status"].(string); ok {
			status = s
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      req.CustomerTransactionID,
		"details": nil,
		"status":  status,
		"errors":  []interface{}{},
	})
}

// =============================================================================
// PendingTransactionsTreasuryModule - Cancel Endpoint
// =============================================================================

// TreasuryCancel handles POST /pendingtransactionstreasurymodule/cancel
func TreasuryCancel(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)
	addedBy := addedByFromToken(user)

	var req TreasuryCancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "0", "error": "Invalid JSON body"})
		return
	}

	if req.CustomerAssetAccountsTransactionsId <= 0 {
		c.JSON(http.StatusOK, requiredField("customerAssetAccountsTransactionsId"))
		return
	}

	// SP only accepts: @CustomerAssetAccountsTransactionsId, @SiteUsersID, @AddedBy
	res, err := auth.ExecSP(db.DB, "v1_AdminRole_PendingTransactionsTreasuryModule_Cancel",
		map[string]interface{}{
			"CustomerAssetAccountsTransactionsId": req.CustomerAssetAccountsTransactionsId,
			"SiteUsersID":                         siteUsersId,
			"AddedBy":                             addedBy,
		}, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": "Invalid SP response"})
		return
	}

	status := asString(row["Status"])
	if status != "1" {
		c.JSON(http.StatusOK, gin.H{
			"id":      row["Id"],
			"details": row["Details"],
			"status":  status,
			"errors":  row["Errors"],
		})
		return
	}

	// Parse response to get cancellation request ID
	var details map[string]interface{}
	if detailsStr, ok := row["Details"].(string); ok && detailsStr != "" {
		_ = json.Unmarshal([]byte(detailsStr), &details)
	}

	cancellationRequestId := 0
	if details != nil {
		if v, ok := details["customerTransactionCancellationRequestsId"]; ok {
			cancellationRequestId = asInt(v)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id": cancellationRequestId,
		"details": TreasuryCancelResponse{
			CustomerTransactionCancellationRequestsId: cancellationRequestId,
			CustomerAssetAccountsTransactionsId:       req.CustomerAssetAccountsTransactionsId,
		},
		"status": "1",
		"errors": []interface{}{},
	})
}

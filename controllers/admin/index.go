package admin

import (
	"cloud-web-phoenix-customer-v1-go/controllers/auth"
	"fmt"
	"strconv"

	"cloud-web-phoenix-customer-v1-go/db"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetUserCounts(c *gin.Context) {

	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	timeOption := c.Query("timeOption")
	spName := "v1_AdminRole_DashboardModule_GetUserCounts"
	if spName == "" {
		c.JSON(400, gin.H{"error": "Invalid chart endpoint"})
		return
	}

	params := map[string]interface{}{
		"SiteUsersId": user["id"],
	}
	if timeOption != "" {
		params["TimeOption"] = timeOption
	}
	res, err := auth.ExecSP(
		db.DB,
		spName,
		params,
		1, // single row
	)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	result, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(500, gin.H{"error": "Invalid SP response"})
		return
	}

	// ---------- Parse Details JSON safely ----------
	var details interface{} = []interface{}{}

	if raw, ok := result["Details"].(string); ok && raw != "" {
		cleanJSON := strings.ReplaceAll(raw, "'", "\"")
		if err := json.Unmarshal([]byte(cleanJSON), &details); err != nil {
			c.JSON(500, gin.H{
				"error":   "Invalid details data",
				"details": err.Error(),
			})
			return
		}
	}

	c.JSON(200, gin.H{
		"id":      result["Id"],
		"details": details,
		"status":  result["Status"],
		"errors":  []interface{}{},
		"message": "User counts fetched successfully",
	})
}
func GetAdminUsersData(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	siteUsersId := user["id"].(int)

	// .NET: QueryRecordListDto bind
	q := ParseQueryRecordList(c.Request.URL.Query())

	// .NET: Load list selections (SP) + OverrideWithExistingListSelections
	loadSelections := func() *ListSelections {
		sp := "v1_General_ListSelectionsModule_GetSiteUsersListSelections"
		params := map[string]interface{}{
			"SiteUsersId": siteUsersId,
			"TrackingId":  "DefaultTrackingID",
			"ListKey":     "AdminUsers",
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

	ex := loadSelections()
	OverrideWithSelections(&q, ex)

	// Endpoint specific config (whitelist)
	cfg := ListSPConfig{
		ListKey:    "AdminUsers",
		TrackingID: "DefaultTrackingID",
		ColumnMap: map[string]string{
			"adminUsers__Id":             "AdminUsers.Id",
			"adminUsers__AdminUsersCode": "AdminUsers.AdminUsersCode",
			"adminUsers__FirstName":      "AdminUsers.FirstName",
			"adminUsers__LastName":       "AdminUsers.LastName",
			"siteUsers__EmailAddress":    "SiteUsers.EmailAddress",
			"siteUsers__Username":        "SiteUsers.Username",
			"siteUsers__bSuppressed":     "SiteUsers.bSuppressed",
			"adminUsers__AddDate":        "AdminUsers.AddDate",
		},
		SearchFields: []string{
			"AdminUsers.FirstName",
			"AdminUsers.LastName",
			"SiteUsers.EmailAddress",
			"AdminUsers.AdminUsersCode",
			"SiteUsers.Username",
		},
	}

	// .NET: AdminListHelper -> SP params
	spParams := BuildListSPParams(q, siteUsersId, cfg)

	// .NET: SP call
	spName := "v1_AdminRole_AdminUsersModule_List"
	res, err := auth.ExecSP(db.DB, spName, spParams, 2)

	// .NET: columns metadata (frontend ko filter/sort UI ke liye)
	columns := GetAdminUsersModuleListColumns()

	respond := func(listData []map[string]interface{}, total int) {
		details := BuildListDetails(columns, listData, q, total)
		c.JSON(http.StatusOK, gin.H{
			"status":  "1",
			"id":      siteUsersId,
			"errors":  []string{},
			"details": details,
		})
	}

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			respond([]map[string]interface{}{}, 0)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	rows, ok := res.([]map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": "Invalid SP response"})
		return
	}

	// .NET: row mapping + HowManyResults total count
	listData := make([]map[string]interface{}, 0, len(rows))
	total := 0

	for _, row := range rows {
		item := NormalizeRowKeys(row)
		listData = append(listData, item)

		if v, ok := row["HowManyResults"]; ok && v != nil {
			switch t := v.(type) {
			case int:
				total = t
			case int64:
				total = int(t)
			case float64:
				total = int(t)
			default:
				i, _ := strconv.Atoi(fmt.Sprint(t))
				if i > 0 {
					total = i
				}
			}
		}
	}

	respond(listData, total)
}
func GetLicenseAdminUsersData(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	siteUsersId := user["id"].(int)

	// .NET: QueryRecordListDto bind
	q := ParseQueryRecordList(c.Request.URL.Query())

	// .NET: Load list selections (SP) + OverrideWithExistingListSelections
	loadSelections := func() *ListSelections {
		sp := "v1_General_ListSelectionsModule_GetSiteUsersListSelections"
		params := map[string]interface{}{
			"SiteUsersId": siteUsersId,
			"TrackingId":  "DefaultTrackingID",
			"ListKey":     "LicenseeAdminUsers",
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

	ex := loadSelections()
	OverrideWithSelections(&q, ex)

	// Endpoint specific config (whitelist)
	cfg := ListSPConfig{
		ListKey:    "LicenseeAdminUsers",
		TrackingID: "DefaultTrackingID",
		ColumnMap: map[string]string{
			"adminUsers__Id":             "AdminUsers.Id",
			"adminUsers__AdminUsersCode": "AdminUsers.AdminUsersCode",
			"adminUsers__FirstName":      "AdminUsers.FirstName",
			"adminUsers__LastName":       "AdminUsers.LastName",
			"adminUsers__JobTitle":       "AdminUsers.JobTitle",
			"siteUsers__EmailAddress":    "SiteUsers.EmailAddress",
			"siteUsers__Username":        "SiteUsers.Username",
			"siteUsers__bSuppressed":     "SiteUsers.bSuppressed",
			"adminUsers__AddDate":        "AdminUsers.AddDate",
			"licensees__LicenseeName":    "Licensees.LicenseeName",
		},
		SearchFields: []string{
			"AdminUsers.FirstName",
			"AdminUsers.LastName",
			"SiteUsers.EmailAddress",
			"AdminUsers.AdminUsersCode",
			"SiteUsers.Username",
			"AdminUsers.JobTitle",
			"Licensees.LicenseeName",
		},
	}

	// .NET: AdminListHelper -> SP params
	spParams := BuildListSPParams(q, siteUsersId, cfg)

	// .NET: SP call
	spName := "v1_AdminRole_LicenseeAdminUsersModule_List"
	res, err := auth.ExecSP(db.DB, spName, spParams, 2)

	// .NET: columns metadata (frontend ko filter/sort UI ke liye)
	columns := GetLicenseeAdminUsersModuleListColumns()

	respond := func(listData []map[string]interface{}, total int) {
		details := BuildListDetails(columns, listData, q, total)
		c.JSON(http.StatusOK, gin.H{
			"status":  "1",
			"id":      siteUsersId,
			"errors":  []string{},
			"details": details,
		})
	}

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			respond([]map[string]interface{}{}, 0)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	rows, ok := res.([]map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": "Invalid SP response"})
		return
	}

	// .NET: row mapping + HowManyResults total count
	listData := make([]map[string]interface{}, 0, len(rows))
	total := 0

	for _, row := range rows {
		item := NormalizeRowKeys(row)
		listData = append(listData, item)

		if v, ok := row["HowManyResults"]; ok && v != nil {
			switch t := v.(type) {
			case int:
				total = t
			case int64:
				total = int(t)
			case float64:
				total = int(t)
			default:
				i, _ := strconv.Atoi(fmt.Sprint(t))
				if i > 0 {
					total = i
				}
			}
		}
	}

	respond(listData, total)
}
func GetAdminRolesData(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	siteUsersId := user["id"].(int)

	// .NET: QueryRecordListDto bind
	q := ParseQueryRecordList(c.Request.URL.Query())

	// .NET: Load list selections (SP) + OverrideWithExistingListSelections
	loadSelections := func() *ListSelections {
		sp := "v1_General_ListSelectionsModule_GetSiteUsersListSelections"
		params := map[string]interface{}{
			"SiteUsersId": siteUsersId,
			"TrackingId":  "DefaultTrackingID",
			"ListKey":     "AdminRoles",
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

	ex := loadSelections()
	OverrideWithSelections(&q, ex)

	// Endpoint specific config (whitelist)
	cfg := ListSPConfig{
		ListKey:    "AdminRoles",
		TrackingID: "DefaultTrackingID",
		ColumnMap: map[string]string{
			"adminRoles__Id":          "AdminRoles.Id",
			"adminRoles__Name":        "AdminRoles.Name",
			"adminRoles__Level":       "AdminRoles.Level",
			"adminRoles__bSuppressed": "AdminRoles.bSuppressed",
			"adminRoles__AddDate":     "AdminRoles.AddDate",
		},
		SearchFields: []string{
			"AdminRoles.Name",
			"AdminRoles.Level",
		},
	}

	// .NET: AdminListHelper -> SP params
	spParams := BuildListSPParams(q, siteUsersId, cfg)

	// .NET: SP call
	spName := "v1_AdminRole_AdminRolesModule_List"
	res, err := auth.ExecSP(db.DB, spName, spParams, 2)

	// .NET: columns metadata (frontend ko filter/sort UI ke liye)
	columns := GetAdminRolesModuleListColumns()

	respond := func(listData []map[string]interface{}, total int) {
		details := BuildListDetails(columns, listData, q, total)
		c.JSON(http.StatusOK, gin.H{
			"status":  "1",
			"id":      siteUsersId,
			"errors":  []string{},
			"details": details,
		})
	}

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			respond([]map[string]interface{}{}, 0)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	rows, ok := res.([]map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": "Invalid SP response"})
		return
	}

	// .NET: row mapping + HowManyResults total count
	listData := make([]map[string]interface{}, 0, len(rows))
	total := 0

	for _, row := range rows {
		item := NormalizeRowKeys(row)
		listData = append(listData, item)

		if v, ok := row["HowManyResults"]; ok && v != nil {
			switch t := v.(type) {
			case int:
				total = t
			case int64:
				total = int(t)
			case float64:
				total = int(t)
			default:
				i, _ := strconv.Atoi(fmt.Sprint(t))
				if i > 0 {
					total = i
				}
			}
		}
	}

	respond(listData, total)
}
func GetAdminRolesCreate(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	// siteUsersId := 0
	// if v, ok := user["id"].(int); ok {
	// 	siteUsersId = v
	// }

	// URL: /create?level=Admin|Licensee|LicenseeBrand
	level := c.Query("level")
	if level == "" {
		level = "Admin"
	}
	row, details, err := LoadAdminRolesCreateDetails(level)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"id":      0,
			"details": nil,
			"status":  "0",
			"errors":  []map[string]interface{}{{"fieldName": "General", "messageCode": err.Error()}},
		})
		return
	}

	// SP status/errors
	status := "1"
	if s, ok := row["Status"]; ok {
		// SP might return int/bit
		switch v := s.(type) {
		case int:
			if v == 0 {
				status = "0"
			}
		case int64:
			if v == 0 {
				status = "0"
			}
		case float64:
			if int(v) == 0 {
				status = "0"
			}
		case string:
			if v == "0" {
				status = "0"
			}
		}
	}

	// Errors: SP returns @ValidationMessage as string
	errorsArr := []interface{}{}
	if status == "0" {
		if msg, ok := row["Errors"].(string); ok && msg != "" {
			// aap chahein to isko .NET style field errors me parse bhi kar sakte ho
			errorsArr = append(errorsArr, gin.H{
				"fieldName":   "AdminRoleLevel",
				"messageCode": msg,
			})
		}
	}

	// metadata always same
	metadata := GetAdminRolesCreateMetadata()

	c.JSON(http.StatusOK, gin.H{
		"id":       0,
		"details":  details,
		"metadata": metadata,
		"status":   status,
		"errors":   errorsArr,
	})
}

func PostCreateAdminRole(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	siteUsersId := user["id"].(int)
	addedBy := getAddedByFromToken(user)

	var req AdminRoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "0", "errors": []string{"Invalid JSON body"}})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "0", "errors": []string{"Name is required"}})
		return
	}
	if !isValidAdminRoleLevel(req.Level) {
		c.JSON(http.StatusBadRequest, gin.H{"status": "0", "errors": []string{"Invalid level"}})
		return
	}
	if req.AccessRights == nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "0", "errors": []string{"AccessRights is required"}})
		return
	}

	dbRes, err := spCreateAdminRole(siteUsersId, addedBy, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "errors": []string{err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      dbRes.Id,
		"details": dbRes.Details,
		"status":  dbRes.Status,
		"errors":  dbRes.Errors,
	})
}

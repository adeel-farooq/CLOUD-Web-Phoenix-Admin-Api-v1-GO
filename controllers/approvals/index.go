package approvals

import (
    "cloud-web-phoenix-customer-v1-go/db"
    "cloud-web-phoenix-customer-v1-go/controllers/auth"
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

// GET /approvalsmodule/list
func GetApprovalsList(c *gin.Context) {
    claims, exists := c.Get("claims")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
        return
    }

    userClaims := claims.(jwt.MapClaims)
    siteUsersId, ok := getUserIdFromClaims(userClaims)
    if !ok || siteUsersId == 0 {
        c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized: invalid user id"})
        return
    }

    // baaki tumhara existing logic same
    q := ParseQueryRecordList(c.Request.URL.Query())
    ex := LoadSelections(siteUsersId, "DefaultTrackingID", "Admin_ApprovalsList")
    OverrideWithSelections(&q, ex)

    cfg := ListSPConfig{
        ListKey:    "Admin_ApprovalsList",
        TrackingID: "DefaultTrackingID",
        ColumnMap:  approvalsColumnMap(),
        SearchFields: []string{
            "Approvals.Code",
            "Approvals.Type",
            "Approvals.Description",
        },
    }

    spParams := BuildListSPParams(q, siteUsersId, cfg)
    spName := "v1_AdminRole_ApprovalsModule_GetApprovalsList"

    res, err := auth.ExecSP(db.DB, spName, spParams, 2)
    columns := GetApprovalsModuleListColumns()

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
        return
    }

    rows, ok := res.([]map[string]interface{})
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": "Invalid SP response"})
        return
    }

    listData := []map[string]interface{}{}
    total := 0
    for _, row := range rows {
        item := NormalizeRowKeys(row)
        listData = append(listData, item)
        if v, ok := row["HowManyResults"]; ok && v != nil {
            total = toInt(v)
        }
    }

    details := BuildListDetails(columns, listData, q, total)
    c.JSON(http.StatusOK, gin.H{"status": "1", "id": siteUsersId, "errors": []string{}, "details": details})
}





// GET /approvalsmodule/view
func ViewApprovalDetails(c *gin.Context) {
    claims, exists := c.Get("claims")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
        return
    }

    userClaims := claims.(jwt.MapClaims)
    siteUsersId, ok := getUserIdFromClaims(userClaims)
    if !ok || siteUsersId == 0 {
        c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized: invalid user id"})
        return
    }

    idStr := c.Query("ApprovalsId")
    if idStr == "" {
        c.JSON(http.StatusBadRequest, gin.H{"status": "0", "error": "ApprovalsId is required"})
        return
    }
    id, _ := strconv.Atoi(idStr)

    spName := "v1_AdminRole_ApprovalsModule_GetViewDetails"
    params := map[string]interface{}{
        "ApprovalsId": id,
        "SiteUsersID": siteUsersId,
    }

    res, err := auth.ExecSP(db.DB, spName, params, 1)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
        return
    }

    row, ok := auth.AsSingleRow(res)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": "Invalid SP response"})
        return
    }

    metadata := GetApprovalsViewMetadata()

    c.JSON(http.StatusOK, gin.H{
        "id":       toInt(row["Id"]),
        "details":  row["Details"],
        "metadata": metadata,
        "status":   row["Status"],
        "errors":   []string{},
    })
}




// POST /approvalsmodule/process
func ProcessApproval(c *gin.Context) {
    claims, exists := c.Get("claims")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
        return
    }

    userClaims := claims.(jwt.MapClaims)
    siteUsersId, ok := getUserIdFromClaims(userClaims)
    if !ok || siteUsersId == 0 {
        c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized: invalid user id"})
        return
    }

    editedBy := getAddedByFromClaims(userClaims)

    var req ProcessApprovalRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"status": "0", "errors": []string{"Invalid JSON body"}})
        return
    }

    spName := "v1_AdminRole_ApprovalsModule_ProcessApproval"
    params := map[string]interface{}{
    "ApprovalsId": req.ApprovalsId,
    "BApprove":    req.BApprove,
    "SiteUsersID": siteUsersId,
    "EditedBy":    editedBy,
    "CanApprove":  1, // or true
}


    res, err := auth.ExecSP(db.DB, spName, params, 1)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "errors": []string{err.Error()}})
        return
    }

    row, ok := auth.AsSingleRow(res)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "errors": []string{"Invalid SP response"}})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "id":      row["Id"],
        "details": row["Details"],
        "status":  row["Status"],
        "errors":  row["Errors"],
    })
}



// --------------------
// Helper: getAddedByFromClaims
// --------------------
func getAddedByFromClaims(claims jwt.MapClaims) string {
    if v, ok := claims["username"].(string); ok {
        return v
    }
    if v, ok := claims["email"].(string); ok {
        return v
    }
    if v, ok := claims["UserCode"].(string); ok {
        return v
    }
    return ""
}


// --------------------
// Helper: getAddedByFromToken
// --------------------
func getAddedByFromToken(user map[string]interface{}) string {
    if v, ok := user["username"].(string); ok {
        return v
    }
    if v, ok := user["email"].(string); ok {
        return v
    }
    return ""
}

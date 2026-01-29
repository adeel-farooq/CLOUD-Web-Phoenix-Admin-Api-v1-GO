package approval

import (
    "fmt"
    "log"
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"

    "cloud-web-phoenix-customer-v1-go/db"
    "cloud-web-phoenix-customer-v1-go/models"
)


// ---------------------------------------------------------
// Helper: Convert DB model → Frontend expected JSON format
// ---------------------------------------------------------
func MapApprovalTypeToViewResponse(model models.ApprovalType) models.ApprovalTypeViewResponse {

    var createdDate string
    if model.CreatedDate != nil {
        createdDate = model.CreatedDate.Format(time.RFC3339)
    }

    var createdBy string
    if model.CreatedBy != nil {
        createdBy = *model.CreatedBy
    }

    return models.ApprovalTypeViewResponse{
        ID: model.Id,
        Details: models.ApprovalTypeDetailsDto{
            Code:         fmt.Sprintf("AT%d", model.Id),
            Name:         model.Name,
            MinApprovers: model.ApprovalLevels,
            AddDate:      createdDate,
            AddedBy:      createdBy,
        },
        Metadata: []any{},
    }
}



// ---------------------------------------------------------
// GET LIST — /approvaltypesmodule/list
// ---------------------------------------------------------
func GetApprovalTypesListAsync(c *gin.Context) {

    userInfo, exists := c.Get("user")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
        return
    }

    userClaims, ok := userInfo.(jwt.MapClaims)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid user claims"})
        return
    }

    var userId string
    if val, ok := userClaims["user_id"]; ok {
        userId = fmt.Sprintf("%v", val)
    } else if val, ok := userClaims["UserCode"]; ok {
        userId = fmt.Sprintf("%v", val)
    }

    pageNo, _ := strconv.Atoi(c.DefaultQuery("pageNo", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

    if pageNo <= 0 {
        pageNo = 1
    }
    if pageSize <= 0 {
        pageSize = 10
    }

    approvalTypes, totalCount, err := db.GetApprovalTypesList(pageNo, pageSize)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Failed to retrieve approval types"})
        return
    }

    var responses []models.ApprovalTypeViewResponse
    for _, at := range approvalTypes {
        responses = append(responses, MapApprovalTypeToViewResponse(at))
    }

    log.Printf("[DEBUG] User %s fetched %d approval types", userId, len(responses))

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": gin.H{
            "approvalTypes": responses,
            "totalRecords":  totalCount,
            "pageNo":        pageNo,
            "pageSize":      pageSize,
        },
    })
}



// ---------------------------------------------------------
// VIEW — /approvaltypesmodule/view?id=
// ---------------------------------------------------------
func ViewApprovalTypesAsync(c *gin.Context) {

    var query models.ViewApprovalTypeCommandDto
    if err := c.ShouldBindQuery(&query); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": true, "message": "Invalid request parameters"})
        return
    }

    approvalType, err := db.GetApprovalTypeByID(query.Id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Database error"})
        return
    }

    if approvalType == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": true, "message": "Approval type not found"})
        return
    }

    response := MapApprovalTypeToViewResponse(*approvalType)

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    response,
    })
}



// ---------------------------------------------------------
// GET EDIT — /approvaltypesmodule/edit?id=
// ---------------------------------------------------------
func GetEditApprovalTypesAsync(c *gin.Context) {

    userInfo, exists := c.Get("user")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": true, "message": "Unauthorized"})
        return
    }

    userClaims, ok := userInfo.(jwt.MapClaims)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": true, "message": "Invalid user claims"})
        return
    }

    var userId string
    if val, ok := userClaims["user_id"]; ok {
        userId = fmt.Sprintf("%v", val)
    } else if val, ok := userClaims["UserCode"]; ok {
        userId = fmt.Sprintf("%v", val)
    }

    var query models.ApprovalTypeEditCommandDto
    if err := c.ShouldBindQuery(&query); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": true, "message": "Invalid request parameters", "details": "id is required"})
        return
    }

    approvalType, err := db.GetApprovalTypeByID(query.Id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Database error"})
        return
    }

    if approvalType == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": true, "message": "Approval type not found"})
        return
    }

    log.Printf("[DEBUG] User %s requested edit for approval type ID %d", userId, query.Id)

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    approvalType,
    })
}



// ---------------------------------------------------------
// POST EDIT — /approvaltypesmodule/edit
// ---------------------------------------------------------
func EditApprovalTypesAsync(c *gin.Context) {

    userInfo, exists := c.Get("user")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": true, "message": "Unauthorized"})
        return
    }

    userClaims, ok := userInfo.(jwt.MapClaims)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": true, "message": "Invalid user claims"})
        return
    }

    var userId, userName string

    if val, ok := userClaims["user_id"]; ok {
        userId = fmt.Sprintf("%v", val)
    } else if val, ok := userClaims["UserCode"]; ok {
        userId = fmt.Sprintf("%v", val)
    }

    if val, ok := userClaims["username"]; ok {
        userName = fmt.Sprintf("%v", val)
    } else if val, ok := userClaims["FirstName"]; ok {
        userName = fmt.Sprintf("%v", val)
    }

    var editReq models.ApprovalTypeEditDto
    if err := c.ShouldBindJSON(&editReq); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": true, "message": "Invalid request body", "details": err.Error()})
        return
    }

    if editReq.Name == "" || editReq.ApprovalLevels <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": true, "message": "Validation failed"})
        return
    }

    err := db.UpdateApprovalType(&editReq, userName)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": true, "message": "Failed to update approval type"})
        return
    }

    log.Printf("[INFO] User %s updated approval type ID %d", userId, editReq.Id)

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Approval type updated successfully",
        "data": gin.H{
            "id":   editReq.Id,
            "name": editReq.Name,
        },
    })
}

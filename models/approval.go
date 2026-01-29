package models

import "time"

// ApprovalType represents approval type configuration
type ApprovalType struct {
    Id                        int        `json:"id"`
    Name                      string     `json:"name"`
    Description               string     `json:"description"`
    RequiresMultipleApprovals bool       `json:"requiresMultipleApprovals"`
    ApprovalLevels            int        `json:"approvalLevels"`
    MaxApprovalTimeMinutes    int        `json:"maxApprovalTimeMinutes"`
    IsActive                  bool       `json:"isActive"`
    CreatedDate               *time.Time `json:"createdDate"`
    CreatedBy                 *string    `json:"createdBy"`
    ModifiedDate              *time.Time `json:"modifiedDate,omitempty"`
    ModifiedBy                *string    `json:"modifiedBy,omitempty"`
}

// Approval represents an approval request
type Approval struct {
    Id             int        `json:"id"`
    ApprovalTypeId int        `json:"approvalTypeId"`
    ReferenceId    string     `json:"referenceId"`
    ReferenceType  string     `json:"referenceType"`
    Amount         *float64   `json:"amount,omitempty"`
    Status         string     `json:"status"`
    CreatedDate    *time.Time `json:"createdDate"`
    CreatedBy      *string    `json:"createdBy"`
    Notes          string     `json:"notes"`
    ApprovedDate   *time.Time `json:"approvedDate,omitempty"`
    ApprovedBy     *string    `json:"approvedBy,omitempty"`
    ApprovalNotes  *string    `json:"approvalNotes"`
}

// --- DTOs for list responses ---
type ApprovalTypeListDto struct {
    ApprovalTypes []ApprovalType `json:"approvalTypes"`
    TotalRecords  int            `json:"totalRecords"`
    PageNo        int            `json:"pageNo"`
    PageSize      int            `json:"pageSize"`
}

// --- DTOs for edit requests ---
type ApprovalTypeEditDto struct {
    Id                        int    `json:"id" binding:"required"`
    Name                      string `json:"name" binding:"required"`
    Description               string `json:"description"`
    RequiresMultipleApprovals bool   `json:"requiresMultipleApprovals"`
    ApprovalLevels            int    `json:"approvalLevels" binding:"required,gt=0"`
    MaxApprovalTimeMinutes    int    `json:"maxApprovalTimeMinutes"`
    IsActive                  bool   `json:"isActive"`
}

type ApprovalTypeEditCommandDto struct {
    Id int `form:"id" binding:"required"`
}

type ViewApprovalTypeCommandDto struct {
    Id int `form:"id" binding:"required"`
}

type QueryRecordListDto struct {
    PageNo     int    `form:"pageNo"`
    PageSize   int    `form:"pageSize"`
    SearchText string `form:"searchText"`
}

// --- NEW: DTOs for frontend expected format ---
type ApprovalTypeDetailsDto struct {
    Code         string `json:"code"`
    Name         string `json:"name"`
    MinApprovers int    `json:"minApprovers"`
    AddDate      string `json:"addDate"`
    AddedBy      string `json:"addedBy"`
}

type ApprovalTypeViewResponse struct {
    ID       int                    `json:"id"`
    Details  ApprovalTypeDetailsDto `json:"details"`
    Metadata []any                  `json:"metadata"`
}

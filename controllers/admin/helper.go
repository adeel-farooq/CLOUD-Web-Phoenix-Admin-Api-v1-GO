package admin

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func SendAuthError(c *gin.Context, errorType int) {
	errorMsg := []gin.H{}
	if errorType == 0 {

		errorMsg = []gin.H{
			{"fieldName": "Username", "messageCode": "Username_Or_Password_Incorrect"},
			{"fieldName": "Password", "messageCode": "Username_Or_Password_Incorrect"},
		}
	} else if errorType == 1 {
		errorMsg = []gin.H{
			{"fieldName": "TfaCode", "messageCode": "Invalid"},
		}
	} else if errorType == 2 {
		errorMsg = []gin.H{
			{"fieldName": "TfaType", "messageCode": "TfaType_Invalid"},
		}

	}
	c.JSON(http.StatusUnauthorized, gin.H{
		"id":      0,
		"details": nil,
		"status":  "0",
		"errors":  errorMsg,
	})
}

// getChartSPName maps chart endpoints to their respective stored procedure names
func getChartSPName(path string) string {
	switch {
	case strings.Contains(path, "/depositschart"):
		return "v1_AdminRole_DashboardModule_GetDeposits"
	case strings.Contains(path, "/withdrawalschart"):
		return "v1_AdminRole_DashboardModule_GetWithdrawals"
	case strings.Contains(path, "/cardsissuedchart"):
		return "v1_AdminRole_DashboardModule_GetNumberCardsIssued"
	case strings.Contains(path, "/cardspendchart"):
		return "v1_AdminRole_DashboardModule_GetCardSpend"
	case strings.Contains(path, "/turnoverchart"):
		return "v1_AdminRole_DashboardModule_GetTurnover"
	case strings.Contains(path, "/revenuechart"):
		return "v1_AdminRole_DashboardModule_GetRevenue"
	case strings.Contains(path, "/usercounts"):
		return "v1_AdminRole_DashboardModule_GetUserCounts"
	default:
		return ""
	}
}

// func GetSiteUsersListSelections(siteUsersId int, trackingId, listKey string) (*SiteUsersListSelectionsDto, error) {
//     query := "exec v1_General_ListSelectionsModule_GetSiteUsersListSelections @SiteUsersId=@SiteUsersId, @TrackingId=@TrackingId, @ListKey=@ListKey"
//     row := db.DB.QueryRow(query,
//         sql.Named("SiteUsersId", siteUsersId),
//         sql.Named("TrackingId", trackingId),
//         sql.Named("ListKey", listKey),
//     )

//     var filtersJSON, sortBy, search string
//     if err := row.Scan(&filtersJSON, &sortBy, &search); err != nil {
//         return nil, err
//     }

//     filters := make(map[string]string)
//     _ = json.Unmarshal([]byte(filtersJSON), &filters)

//     return &SiteUsersListSelectionsDto{
//         Filters: filters,
//         SortBy:  sortBy,
//         Search:  search,
//     }, nil
// }

// func GetAdminUserList(query QueryRecordListDto, siteUsersId int) (*AdminUserListDto, error) {
//     sp := "exec v1_AdminRole_AdminUsersModule_List @User_SiteUsersID=@User_SiteUsersID"
//     rows, err := db.DB.Query(sp, sql.Named("User_SiteUsersID", siteUsersId))
//     if err != nil {
//         return nil, err
//     }
//     defer rows.Close()

//     var users []AdminUserListDataRow
//     for rows.Next() {
//         var row AdminUserListDataRow
//         if err := rows.Scan(&row.ID, &row.Username, &row.Email); err != nil {
//             return nil, err
//         }
//         users = append(users, row)
//     }

//     return &AdminUserListDto{
//         Users: users,
//         Total: len(users),
//     }, nil
// }

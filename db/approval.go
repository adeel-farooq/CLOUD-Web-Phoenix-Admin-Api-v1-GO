package db

import (
	"database/sql"
	"fmt"
	"log"

	"cloud-web-phoenix-customer-v1-go/models"
)


// =============================
// GET APPROVAL TYPES LIST
// =============================
func GetApprovalTypesList(pageNo, pageSize int) ([]models.ApprovalType, int, error) {

	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (pageNo - 1) * pageSize

	// Count
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM ApprovalTypes`
	err := DB.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		log.Printf("[ERROR] Count error: %v", err)
		return nil, 0, err
	}

	// Main Query
	query := `
		SELECT 
			Id,
			Name,
			Description,
			MinApprovers AS ApprovalLevels,
			MaxApprovalTimeMinutes,
			IsActive,
			CreatedDate,
			CreatedBy,
			ModifiedDate,
			ModifiedBy
		FROM ApprovalTypes
		ORDER BY CreatedDate DESC
		OFFSET @Offset ROWS FETCH NEXT @PageSize ROWS ONLY
	`

	rows, err := DB.Query(query, sql.Named("Offset", offset), sql.Named("PageSize", pageSize))
	if err != nil {
		log.Printf("[ERROR] Query error: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var approvalTypes []models.ApprovalType

	for rows.Next() {

		var at models.ApprovalType

		var description sql.NullString
		var approvalLevels sql.NullInt32
		var maxApprovalTimeMinutes sql.NullInt32
		var isActive sql.NullBool
		var createdBy sql.NullString
		var modifiedBy sql.NullString
		var modifiedDate sql.NullTime
		var createdDate sql.NullTime

		err := rows.Scan(
			&at.Id,
			&at.Name,
			&description,
			&approvalLevels,
			&maxApprovalTimeMinutes,
			&isActive,
			&createdDate,
			&createdBy,
			&modifiedDate,
			&modifiedBy,
		)

		if err != nil {
			log.Printf("[ERROR] Scan error: %v", err)
			return nil, 0, err
		}

		// NULL HANDLING
		if description.Valid {
			at.Description = description.String
		}
		if approvalLevels.Valid {
			at.ApprovalLevels = int(approvalLevels.Int32)
		}
		if maxApprovalTimeMinutes.Valid {
			at.MaxApprovalTimeMinutes = int(maxApprovalTimeMinutes.Int32)
		}
		if isActive.Valid {
			at.IsActive = isActive.Bool
		}
		if createdBy.Valid {
			at.CreatedBy = &createdBy.String
		}
		if modifiedDate.Valid {
			at.ModifiedDate = &modifiedDate.Time
		}
		if modifiedBy.Valid {
			at.ModifiedBy = &modifiedBy.String
		}
		if createdDate.Valid {
    at.CreatedDate = &createdDate.Time
}

		approvalTypes = append(approvalTypes, at)
	}

	return approvalTypes, totalCount, nil
}



// =============================
// GET APPROVAL TYPE BY ID
// =============================
func GetApprovalTypeByID(id int) (*models.ApprovalType, error) {

	query := `
		SELECT 
			Id,
			Name,
			Description,
			MinApprovers AS ApprovalLevels,
			MaxApprovalTimeMinutes,
			IsActive,
			CreatedDate,
			CreatedBy,
			ModifiedDate,
			ModifiedBy
		FROM ApprovalTypes
		WHERE Id = @Id
	`

	var at models.ApprovalType

	var description sql.NullString
	var approvalLevels sql.NullInt32
	var maxApprovalTimeMinutes sql.NullInt32
	var isActive sql.NullBool
	var createdBy sql.NullString
	var modifiedBy sql.NullString
	var modifiedDate sql.NullTime
	var createdDate sql.NullTime

	err := DB.QueryRow(query, sql.Named("Id", id)).Scan(
		&at.Id,
		&at.Name,
		&description,
		&approvalLevels,
		&maxApprovalTimeMinutes,
		&isActive,
		&createdDate,
		&createdBy,
		&modifiedDate,
		&modifiedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// NULL HANDLING
	if description.Valid {
		at.Description = description.String
	}
	if approvalLevels.Valid {
		at.ApprovalLevels = int(approvalLevels.Int32)
	}
	if maxApprovalTimeMinutes.Valid {
		at.MaxApprovalTimeMinutes = int(maxApprovalTimeMinutes.Int32)
	}
	if isActive.Valid {
		at.IsActive = isActive.Bool
	}
	if createdBy.Valid {
		at.CreatedBy = &createdBy.String
	}
	if modifiedDate.Valid {
		at.ModifiedDate = &modifiedDate.Time
	}
	if modifiedBy.Valid {
		at.ModifiedBy = &modifiedBy.String
	}
	if createdDate.Valid {
    at.CreatedDate = &createdDate.Time
}
	

	return &at, nil
}



// =============================
// UPDATE APPROVAL TYPE
// =============================
func UpdateApprovalType(approvalType *models.ApprovalTypeEditDto, modifiedBy string) error {

	query := `
		UPDATE ApprovalTypes
		SET 
			Name = @Name,
			Description = @Description,
			MinApprovers = @ApprovalLevels,
			MaxApprovalTimeMinutes = @MaxApprovalTimeMinutes,
			IsActive = @IsActive,
			ModifiedDate = GETUTCDATE(),
			ModifiedBy = @ModifiedBy
		WHERE Id = @Id
	`

	result, err := DB.Exec(
		query,
		sql.Named("Name", approvalType.Name),
		sql.Named("Description", approvalType.Description),
		sql.Named("ApprovalLevels", approvalType.ApprovalLevels),
		sql.Named("MaxApprovalTimeMinutes", approvalType.MaxApprovalTimeMinutes),
		sql.Named("IsActive", approvalType.IsActive),
		sql.Named("ModifiedBy", modifiedBy),
		sql.Named("Id", approvalType.Id),
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("approval type not found")
	}

	return nil
}



// =============================
// CREATE APPROVAL
// =============================
func CreateApproval(approval *models.Approval) (int, error) {

	query := `
		INSERT INTO Approvals (
			ApprovalTypeId, ReferenceId, ReferenceType, Amount,
			Status, CreatedDate, CreatedBy, Notes
		)
		VALUES (
			@ApprovalTypeId, @ReferenceId, @ReferenceType, @Amount,
			@Status, GETUTCDATE(), @CreatedBy, @Notes
		)
		SELECT @@IDENTITY
	`

	var id int

	err := DB.QueryRow(
		query,
		sql.Named("ApprovalTypeId", approval.ApprovalTypeId),
		sql.Named("ReferenceId", approval.ReferenceId),
		sql.Named("ReferenceType", approval.ReferenceType),
		sql.Named("Amount", approval.Amount),
		sql.Named("Status", approval.Status),
		sql.Named("CreatedBy", approval.CreatedBy),
		sql.Named("Notes", approval.Notes),
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}



// =============================
// GET APPROVALS BY TYPE ID
// =============================
func GetApprovalsByTypeID(approvalTypeID int, pageNo, pageSize int) ([]models.Approval, int, error) {

	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (pageNo - 1) * pageSize

	var totalCount int
	countQuery := `SELECT COUNT(*) FROM Approvals WHERE ApprovalTypeId = @ApprovalTypeId`
	err := DB.QueryRow(countQuery, sql.Named("ApprovalTypeId", approvalTypeID)).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT 
			Id, ApprovalTypeId, ReferenceId, ReferenceType, Amount,
			Status, CreatedDate, CreatedBy, Notes, ApprovedDate, ApprovedBy, ApprovalNotes
		FROM Approvals
		WHERE ApprovalTypeId = @ApprovalTypeId
		ORDER BY CreatedDate DESC
		OFFSET @Offset ROWS FETCH NEXT @PageSize ROWS ONLY
	`

	rows, err := DB.Query(
		query,
		sql.Named("ApprovalTypeId", approvalTypeID),
		sql.Named("Offset", offset),
		sql.Named("PageSize", pageSize),
	)

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var approvals []models.Approval

	for rows.Next() {

		var ap models.Approval
		var amount sql.NullFloat64
		var approvedDate sql.NullTime
		var approvedBy sql.NullString

		err := rows.Scan(
			&ap.Id,
			&ap.ApprovalTypeId,
			&ap.ReferenceId,
			&ap.ReferenceType,
			&amount,
			&ap.Status,
			&ap.CreatedDate,
			&ap.CreatedBy,
			&ap.Notes,
			&approvedDate,
			&approvedBy,
			&ap.ApprovalNotes,
		)

		if err != nil {
			return nil, 0, err
		}

		if amount.Valid {
			ap.Amount = &amount.Float64
		}
		if approvedDate.Valid {
			ap.ApprovedDate = &approvedDate.Time
		}
		if approvedBy.Valid {
			ap.ApprovedBy = &approvedBy.String
		}

		approvals = append(approvals, ap)
	}

	return approvals, totalCount, nil
}

package services

import (
	"candidate-backend/internal/models"
	"database/sql"
	"fmt"
)

type ChangeLogService struct {
	db *sql.DB
}

func NewChangeLogService(db *sql.DB) *ChangeLogService {
	return &ChangeLogService{db: db}
}

// GetTaskLogs retrieves all change logs for a task
func (s *ChangeLogService) GetTaskLogs(taskID string) ([]models.ChangeLog, error) {
	rows, err := s.db.Query(`
		SELECT cl.id, cl.task_id, cl.user_id, u.name as user_name,
		       cl.action, cl.details, cl.created_at
		FROM change_logs cl
		JOIN users u ON cl.user_id = u.id
		WHERE cl.task_id = $1
		ORDER BY cl.created_at DESC
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.ChangeLog
	for rows.Next() {
		var log models.ChangeLog
		err := rows.Scan(
			&log.ID, &log.TaskID, &log.UserID, &log.UserName,
			&log.Action, &log.Details, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	if logs == nil {
		logs = []models.ChangeLog{}
	}

	return logs, nil
}

// CreateChangeLog creates a new change log entry
func (s *ChangeLogService) CreateChangeLog(taskID, userID int, action, details string) error {
	_, err := s.db.Exec(
		"INSERT INTO change_logs (task_id, user_id, action, details) VALUES ($1, $2, $3, $4)",
		taskID, userID, action, details,
	)
	return err
}

// FormatChangeDetails formats multiple changes into a readable string
func (s *ChangeLogService) FormatChangeDetails(changes []string) string {
	if len(changes) == 0 {
		return ""
	}

	if len(changes) == 1 {
		return changes[0]
	}

	if len(changes) == 2 {
		return fmt.Sprintf("%s and %s", changes[0], changes[1])
	}

	result := changes[0]
	for i := 1; i < len(changes)-1; i++ {
		result += fmt.Sprintf(", %s", changes[i])
	}
	result += fmt.Sprintf(", and %s", changes[len(changes)-1])

	return result
}

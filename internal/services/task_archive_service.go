package services

import (
	"candidate-backend/internal/models"
	"database/sql"
	"fmt"
)

type TaskArchiveService struct {
	db *sql.DB
}

func NewTaskArchiveService(db *sql.DB) *TaskArchiveService {
	return &TaskArchiveService{db: db}
}

// ArchiveTask archives a task
func (s *TaskArchiveService) ArchiveTask(taskID string, userID int) (*models.Task, string, error) {
	// Check ownership
	var creatorID int
	var title string
	err := s.db.QueryRow("SELECT creator_id, title FROM tasks WHERE id = $1", taskID).Scan(&creatorID, &title)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", fmt.Errorf("task not found")
		}
		return nil, "", err
	}

	if creatorID != userID {
		return nil, "", fmt.Errorf("you can only archive your own tasks")
	}

	var task models.Task
	err = s.db.QueryRow(`
		UPDATE tasks
		SET archived = TRUE, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, title, description, status, creator_id, due_date, archived, created_at, updated_at
	`, taskID).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status,
		&task.CreatorID, &task.DueDate, &task.Archived, &task.CreatedAt, &task.UpdatedAt,
	)

	if err != nil {
		return nil, "", err
	}

	return &task, title, nil
}

// UnarchiveTask restores an archived task
func (s *TaskArchiveService) UnarchiveTask(taskID string, userID int) (*models.Task, string, error) {
	// Check ownership
	var creatorID int
	var title string
	err := s.db.QueryRow("SELECT creator_id, title FROM tasks WHERE id = $1", taskID).Scan(&creatorID, &title)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", fmt.Errorf("task not found")
		}
		return nil, "", err
	}

	if creatorID != userID {
		return nil, "", fmt.Errorf("you can only unarchive your own tasks")
	}

	var task models.Task
	err = s.db.QueryRow(`
		UPDATE tasks
		SET archived = FALSE, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, title, description, status, creator_id, due_date, archived, created_at, updated_at
	`, taskID).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status,
		&task.CreatorID, &task.DueDate, &task.Archived, &task.CreatedAt, &task.UpdatedAt,
	)

	if err != nil {
		return nil, "", err
	}

	return &task, title, nil
}

// GetArchivedTasks retrieves archived tasks with pagination
func (s *TaskArchiveService) GetArchivedTasks(limit, offset int) ([]models.Task, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.title, t.description, t.status, t.creator_id,
		       u.name as creator_name, t.due_date, t.archived, t.created_at, t.updated_at
		FROM tasks t
		JOIN users u ON t.creator_id = u.id
		WHERE t.archived = TRUE
		ORDER BY t.updated_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Status,
			&task.CreatorID, &task.CreatorName, &task.DueDate, &task.Archived,
			&task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	return tasks, nil
}

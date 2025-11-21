package services

import (
	"candidate-backend/internal/models"
	"candidate-backend/internal/validators"
	"database/sql"
	"fmt"
)

type TaskService struct {
	db        *sql.DB
	validator *validators.TaskValidator
}

func NewTaskService(db *sql.DB) *TaskService {
	return &TaskService{
		db:        db,
		validator: validators.NewTaskValidator(),
	}
}

// GetTasks retrieves non-archived tasks with pagination
func (s *TaskService) GetTasks(limit, offset int) ([]models.Task, error) {
	if err := s.validator.ValidatePagination(limit, offset); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(`
		SELECT t.id, t.title, t.description, t.status, t.creator_id,
		       u.name as creator_name, t.due_date, t.archived, t.created_at, t.updated_at
		FROM tasks t
		JOIN users u ON t.creator_id = u.id
		WHERE t.archived = FALSE
		ORDER BY t.created_at DESC
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

// GetTask retrieves a single task by ID
func (s *TaskService) GetTask(taskID string) (*models.Task, error) {
	var task models.Task
	err := s.db.QueryRow(`
		SELECT t.id, t.title, t.description, t.status, t.creator_id,
		       u.name as creator_name, t.due_date, t.archived, t.created_at, t.updated_at
		FROM tasks t
		JOIN users u ON t.creator_id = u.id
		WHERE t.id = $1
	`, taskID).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status,
		&task.CreatorID, &task.CreatorName, &task.DueDate, &task.Archived,
		&task.CreatedAt, &task.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &task, nil
}

// CreateTask creates a new task
func (s *TaskService) CreateTask(req models.CreateTaskRequest, userID int) (*models.Task, error) {
	if err := s.validator.ValidateCreateTask(&req); err != nil {
		return nil, err
	}

	// Set default status if not provided
	if req.Status == "" {
		req.Status = models.StatusToDo
	}

	var task models.Task
	err := s.db.QueryRow(`
		INSERT INTO tasks (title, description, status, creator_id, due_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, title, description, status, creator_id, due_date, archived, created_at, updated_at
	`, req.Title, req.Description, req.Status, userID, req.DueDate).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status,
		&task.CreatorID, &task.DueDate, &task.Archived, &task.CreatedAt, &task.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &task, nil
}

// UpdateTask updates an existing task
func (s *TaskService) UpdateTask(taskID string, req models.UpdateTaskRequest, userID int) (*models.Task, []string, error) {
	if err := s.validator.ValidateUpdateTask(&req); err != nil {
		return nil, nil, err
	}

	// Check ownership
	if err := s.CheckTaskOwnership(taskID, userID); err != nil {
		return nil, nil, err
	}

	// Build dynamic update query
	query := "UPDATE tasks SET "
	args := []interface{}{}
	argCount := 1
	changes := []string{}

	if req.Title != nil {
		query += fmt.Sprintf("title = $%d, ", argCount)
		args = append(args, *req.Title)
		argCount++
		changes = append(changes, fmt.Sprintf("changed title to '%s'", *req.Title))
	}
	if req.Description != nil {
		query += fmt.Sprintf("description = $%d, ", argCount)
		args = append(args, *req.Description)
		argCount++
		changes = append(changes, "updated description")
	}
	if req.Status != nil {
		query += fmt.Sprintf("status = $%d, ", argCount)
		args = append(args, *req.Status)
		argCount++
		changes = append(changes, fmt.Sprintf("changed status to '%s'", *req.Status))
	}
	if req.DueDate != nil {
		query += fmt.Sprintf("due_date = $%d, ", argCount)
		args = append(args, req.DueDate)
		argCount++
	}

	if len(args) == 0 {
		return nil, nil, fmt.Errorf("no fields to update")
	}

	query += fmt.Sprintf("updated_at = CURRENT_TIMESTAMP WHERE id = $%d", argCount)
	args = append(args, taskID)

	query += " RETURNING id, title, description, status, creator_id, due_date, archived, created_at, updated_at"

	var task models.Task
	err := s.db.QueryRow(query, args...).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status,
		&task.CreatorID, &task.DueDate, &task.Archived, &task.CreatedAt, &task.UpdatedAt,
	)

	if err != nil {
		return nil, nil, err
	}

	return &task, changes, nil
}

// DeleteTask deletes a task
func (s *TaskService) DeleteTask(taskID string, userID int) (string, error) {
	// Check ownership and get title
	var creatorID int
	var title string
	err := s.db.QueryRow("SELECT creator_id, title FROM tasks WHERE id = $1", taskID).Scan(&creatorID, &title)
	if err != nil {
		return "", err
	}

	if creatorID != userID {
		return "", fmt.Errorf("you can only delete your own tasks")
	}

	_, err = s.db.Exec("DELETE FROM tasks WHERE id = $1", taskID)
	if err != nil {
		return "", err
	}

	return title, nil
}

// CheckTaskOwnership checks if user owns the task
func (s *TaskService) CheckTaskOwnership(taskID string, userID int) error {
	var creatorID int
	err := s.db.QueryRow("SELECT creator_id FROM tasks WHERE id = $1", taskID).Scan(&creatorID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("task not found")
		}
		return err
	}

	if creatorID != userID {
		return fmt.Errorf("you can only modify your own tasks")
	}

	return nil
}

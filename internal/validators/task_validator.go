package validators

import (
	"candidate-backend/internal/models"
	"errors"
	"strings"
)

type TaskValidator struct{}

func NewTaskValidator() *TaskValidator {
	return &TaskValidator{}
}

// ValidateCreateTask validates task creation request
func (v *TaskValidator) ValidateCreateTask(req *models.CreateTaskRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return errors.New("title is required")
	}

	if len(req.Title) > 500 {
		return errors.New("title must be less than 500 characters")
	}

	if req.Status != "" {
		if err := v.ValidateStatus(req.Status); err != nil {
			return err
		}
	}

	return nil
}

// ValidateUpdateTask validates task update request
func (v *TaskValidator) ValidateUpdateTask(req *models.UpdateTaskRequest) error {
	if req.Title != nil {
		if strings.TrimSpace(*req.Title) == "" {
			return errors.New("title cannot be empty")
		}
		if len(*req.Title) > 500 {
			return errors.New("title must be less than 500 characters")
		}
	}

	if req.Status != nil {
		if err := v.ValidateStatus(*req.Status); err != nil {
			return err
		}
	}

	return nil
}

// ValidateStatus validates task status
func (v *TaskValidator) ValidateStatus(status models.TaskStatus) error {
	validStatuses := map[models.TaskStatus]bool{
		models.StatusToDo:       true,
		models.StatusInProgress: true,
		models.StatusDone:       true,
	}

	if !validStatuses[status] {
		return errors.New("invalid status: must be 'To Do', 'In Progress', or 'Done'")
	}

	return nil
}

// ValidatePagination validates pagination parameters
func (v *TaskValidator) ValidatePagination(limit, offset int) error {
	if limit < 1 || limit > 100 {
		return errors.New("limit must be between 1 and 100")
	}

	if offset < 0 {
		return errors.New("offset must be >= 0")
	}

	return nil
}

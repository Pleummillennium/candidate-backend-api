package validators

import (
	"candidate-backend/internal/models"
	"testing"
)

func TestValidateCreateTask(t *testing.T) {
	validator := NewTaskValidator()

	tests := []struct {
		name    string
		req     models.CreateTaskRequest
		wantErr bool
	}{
		{
			name: "Valid task",
			req: models.CreateTaskRequest{
				Title:       "Test Task",
				Description: "Test Description",
				Status:      models.StatusToDo,
			},
			wantErr: false,
		},
		{
			name: "Empty title",
			req: models.CreateTaskRequest{
				Title:       "",
				Description: "Test Description",
			},
			wantErr: true,
		},
		{
			name: "Title too long",
			req: models.CreateTaskRequest{
				Title: string(make([]byte, 501)),
			},
			wantErr: true,
		},
		{
			name: "Invalid status",
			req: models.CreateTaskRequest{
				Title:  "Test",
				Status: "Invalid Status",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCreateTask(&tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreateTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStatus(t *testing.T) {
	validator := NewTaskValidator()

	tests := []struct {
		name    string
		status  models.TaskStatus
		wantErr bool
	}{
		{"Valid To Do", models.StatusToDo, false},
		{"Valid In Progress", models.StatusInProgress, false},
		{"Valid Done", models.StatusDone, false},
		{"Invalid Status", "Invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateStatus(tt.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePagination(t *testing.T) {
	validator := NewTaskValidator()

	tests := []struct {
		name    string
		limit   int
		offset  int
		wantErr bool
	}{
		{"Valid pagination", 10, 0, false},
		{"Valid limit 1", 1, 0, false},
		{"Valid limit 100", 100, 0, false},
		{"Invalid limit 0", 0, 0, true},
		{"Invalid limit > 100", 101, 0, true},
		{"Invalid offset", 10, -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePagination(tt.limit, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePagination() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

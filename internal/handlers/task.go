package handlers

import (
	"candidate-backend/internal/middleware"
	"candidate-backend/internal/models"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	db *sql.DB
}

func NewTaskHandler(db *sql.DB) *TaskHandler {
	return &TaskHandler{db: db}
}

// GetTasks godoc
// @Summary      Get all tasks
// @Description  Retrieve all tasks with creator information
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {array}   models.Task
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/tasks [get]
func (h *TaskHandler) GetTasks(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT t.id, t.title, t.description, t.status, t.creator_id,
		       u.name as creator_name, t.due_date, t.created_at, t.updated_at
		FROM tasks t
		JOIN users u ON t.creator_id = u.id
		ORDER BY t.created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Status,
			&task.CreatorID, &task.CreatorName, &task.DueDate,
			&task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan task"})
			return
		}
		tasks = append(tasks, task)
	}

	if tasks == nil {
		tasks = []models.Task{}
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")

	var task models.Task
	err := h.db.QueryRow(`
		SELECT t.id, t.title, t.description, t.status, t.creator_id,
		       u.name as creator_name, t.due_date, t.created_at, t.updated_at
		FROM tasks t
		JOIN users u ON t.creator_id = u.id
		WHERE t.id = $1
	`, taskID).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status,
		&task.CreatorID, &task.CreatorName, &task.DueDate,
		&task.CreatedAt, &task.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req models.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default status if not provided
	if req.Status == "" {
		req.Status = models.StatusToDo
	}

	var task models.Task
	err := h.db.QueryRow(`
		INSERT INTO tasks (title, description, status, creator_id, due_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, title, description, status, creator_id, due_date, created_at, updated_at
	`, req.Title, req.Description, req.Status, userID, req.DueDate).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status,
		&task.CreatorID, &task.DueDate, &task.CreatedAt, &task.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	// Log the creation
	h.createChangeLog(task.ID, userID, "created", fmt.Sprintf("Created task: %s", task.Title))

	c.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	taskID := c.Param("id")
	userID, _ := middleware.GetUserID(c)

	// Check if user is the creator
	var creatorID int
	err := h.db.QueryRow("SELECT creator_id FROM tasks WHERE id = $1", taskID).Scan(&creatorID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if creatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own tasks"})
		return
	}

	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
		changes = append(changes, fmt.Sprintf("title to '%s'", *req.Title))
	}
	if req.Description != nil {
		query += fmt.Sprintf("description = $%d, ", argCount)
		args = append(args, *req.Description)
		argCount++
	}
	if req.Status != nil {
		query += fmt.Sprintf("status = $%d, ", argCount)
		args = append(args, *req.Status)
		argCount++
		changes = append(changes, fmt.Sprintf("status to '%s'", *req.Status))
	}
	if req.DueDate != nil {
		query += fmt.Sprintf("due_date = $%d, ", argCount)
		args = append(args, req.DueDate)
		argCount++
	}

	if len(args) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	query += fmt.Sprintf("updated_at = CURRENT_TIMESTAMP WHERE id = $%d", argCount)
	args = append(args, taskID)

	query += " RETURNING id, title, description, status, creator_id, due_date, created_at, updated_at"

	var task models.Task
	err = h.db.QueryRow(query, args...).Scan(
		&task.ID, &task.Title, &task.Description, &task.Status,
		&task.CreatorID, &task.DueDate, &task.CreatedAt, &task.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	// Log the update
	if len(changes) > 0 {
		changeDetails := fmt.Sprintf("Updated %s", changes[0])
		for i := 1; i < len(changes); i++ {
			changeDetails += fmt.Sprintf(", %s", changes[i])
		}
		h.createChangeLog(task.ID, userID, "updated", changeDetails)
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	userID, _ := middleware.GetUserID(c)

	// Check if user is the creator
	var creatorID int
	var title string
	err := h.db.QueryRow("SELECT creator_id, title FROM tasks WHERE id = $1", taskID).Scan(&creatorID, &title)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if creatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own tasks"})
		return
	}

	// Log before deletion (because of CASCADE)
	h.createChangeLog(atoi(taskID), userID, "deleted", fmt.Sprintf("Deleted task: %s", title))

	_, err = h.db.Exec("DELETE FROM tasks WHERE id = $1", taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

func (h *TaskHandler) GetTaskLogs(c *gin.Context) {
	taskID := c.Param("id")

	rows, err := h.db.Query(`
		SELECT cl.id, cl.task_id, cl.user_id, u.name as user_name,
		       cl.action, cl.details, cl.created_at
		FROM change_logs cl
		JOIN users u ON cl.user_id = u.id
		WHERE cl.task_id = $1
		ORDER BY cl.created_at DESC
	`, taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
		return
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan log"})
			return
		}
		logs = append(logs, log)
	}

	if logs == nil {
		logs = []models.ChangeLog{}
	}

	c.JSON(http.StatusOK, logs)
}

func (h *TaskHandler) createChangeLog(taskID, userID int, action, details string) {
	h.db.Exec(
		"INSERT INTO change_logs (task_id, user_id, action, details) VALUES ($1, $2, $3, $4)",
		taskID, userID, action, details,
	)
}

func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

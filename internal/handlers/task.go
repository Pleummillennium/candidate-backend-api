package handlers

import (
	"candidate-backend/internal/middleware"
	"candidate-backend/internal/models"
	"candidate-backend/internal/services"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	taskService      *services.TaskService
	archiveService   *services.TaskArchiveService
	changeLogService *services.ChangeLogService
}

func NewTaskHandler(db *sql.DB) *TaskHandler {
	return &TaskHandler{
		taskService:      services.NewTaskService(db),
		archiveService:   services.NewTaskArchiveService(db),
		changeLogService: services.NewChangeLogService(db),
	}
}

// GetTasks godoc
// @Summary      Get all tasks
// @Description  Retrieve all non-archived tasks with creator information (supports pagination)
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        limit   query     int  false  "Limit number of results (default: 10)"
// @Param        offset  query     int  false  "Offset for pagination (default: 0)"
// @Success      200  {array}   models.Task
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/tasks [get]
func (h *TaskHandler) GetTasks(c *gin.Context) {
	// Get pagination params
	limit := 10
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	tasks, err := h.taskService.GetTasks(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// GetTask godoc
// @Summary      Get task by ID
// @Description  Retrieve a specific task by its ID
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      int  true  "Task ID"
// @Success      200  {object}  models.Task
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.taskService.GetTask(taskID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// CreateTask godoc
// @Summary      Create a new task
// @Description  Create a new task with title, description, and status
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        task  body      models.CreateTaskRequest  true  "Task data"
// @Success      201   {object}  models.Task
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Status      string `json:"status"`
		DueDate     *string `json:"due_date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to service request
	createReq := models.CreateTaskRequest{
		Title:       req.Title,
		Description: req.Description,
		Status:      models.TaskStatus(req.Status),
	}

	task, err := h.taskService.CreateTask(createReq, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log the creation
	_ = h.changeLogService.CreateChangeLog(task.ID, userID, "created", fmt.Sprintf("Created task: %s", task.Title))

	c.JSON(http.StatusCreated, task)
}

// UpdateTask godoc
// @Summary      Update a task
// @Description  Update task information (only the creator can update)
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id    path      int                      true  "Task ID"
// @Param        task  body      models.UpdateTaskRequest  true  "Updated task data"
// @Success      200   {object}  models.Task
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	taskID := c.Param("id")
	userID, _ := middleware.GetUserID(c)

	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, changes, err := h.taskService.UpdateTask(taskID, req, userID)
	if err != nil {
		if err.Error() == "task not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "you can only modify your own tasks" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log the update
	if len(changes) > 0 {
		changeDetails := h.changeLogService.FormatChangeDetails(changes)
		_ = h.changeLogService.CreateChangeLog(task.ID, userID, "updated", changeDetails)
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTask godoc
// @Summary      Delete a task
// @Description  Delete a task (only the creator can delete)
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      int  true  "Task ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	userID, _ := middleware.GetUserID(c)

	taskIDInt, _ := strconv.Atoi(taskID)

	title, err := h.taskService.DeleteTask(taskID, userID)
	if err != nil {
		if err == sql.ErrNoRows || err.Error() == "task not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		if err.Error() == "you can only delete your own tasks" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log before deletion (because of CASCADE)
	_ = h.changeLogService.CreateChangeLog(taskIDInt, userID, "deleted", fmt.Sprintf("Deleted task: %s", title))

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// ArchiveTask godoc
// @Summary      Archive a task
// @Description  Archive a task (only the creator can archive)
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      int  true  "Task ID"
// @Success      200  {object}  models.Task
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/tasks/{id}/archive [post]
func (h *TaskHandler) ArchiveTask(c *gin.Context) {
	taskID := c.Param("id")
	userID, _ := middleware.GetUserID(c)

	task, title, err := h.archiveService.ArchiveTask(taskID, userID)
	if err != nil {
		if err.Error() == "task not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "you can only archive your own tasks" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log the archive action
	h.changeLogService.CreateChangeLog(task.ID, userID, "archived", fmt.Sprintf("Archived task: %s", title))

	c.JSON(http.StatusOK, task)
}

// UnarchiveTask godoc
// @Summary      Unarchive a task
// @Description  Restore an archived task (only the creator can unarchive)
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      int  true  "Task ID"
// @Success      200  {object}  models.Task
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/tasks/{id}/unarchive [post]
func (h *TaskHandler) UnarchiveTask(c *gin.Context) {
	taskID := c.Param("id")
	userID, _ := middleware.GetUserID(c)

	task, title, err := h.archiveService.UnarchiveTask(taskID, userID)
	if err != nil {
		if err.Error() == "task not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "you can only unarchive your own tasks" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log the unarchive action
	h.changeLogService.CreateChangeLog(task.ID, userID, "unarchived", fmt.Sprintf("Restored task: %s", title))

	c.JSON(http.StatusOK, task)
}

// GetArchivedTasks godoc
// @Summary      Get archived tasks
// @Description  Retrieve all archived tasks with creator information (supports pagination)
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        limit   query     int  false  "Limit number of results (default: 10)"
// @Param        offset  query     int  false  "Offset for pagination (default: 0)"
// @Success      200  {array}   models.Task
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/tasks/archived [get]
func (h *TaskHandler) GetArchivedTasks(c *gin.Context) {
	// Get pagination params
	limit := 10
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	tasks, err := h.archiveService.GetArchivedTasks(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// GetTaskLogs godoc
// @Summary      Get task change logs
// @Description  Retrieve all change logs for a specific task
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      int  true  "Task ID"
// @Success      200  {array}   models.ChangeLog
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/tasks/{id}/logs [get]
func (h *TaskHandler) GetTaskLogs(c *gin.Context) {
	taskID := c.Param("id")

	logs, err := h.changeLogService.GetTaskLogs(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

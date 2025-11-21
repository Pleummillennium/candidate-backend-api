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

type CommentHandler struct {
	db *sql.DB
}

func NewCommentHandler(db *sql.DB) *CommentHandler {
	return &CommentHandler{db: db}
}

// GetComments godoc
// @Summary      Get task comments
// @Description  Retrieve all comments for a specific task
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      int  true  "Task ID"
// @Success      200  {array}   models.Comment
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/tasks/{id}/comments [get]
func (h *CommentHandler) GetComments(c *gin.Context) {
	taskID := c.Param("id")

	// Check if task exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM tasks WHERE id = $1)", taskID).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	rows, err := h.db.Query(`
		SELECT c.id, c.task_id, c.user_id, u.name as user_name,
		       c.content, c.created_at, c.updated_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.task_id = $1
		ORDER BY c.created_at ASC
	`, taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(
			&comment.ID, &comment.TaskID, &comment.UserID, &comment.UserName,
			&comment.Content, &comment.CreatedAt, &comment.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan comment"})
			return
		}
		comments = append(comments, comment)
	}

	if comments == nil {
		comments = []models.Comment{}
	}

	c.JSON(http.StatusOK, comments)
}

// CreateComment godoc
// @Summary      Create a comment
// @Description  Add a new comment to a task
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id       path      int                          true  "Task ID"
// @Param        comment  body      models.CreateCommentRequest  true  "Comment data"
// @Success      201      {object}  models.Comment
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/tasks/{id}/comments [post]
func (h *CommentHandler) CreateComment(c *gin.Context) {
	taskID := c.Param("id")
	userID, _ := middleware.GetUserID(c)

	// Check if task exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM tasks WHERE id = $1)", taskID).Scan(&exists)
	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	var req models.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var comment models.Comment
	err = h.db.QueryRow(`
		INSERT INTO comments (task_id, user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, task_id, user_id, content, created_at, updated_at
	`, taskID, userID, req.Content).Scan(
		&comment.ID, &comment.TaskID, &comment.UserID,
		&comment.Content, &comment.CreatedAt, &comment.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	// Log the comment creation
	taskIDInt, _ := strconv.Atoi(taskID)
	h.createChangeLog(taskIDInt, userID, "commented", fmt.Sprintf("Added a comment"))

	c.JSON(http.StatusCreated, comment)
}

// UpdateComment godoc
// @Summary      Update a comment
// @Description  Update comment content (only the comment creator can update)
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id       path      int                          true  "Comment ID"
// @Param        comment  body      models.UpdateCommentRequest  true  "Updated comment data"
// @Success      200      {object}  models.Comment
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/comments/{id} [put]
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	commentID := c.Param("id")
	userID, _ := middleware.GetUserID(c)

	// Check if comment exists and user is the owner
	var comment models.Comment
	err := h.db.QueryRow(`
		SELECT id, task_id, user_id, content, created_at, updated_at
		FROM comments WHERE id = $1
	`, commentID).Scan(
		&comment.ID, &comment.TaskID, &comment.UserID,
		&comment.Content, &comment.CreatedAt, &comment.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check ownership - only the comment creator can update it
	if comment.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own comments"})
		return
	}

	var req models.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.db.QueryRow(`
		UPDATE comments
		SET content = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING id, task_id, user_id, content, created_at, updated_at
	`, req.Content, commentID).Scan(
		&comment.ID, &comment.TaskID, &comment.UserID,
		&comment.Content, &comment.CreatedAt, &comment.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment"})
		return
	}

	// Log the comment update
	h.createChangeLog(comment.TaskID, userID, "updated_comment", "Updated a comment")

	c.JSON(http.StatusOK, comment)
}

// DeleteComment godoc
// @Summary      Delete a comment
// @Description  Delete a comment (only the comment creator can delete)
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        id   path      int  true  "Comment ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/comments/{id} [delete]
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	commentID := c.Param("id")
	userID, _ := middleware.GetUserID(c)

	// Check if comment exists and user is the owner
	var taskID, ownerID int
	err := h.db.QueryRow(`
		SELECT task_id, user_id FROM comments WHERE id = $1
	`, commentID).Scan(&taskID, &ownerID)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check ownership - only the comment creator can delete it
	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own comments"})
		return
	}

	// Log before deletion
	h.createChangeLog(taskID, userID, "deleted_comment", "Deleted a comment")

	_, err = h.db.Exec("DELETE FROM comments WHERE id = $1", commentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}

func (h *CommentHandler) createChangeLog(taskID, userID int, action, details string) {
	h.db.Exec(
		"INSERT INTO change_logs (task_id, user_id, action, details) VALUES ($1, $2, $3, $4)",
		taskID, userID, action, details,
	)
}

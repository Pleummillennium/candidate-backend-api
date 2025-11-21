package handlers

// This file contains additional Swagger annotations for documentation
// The actual implementations are in the respective handler files

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
// @Failure      403   {object}  map[string]string  "Not the task owner"
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /api/tasks/{id} [put]

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
// @Failure      403  {object}  map[string]string  "Not the task owner"
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/tasks/{id} [delete]

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
// @Failure      403      {object}  map[string]string  "Not the comment owner"
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/comments/{id} [put]

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
// @Failure      403  {object}  map[string]string  "Not the comment owner"
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /api/comments/{id} [delete]

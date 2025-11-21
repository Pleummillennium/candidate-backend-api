package validators

import (
	"candidate-backend/internal/models"
	"errors"
	"strings"
)

type CommentValidator struct{}

func NewCommentValidator() *CommentValidator {
	return &CommentValidator{}
}

// ValidateCreateComment validates comment creation request
func (v *CommentValidator) ValidateCreateComment(req *models.CreateCommentRequest) error {
	if strings.TrimSpace(req.Content) == "" {
		return errors.New("comment content is required")
	}

	if len(req.Content) > 5000 {
		return errors.New("comment must be less than 5000 characters")
	}

	return nil
}

// ValidateUpdateComment validates comment update request
func (v *CommentValidator) ValidateUpdateComment(req *models.UpdateCommentRequest) error {
	if strings.TrimSpace(req.Content) == "" {
		return errors.New("comment content is required")
	}

	if len(req.Content) > 5000 {
		return errors.New("comment must be less than 5000 characters")
	}

	return nil
}

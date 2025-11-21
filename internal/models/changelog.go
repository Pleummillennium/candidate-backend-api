package models

import "time"

type ChangeLog struct {
	ID        int       `json:"id"`
	TaskID    int       `json:"task_id"`
	UserID    int       `json:"user_id"`
	UserName  string    `json:"user_name,omitempty"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}

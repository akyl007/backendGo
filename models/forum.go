package models

import "time"

type ForumMessage struct {
    ID        int       `json:"id"`
    UserID    int       `json:"user_id"`
    Username  string    `json:"username"`
    UserRole  string    `json:"user_role"`
    Message   string    `json:"message"`
    CreatedAt time.Time `json:"created_at"`
}

type CreateForumMessageRequest struct {
    Message string `json:"message"`
} 
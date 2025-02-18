package handlers

import (
	"asii/db"
	"asii/models"
	"asii/utils"
	"encoding/json"
	"log"
	"net/http"
)

func GetForumMessages(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
        SELECT f.id, f.user_id, u.username, u.role, f.message, f.created_at
        FROM forum_messages f
        JOIN users u ON f.user_id = u.id
        ORDER BY f.created_at DESC
    `)
	if err != nil {
		log.Printf("Error querying forum messages: %v", err)
		http.Error(w, "Error fetching forum messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []models.ForumMessage
	for rows.Next() {
		var msg models.ForumMessage
		err := rows.Scan(&msg.ID, &msg.UserID, &msg.Username, &msg.UserRole, &msg.Message, &msg.CreatedAt)
		if err != nil {
			log.Printf("Error scanning forum message: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func CreateForumMessage(w http.ResponseWriter, r *http.Request) {
	claims := utils.UserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CreateForumMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, "Message cannot be empty", http.StatusBadRequest)
		return
	}

	var messageID int
	err := db.DB.QueryRow(`
        INSERT INTO forum_messages (user_id, message, created_at)
        VALUES ($1, $2, CURRENT_TIMESTAMP)
        RETURNING id`,
		claims.UserID, req.Message).Scan(&messageID)

	if err != nil {
		log.Printf("Error creating forum message: %v", err)
		http.Error(w, "Error creating message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Message created successfully",
		"id":      messageID,
	})
}

func DeleteForumMessage(w http.ResponseWriter, r *http.Request) {
	claims := utils.UserFromContext(r.Context())
	messageID := r.URL.Query().Get("id")

	var canDelete bool
	var err error

	if claims.Role == "admin" {
		canDelete = true
	} else {
		err = db.DB.QueryRow(`
            SELECT EXISTS(
                SELECT 1 FROM forum_messages 
                WHERE id = $1 AND user_id = $2
            )`, messageID, claims.UserID).Scan(&canDelete)
	}

	if err != nil || !canDelete {
		http.Error(w, "Not authorized to delete this message", http.StatusForbidden)
		return
	}

	_, err = db.DB.Exec("DELETE FROM forum_messages WHERE id = $1", messageID)
	if err != nil {
		http.Error(w, "Error deleting message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Message deleted successfully",
	})
}

package handlers

import (
	"asii/db"
	"asii/models"
	"asii/utils"
	"encoding/json"
	"log"
	"net/http"
)

func CreateFeedback(w http.ResponseWriter, r *http.Request) {

	claims := utils.UserFromContext(r.Context())
	if claims.Role != "student" {
		http.Error(w, "Only students can provide feedback", http.StatusForbidden)
		return
	}

	var req models.FeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding feedback request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var courseExists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM courses WHERE id = $1)", req.CourseID).Scan(&courseExists)
	if err != nil || !courseExists {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	var teacherExists bool
	err = db.DB.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM users u 
            JOIN courses c ON c.teacher_id = u.id 
            WHERE u.id = $1 AND u.role = 'teacher' AND c.id = $2
        )`, req.TeacherID, req.CourseID).Scan(&teacherExists)
	if err != nil || !teacherExists {
		http.Error(w, "Teacher not found or not assigned to this course", http.StatusBadRequest)
		return
	}

	var isEnrolled bool
	err = db.DB.QueryRow(`
        SELECT EXISTS(
            SELECT 1 FROM course_students 
            WHERE course_id = $1 AND student_id = $2
        )`, req.CourseID, claims.UserID).Scan(&isEnrolled)
	if err != nil || !isEnrolled {
		http.Error(w, "Student is not enrolled in this course", http.StatusForbidden)
		return
	}

	_, err = db.DB.Exec(`
        INSERT INTO feedback (course_id, student_id, teacher_id, comment)
        VALUES ($1, $2, $3, $4)`,
		req.CourseID, claims.UserID, req.TeacherID, req.Comment)

	if err != nil {
		log.Printf("Error creating feedback: %v", err)
		http.Error(w, "Error saving feedback", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Feedback successfully sent",
	})
}

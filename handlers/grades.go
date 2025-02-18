package handlers

import (
	"asii/db"
	"encoding/json"
	"net/http"
)

type GradeRequest struct {
	LessonID  int `json:"lesson_id"`
	StudentID int `json:"student_id"`
	Grade     int `json:"grade"`
}

func AssignGrade(w http.ResponseWriter, r *http.Request) {
	var req GradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Grade < 0 || req.Grade > 100 {
		http.Error(w, "Grade must be between 0 and 100", http.StatusBadRequest)
		return
	}

	result, err := db.DB.Exec(`
		UPDATE lesson_attendance 
		SET grade = $1 
		WHERE lesson_id = $2 AND student_id = $3`,
		req.Grade, req.LessonID, req.StudentID)

	if err != nil {
		http.Error(w, "Error assigning grade", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Student not found in this lesson", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Grade assigned successfully"})
}

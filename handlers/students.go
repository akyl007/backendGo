package handlers

import (
	"asii/db"
	"asii/models"
	"encoding/json"
	"net/http"
)

func GetStudents(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, username FROM users WHERE role = 'student'")
	if err != nil {
		http.Error(w, "Ошибка загрузки студентов", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var student models.Student
		if err := rows.Scan(&student.ID, &student.Username); err != nil {
			http.Error(w, "Ошибка обработки данных", http.StatusInternalServerError)
			return
		}
		students = append(students, student)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(students)
}

func AssignStudents(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CourseID   int   `json:"course_id"`
		StudentIDs []int `json:"student_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return
	}

	for _, studentID := range req.StudentIDs {
		_, err := tx.Exec(`
			INSERT INTO course_students (course_id, student_id) 
			VALUES ($1, $2) 
			ON CONFLICT (course_id, student_id) DO NOTHING`,
			req.CourseID, studentID)
		if err != nil {
			tx.Rollback()
			http.Error(w, "Ошибка назначения студента", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Students assigned successfully"})
}

func UpdateAttendance(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LessonID   int  `json:"lesson_id"`
		StudentID  int  `json:"student_id"`
		Attendance bool `json:"attendance"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	result, err := db.DB.Exec(`
		INSERT INTO lesson_attendance (lesson_id, student_id, attendance)
		VALUES ($1, $2, $3)
		ON CONFLICT (lesson_id, student_id) 
		DO UPDATE SET attendance = $3`,
		req.LessonID, req.StudentID, req.Attendance)

	if err != nil {
		http.Error(w, "Error updating attendance", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Student not found in this lesson", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Attendance updated successfully"})
}

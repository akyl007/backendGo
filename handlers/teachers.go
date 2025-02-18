package handlers

import (
	"asii/db"
	"asii/models"
	"encoding/json"
	"net/http"
)

func GetTeachers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT id, username FROM users WHERE role = 'teacher'")
	if err != nil {
		http.Error(w, "Ошибка загрузки преподавателей", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var teachers []models.Teacher
	for rows.Next() {
		var teacher models.Teacher
		if err := rows.Scan(&teacher.ID, &teacher.Username); err != nil {
			http.Error(w, "Ошибка обработки данных", http.StatusInternalServerError)
			return
		}
		teachers = append(teachers, teacher)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teachers)
}

func AssignTeacher(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CourseID  int `json:"course_id"`
		TeacherID int `json:"teacher_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный запрос", http.StatusBadRequest)
		return
	}

	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND role = 'teacher')",
		req.TeacherID).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "Преподаватель не найден", http.StatusNotFound)
		return
	}

	_, err = db.DB.Exec("UPDATE courses SET teacher_id = $1 WHERE id = $2",
		req.TeacherID, req.CourseID)
	if err != nil {
		http.Error(w, "Ошибка назначения преподавателя", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Преподаватель успешно назначен"})
}

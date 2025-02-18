package handlers

import (
	"asii/db"
	"asii/models"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func DeleteLesson(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID урока обязателен", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный ID урока", http.StatusBadRequest)
		return
	}

	_, err = db.DB.Exec("DELETE FROM lesson_students WHERE lesson_id = $1", id)
	if err != nil {
		http.Error(w, "Ошибка удаления связей урока", http.StatusInternalServerError)
		return
	}

	result, err := db.DB.Exec("DELETE FROM lessons WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Ошибка удаления урока", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Ошибка получения результата", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Урок не найден", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Урок успешно удален"})
}

func GetLessons(w http.ResponseWriter, r *http.Request) {
	courseID := r.URL.Query().Get("course_id")
	if courseID == "" {
		http.Error(w, "Course ID is required", http.StatusBadRequest)
		return
	}

	query := `
		SELECT 
			l.id,
			l.name,
			l.description,
			l.course_id
		FROM lessons l
		WHERE l.course_id = $1
		ORDER BY l.id
	`

	rows, err := db.DB.Query(query, courseID)
	if err != nil {
		log.Printf("Error querying lessons: %v", err)
		http.Error(w, "Error loading lessons", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var lessons []models.Lesson
	for rows.Next() {
		var lesson models.Lesson
		err := rows.Scan(
			&lesson.ID,
			&lesson.Name,
			&lesson.Description,
			&lesson.CourseID,
		)

		if err != nil {
			log.Printf("Error scanning lesson: %v", err)
			http.Error(w, "Error processing lesson data", http.StatusInternalServerError)
			return
		}

		lessons = append(lessons, lesson)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating lessons: %v", err)
		http.Error(w, "Error processing lessons", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lessons)
}

func GetLesson(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID урока обязателен", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный ID урока", http.StatusBadRequest)
		return
	}

	var lesson models.Lesson
	var teacherID sql.NullInt64
	var teacherUsername sql.NullString

	err = db.DB.QueryRow(`
		SELECT 
			l.id, 
			l.name, 
			l.description, 
			l.course_id,
			l.teacher_id,
			u.username as teacher_username
		FROM lessons l
		LEFT JOIN users u ON l.teacher_id = u.id AND u.role = 'teacher'
		WHERE l.id = $1
	`, id).Scan(
		&lesson.ID,
		&lesson.Name,
		&lesson.Description,
		&lesson.CourseID,
		&teacherID,
		&teacherUsername,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Урок не найден", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error querying lesson: %v", err)
		http.Error(w, "Ошибка получения урока", http.StatusInternalServerError)
		return
	}

	if teacherID.Valid && teacherUsername.Valid {
		tID := int(teacherID.Int64)
		lesson.TeacherID = &tID
		lesson.Teacher = &models.Teacher{
			ID:       tID,
			Username: teacherUsername.String,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lesson)
}

func GetLessonWithStudents(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID урока обязателен", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный ID урока", http.StatusBadRequest)
		return
	}

	var lesson models.LessonWithStudents
	var teacherID sql.NullInt64
	var teacherUsername sql.NullString

	err = db.DB.QueryRow(`
		SELECT l.id, l.name, l.description, l.teacher_id,
			   t.username
		FROM lessons l
		LEFT JOIN users t ON l.teacher_id = t.id
		WHERE l.id = $1
	`, id).Scan(
		&lesson.ID, &lesson.Name, &lesson.Description,
		&teacherID, &teacherUsername,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Урок не найден", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Ошибка получения урока", http.StatusInternalServerError)
		return
	}

	if teacherID.Valid && teacherUsername.Valid {
		tID := int(teacherID.Int64)
		lesson.Teacher = &models.Teacher{
			ID:       tID,
			Username: teacherUsername.String,
		}
		lesson.TeacherID = &tID
	}

	rows, err := db.DB.Query(`
		SELECT 
			u.id, 
			u.username, 
			la.grade,
			COALESCE(la.attendance, false) as attendance
		FROM users u
		JOIN lesson_attendance la ON u.id = la.student_id
		WHERE la.lesson_id = $1 AND u.role = 'student'
	`, id)
	if err != nil {
		http.Error(w, "Ошибка получения списка студентов", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var student models.StudentWithAttendance
		var grade sql.NullInt64
		if err := rows.Scan(&student.ID, &student.Username, &grade, &student.Attendance); err != nil {
			http.Error(w, "Ошибка обработки данных студента", http.StatusInternalServerError)
			return
		}
		if grade.Valid {
			gradeInt := int(grade.Int64)
			student.Grade = &gradeInt
		}
		lesson.Students = append(lesson.Students, student)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lesson)
}

func CreateLesson(w http.ResponseWriter, r *http.Request) {
	var lesson models.Lesson
	if err := json.NewDecoder(r.Body).Decode(&lesson); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("Creating lesson: %+v", lesson)

	var courseExists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM courses WHERE id = $1)", lesson.CourseID).Scan(&courseExists)
	if err != nil {
		log.Printf("Error checking course existence: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !courseExists {
		log.Printf("Course with ID %d not found", lesson.CourseID)
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	if lesson.TeacherID != nil {
		var teacherExists bool
		err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND role = 'teacher')", *lesson.TeacherID).Scan(&teacherExists)
		if err != nil {
			log.Printf("Error checking teacher existence: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		if !teacherExists {
			log.Printf("Teacher with ID %d not found", *lesson.TeacherID)
			http.Error(w, "Teacher not found", http.StatusBadRequest)
			return
		}
	}

	err = db.DB.QueryRow(`
		INSERT INTO lessons (name, description, course_id, teacher_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, lesson.Name, lesson.Description, lesson.CourseID, lesson.TeacherID).Scan(&lesson.ID)

	if err != nil {
		log.Printf("Error creating lesson: %v", err)
		http.Error(w, "Error creating lesson", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully created lesson with ID: %d", lesson.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(lesson)
}

func GetLessonsByCourse(w http.ResponseWriter, r *http.Request) {
	courseID := r.URL.Query().Get("course_id")
	if courseID == "" {
		http.Error(w, "Course ID required", http.StatusBadRequest)
		return
	}

	rows, err := db.DB.Query(`
		SELECT l.id, l.name, l.description, l.course_id, l.teacher_id,
			   t.username as teacher_name
		FROM lessons l
		LEFT JOIN users t ON l.teacher_id = t.id
		WHERE l.course_id = $1
		ORDER BY l.id
	`, courseID)
	if err != nil {
		http.Error(w, "Error loading lessons", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var lessons []models.LessonWithTeacher
	for rows.Next() {
		var lesson models.LessonWithTeacher
		var teacherID sql.NullInt64
		var teacherName sql.NullString

		err := rows.Scan(
			&lesson.ID,
			&lesson.Name,
			&lesson.Description,
			&lesson.CourseID,
			&teacherID,
			&teacherName,
		)
		if err != nil {
			http.Error(w, "Error processing data", http.StatusInternalServerError)
			return
		}

		if teacherID.Valid && teacherName.Valid {
			tID := int(teacherID.Int64)
			lesson.Teacher = &models.Teacher{
				ID:       tID,
				Username: teacherName.String,
			}
			lesson.TeacherID = &tID
		}

		lessons = append(lessons, lesson)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lessons)
}

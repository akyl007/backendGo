package handlers

import (
	"asii/db"
	"asii/models"
	"asii/utils"
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func CreateCourse(w http.ResponseWriter, r *http.Request) {
	// Получаем данные пользователя из контекста
	claims := utils.UserFromContext(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Проверяем, что пользователь является преподавателем
	if claims.Role != "teacher" {
		http.Error(w, "Only teachers can create courses", http.StatusForbidden)
		return
	}

	var course models.Course
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("Creating course: %+v", course)

	// Устанавливаем ID преподавателя из токена
	course.TeacherID = &claims.UserID

	if course.TeacherID != nil {
		var exists bool
		err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND role = 'teacher')", *course.TeacherID).Scan(&exists)
		if err != nil {
			log.Printf("Error checking teacher existence: %v", err)
			http.Error(w, "Error validating teacher", http.StatusInternalServerError)
			return
		}
		if !exists {
			log.Printf("Teacher with ID %d not found", *course.TeacherID)
			http.Error(w, "Teacher not found", http.StatusBadRequest)
			return
		}
	}

	var tableExists bool
	err := db.DB.QueryRow(`
        SELECT EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_schema = 'public' 
            AND table_name = 'courses'
        )
    `).Scan(&tableExists)

	if err != nil {
		log.Printf("Error checking if courses table exists: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !tableExists {
		log.Printf("Courses table does not exist, creating...")
		_, err = db.DB.Exec(`
            CREATE TABLE courses (
                id SERIAL PRIMARY KEY,
                name VARCHAR(255) NOT NULL,
                description TEXT,
                teacher_id INTEGER REFERENCES users(id)
            )
        `)
		if err != nil {
			log.Printf("Error creating courses table: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}

	err = db.DB.QueryRow(
		"INSERT INTO courses (name, description, teacher_id) VALUES ($1, $2, $3) RETURNING id",
		course.Name, course.Description, course.TeacherID,
	).Scan(&course.ID)

	if err != nil {
		log.Printf("Error creating course: %v", err)
		http.Error(w, "Error creating course", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully created course with ID: %d", course.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(course)
}

func GetCourses(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
        SELECT c.id, c.name, c.description, c.teacher_id,
               t.id, t.username
        FROM courses c
        LEFT JOIN users t ON c.teacher_id = t.id AND t.role = 'teacher'
    `)
	if err != nil {
		http.Error(w, "Error loading courses", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var courses []models.CourseWithTeacher
	for rows.Next() {
		var course models.CourseWithTeacher
		var teacherID sql.NullInt64
		var teacherUsername sql.NullString

		err := rows.Scan(
			&course.ID, &course.Name, &course.Description, &teacherID,
			&teacherID, &teacherUsername,
		)
		if err != nil {
			http.Error(w, "Error processing data", http.StatusInternalServerError)
			return
		}

		if teacherID.Valid && teacherUsername.Valid {
			tID := int(teacherID.Int64)
			course.Teacher = &models.Teacher{
				ID:       tID,
				Username: teacherUsername.String,
			}
			course.TeacherID = &tID
		}

		courses = append(courses, course)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(courses)
}

func UpdateCourse(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Course ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	var course models.Course
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var exists bool
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM courses WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		log.Printf("Error checking course existence: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	if course.TeacherID != nil {
		err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND role = 'teacher')", *course.TeacherID).Scan(&exists)
		if err != nil {
			log.Printf("Error checking teacher existence: %v", err)
			http.Error(w, "Error validating teacher", http.StatusInternalServerError)
			return
		}
		if !exists {
			log.Printf("Teacher with ID %d not found", *course.TeacherID)
			http.Error(w, "Teacher not found", http.StatusBadRequest)
			return
		}
	}

	result, err := db.DB.Exec(
		"UPDATE courses SET name = $1, description = $2, teacher_id = $3 WHERE id = $4",
		course.Name, course.Description, course.TeacherID, id,
	)
	if err != nil {
		log.Printf("Error updating course: %v", err)
		http.Error(w, "Error updating course", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting affected rows: %v", err)
		http.Error(w, "Error getting result", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	var updatedCourse models.CourseWithTeacher
	err = db.DB.QueryRow(`
        SELECT c.id, c.name, c.description, c.teacher_id,
               t.id, t.username
        FROM courses c
        LEFT JOIN users t ON c.teacher_id = t.id AND t.role = 'teacher'
        WHERE c.id = $1
    `, id).Scan(
		&updatedCourse.ID,
		&updatedCourse.Name,
		&updatedCourse.Description,
		&updatedCourse.TeacherID,
		&updatedCourse.Teacher.ID,
		&updatedCourse.Teacher.Username,
	)
	if err != nil {
		log.Printf("Error fetching updated course: %v", err)
		http.Error(w, "Error fetching updated course", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedCourse)
}

func DeleteCourse(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Course ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("DELETE FROM lesson_students WHERE lesson_id IN (SELECT id FROM lessons WHERE course_id = $1)", id)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Error deleting lesson students", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("DELETE FROM lessons WHERE course_id = $1", id)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Error deleting lessons", http.StatusInternalServerError)
		return
	}

	result, err := tx.Exec("DELETE FROM courses WHERE id = $1", id)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Error deleting course", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error getting result", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Course deleted successfully"})
}

func GetCourse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	courseID := vars["id"]

	if courseID == "" {
		http.Error(w, "Course ID is required", http.StatusBadRequest)
		return
	}

	query := `
        SELECT 
            c.id, 
            c.name, 
            c.description,
            c.teacher_id,
            u.username as teacher_name,
            COUNT(DISTINCT cs.student_id) as students_count,
            COUNT(DISTINCT l.id) as lessons_count
        FROM courses c
        LEFT JOIN users u ON c.teacher_id = u.id
        LEFT JOIN course_students cs ON c.id = cs.course_id
        LEFT JOIN lessons l ON c.id = l.course_id
        WHERE c.id = $1
        GROUP BY c.id, c.name, c.description, c.teacher_id, u.username
    `

	var course struct {
		models.Course
		TeacherName   *string `json:"teacher_name,omitempty"`
		StudentsCount int     `json:"students_count"`
		LessonsCount  int     `json:"lessons_count"`
	}

	var teacherName sql.NullString
	err := db.DB.QueryRow(query, courseID).Scan(
		&course.ID,
		&course.Name,
		&course.Description,
		&course.TeacherID,
		&teacherName,
		&course.StudentsCount,
		&course.LessonsCount,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Error fetching course: %v", err)
		http.Error(w, "Error fetching course", http.StatusInternalServerError)
		return
	}

	if teacherName.Valid {
		course.TeacherName = &teacherName.String
	}

	studentsQuery := `
        SELECT u.id, u.username
        FROM users u
        JOIN course_students cs ON u.id = cs.student_id
        WHERE cs.course_id = $1
    `

	rows, err := db.DB.Query(studentsQuery, courseID)
	if err != nil {
		log.Printf("Error fetching students: %v", err)
		http.Error(w, "Error fetching course students", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var student models.Student
		if err := rows.Scan(&student.ID, &student.Username); err != nil {
			log.Printf("Error scanning student: %v", err)
			continue
		}
		students = append(students, student)
	}

	response := struct {
		models.Course
		TeacherName   *string          `json:"teacher_name,omitempty"`
		StudentsCount int              `json:"students_count"`
		LessonsCount  int              `json:"lessons_count"`
		Students      []models.Student `json:"students,omitempty"`
	}{
		Course:        course.Course,
		TeacherName:   course.TeacherName,
		StudentsCount: course.StudentsCount,
		LessonsCount:  course.LessonsCount,
		Students:      students,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

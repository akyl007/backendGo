package handlers

import (
	"asii/db"
	"asii/models"
	"asii/utils"
	"database/sql"
	"encoding/json"
	"net/http"
)

func GetStudentProgress(w http.ResponseWriter, r *http.Request) {
	claims := utils.UserFromContext(r.Context())
	studentID := claims.UserID

	rows, err := db.DB.Query(`
        SELECT 
            u.id as student_id,
            u.username,
            c.name as course_name,
            l.name as lesson_name,
            ls.grade,
            CASE WHEN ls.grade IS NOT NULL THEN true ELSE false END as completed,
            COALESCE(AVG(ls.grade) OVER (PARTITION BY c.id), 0) as average_grade
        FROM users u
        JOIN lesson_students ls ON u.id = ls.student_id
        JOIN lessons l ON ls.lesson_id = l.id
        JOIN courses c ON l.course_id = c.id
        WHERE u.id = $1
        ORDER BY c.name, l.name
    `, studentID)

	if err != nil {
		http.Error(w, "Error fetching progress", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var progress []models.StudentProgress
	for rows.Next() {
		var p models.StudentProgress
		var grade sql.NullInt64
		err := rows.Scan(
			&p.StudentID,
			&p.Username,
			&p.CourseName,
			&p.LessonName,
			&grade,
			&p.Completed,
			&p.AverageGrade,
		)
		if err != nil {
			http.Error(w, "Error processing progress data", http.StatusInternalServerError)
			return
		}
		if grade.Valid {
			gradeInt := int(grade.Int64)
			p.Grade = &gradeInt
		}
		progress = append(progress, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progress)
}

func GetCourseProgress(w http.ResponseWriter, r *http.Request) {
	courseID := r.URL.Query().Get("course_id")
	if courseID == "" {
		http.Error(w, "Course ID required", http.StatusBadRequest)
		return
	}

	var progress models.CourseProgress
	err := db.DB.QueryRow(`
        WITH CourseStats AS (
            SELECT 
                c.id as course_id,
                c.name as course_name,
                COUNT(DISTINCT l.id) as total_lessons,
                COUNT(DISTINCT CASE WHEN ls.grade IS NOT NULL THEN l.id END) as completed_lessons,
                COALESCE(AVG(ls.grade), 0) as average_grade
            FROM courses c
            LEFT JOIN lessons l ON c.id = l.course_id
            LEFT JOIN lesson_students ls ON l.id = ls.lesson_id
            WHERE c.id = $1
            GROUP BY c.id, c.name
        )
        SELECT 
            course_id,
            course_name,
            total_lessons,
            completed_lessons,
            average_grade
        FROM CourseStats
    `, courseID).Scan(
		&progress.CourseID,
		&progress.CourseName,
		&progress.TotalLessons,
		&progress.Completed,
		&progress.AverageGrade,
	)

	if err != nil {
		http.Error(w, "Error fetching course progress", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progress)
}

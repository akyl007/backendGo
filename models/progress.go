package models

type StudentProgress struct {
	StudentID    int     `json:"student_id"`
	Username     string  `json:"username"`
	CourseName   string  `json:"course_name"`
	LessonName   string  `json:"lesson_name"`
	Grade        *int    `json:"grade,omitempty"`
	Completed    bool    `json:"completed"`
	AverageGrade float64 `json:"average_grade"`
}

type CourseProgress struct {
	CourseID     int     `json:"course_id"`
	CourseName   string  `json:"course_name"`
	TotalLessons int     `json:"total_lessons"`
	Completed    int     `json:"completed_lessons"`
	AverageGrade float64 `json:"average_grade"`
}

package models

type Lesson struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CourseID    int      `json:"course_id"`
	TeacherID   *int     `json:"teacher_id,omitempty"`
	Teacher     *Teacher `json:"teacher,omitempty"`
}

type LessonWithTeacher struct {
	Lesson
	TeacherName *string `json:"teacher_name,omitempty"`
}

type LessonWithStudents struct {
	Lesson
	Students []StudentWithAttendance `json:"students"`
} 
package models

type Course struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TeacherID   *int   `json:"teacher_id,omitempty"`
}

type CourseWithTeacher struct {
	Course
	Teacher *Teacher `json:"teacher,omitempty"`
}

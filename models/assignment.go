package models

type CourseAssignment struct {
	CourseID   int  `json:"course_id"`
	StudentID  int  `json:"student_id"`
	TeacherID  int  `json:"teacher_id"`
}

type LessonAttendance struct {
	LessonID   int  `json:"lesson_id"`
	StudentID  int  `json:"student_id"`
	Attendance bool `json:"attendance"`
	Grade      *int `json:"grade,omitempty"`
}

type StudentWithAttendance struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	Attendance bool   `json:"attendance"`
	Grade      *int   `json:"grade,omitempty"`
}

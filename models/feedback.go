package models

import "time"

type Feedback struct {
    ID        int       `json:"id"`
    CourseID  int       `json:"course_id"`
    StudentID int       `json:"student_id"`
    TeacherID int       `json:"teacher_id"`
    Comment   string    `json:"comment"`
    CreatedAt time.Time `json:"created_at"`
}

type FeedbackRequest struct {
    CourseID  int    `json:"course_id"`
    TeacherID int    `json:"teacher_id"`
    Comment   string `json:"comment"`
} 
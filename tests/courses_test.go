package tests

import (
	"asii/db"
	"asii/handlers"
	"asii/models"
	"asii/utils"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateCourse(t *testing.T) {

	setupTestDB()
	defer cleanupTestDB()

	t.Run("Create Course as Teacher", func(t *testing.T) {

		_, err := db.DB.Exec(`
			INSERT INTO users (username, password, role) 
			VALUES ($1, $2, $3)`,
			"testteacher", "$2a$10$somehashedpassword", "teacher")
		if err != nil {
			t.Fatalf("Failed to create test teacher: %v", err)
		}

		course := models.Course{
			Name:        "Test Course",
			Description: "Test Description",
		}
		body, _ := json.Marshal(course)
		req := httptest.NewRequest("POST", "/api/course", bytes.NewBuffer(body))

		token, _ := utils.GenerateToken(1, "testteacher", "teacher")
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), "user", &utils.Claims{
			UserID:   1,
			Username: "testteacher",
			Role:     "teacher",
		})
		req = req.WithContext(ctx)

		handlers.CreateCourse(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
		}

		var response models.Course
		json.NewDecoder(w.Body).Decode(&response)
		if response.Name != course.Name {
			t.Errorf("Expected course name %s, got %s", course.Name, response.Name)
		}
	})

	t.Run("Create Course as Student", func(t *testing.T) {
		// Создаем тестового студента
		_, err := db.DB.Exec(`
			INSERT INTO users (username, password, role) 
			VALUES ($1, $2, $3)`,
			"teststudent", "$2a$10$somehashedpassword", "student")
		if err != nil {
			t.Fatalf("Failed to create test student: %v", err)
		}

		course := models.Course{
			Name:        "Student Course",
			Description: "Test Description",
		}
		body, _ := json.Marshal(course)
		req := httptest.NewRequest("POST", "/api/course", bytes.NewBuffer(body))

		token, _ := utils.GenerateToken(2, "teststudent", "student")
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), "user", &utils.Claims{
			UserID:   2,
			Username: "teststudent",
			Role:     "student",
		})
		req = req.WithContext(ctx)

		handlers.CreateCourse(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
		}
	})
}

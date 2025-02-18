package tests

import (
	"asii/db"
	"asii/handlers"
	_ "asii/models"
	"bytes"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogin(t *testing.T) {

	setupTestDB()
	defer cleanupTestDB()

	t.Run("Successful Login", func(t *testing.T) {
		// Создаем тестового пользователя с правильным хешем пароля
		password := "testpass"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			t.Fatalf("Failed to hash password: %v", err)
		}

		_, err = db.DB.Exec(`
			INSERT INTO users (username, password, role) 
			VALUES ($1, $2, $3)`,
			"teststudent", string(hashedPassword), "student")
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		loginReq := handlers.LoginRequest{
			Username: "teststudent",
			Password: "testpass",
		}
		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.Login(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string]string
		json.NewDecoder(w.Body).Decode(&response)
		if _, exists := response["token"]; !exists {
			t.Error("Expected token in response")
		}
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		loginReq := handlers.LoginRequest{
			Username: "wronguser",
			Password: "wrongpass",
		}
		body, _ := json.Marshal(loginReq)
		req := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.Login(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})
}

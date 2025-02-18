package tests

import (
	"asii/db"
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

func setupTestDB() {
	// Подключение к тестовой базе данных с правильными учетными данными
	testDB, err := sql.Open("postgres", "postgres://myuser:mypassword@localhost:5432/test_mini_moodle?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	// Проверяем подключение
	err = testDB.Ping()
	if err != nil {
		log.Fatalf("Error connecting to test database: %v", err)
	}

	// Создаем необходимые таблицы
	_, err = testDB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL
		);

		CREATE TABLE IF NOT EXISTS courses (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			teacher_id INTEGER REFERENCES users(id)
		);

		CREATE TABLE IF NOT EXISTS forum_messages (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id),
			message TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Fatalf("Error creating tables: %v", err)
	}

	// Очистка таблиц перед тестами
	clearTables(testDB)

	// Присваиваем тестовую БД глобальной переменной
	db.DB = testDB
}

func clearTables(db *sql.DB) {
	_, err := db.Exec(`
        TRUNCATE forum_messages, courses, users RESTART IDENTITY CASCADE;
    `)
	if err != nil {
		log.Fatal(err)
	}
}

func insertTestData(db *sql.DB) {
	// Создаем тестовых пользователей
	_, err := db.Exec(`
        INSERT INTO users (username, password, role) VALUES 
        ('testteacher', '$2a$10$somehashedpassword', 'teacher'),
        ('testadmin', '$2a$10$somehashedpassword', 'admin');
    `)
	if err != nil {
		log.Fatal(err)
	}
}

func cleanupTestDB() {
	if db.DB != nil {
		db.DB.Close()
	}
}

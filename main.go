package main

import (
	"asii/config"
	"asii/db"
	"asii/routes"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Загрузка конфигурации
	if err := config.LoadConfig("config/config.json"); err != nil {
		log.Fatal("Cannot load config:", err)
	}

	// База данных
	if err := db.Connect(); err != nil {
		log.Fatal("Cannot connect to database:", err)
	}
	defer db.DB.Close()

	// Маршрутизация
	router := routes.SetupRouter()

	// CORS
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		router.ServeHTTP(w, r)
	})

	// Сервер
	serverAddr := fmt.Sprintf(":%d", config.AppConfig.Server.Port)
	log.Printf("Server starting on http://localhost%s", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, handler))
}

// @title People API
// @version 1.0
// @description Сервис для обогащения данных о людях
// @host localhost:8080
// @BasePath /
// @schemes http

package main

import (
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"net/http"
	_ "whois-api/docs"
	"whois-api/internal/config"
	"whois-api/internal/handlers"
	"whois-api/internal/middleware"
	"whois-api/internal/repositories"
	"whois-api/internal/service"
)

func main() {
	logger := config.NewLogger(config.Config{})
	defer logger.Sync()

	cfg := config.LoadConfig(logger)
	logger = config.NewLogger(cfg)
	defer logger.Sync()

	logger.Info("Configuration loaded", zap.String("log_level", cfg.LogLevel))

	db := config.NewDB(cfg, logger)
	defer db.Close()
	logger.Info("Database connection established")

	repo := repositories.NewPeopleRepository(db, logger)
	svc := service.NewPeopleService(logger, cfg)
	handler := handlers.NewPeopleHandler(repo, svc, logger)

	r := mux.NewRouter()
	r.Use(middleware.LoggingMiddleware(logger))
	r.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)
	r.HandleFunc("/people", handler.CreatePeople).Methods("POST")
	r.HandleFunc("/people", handler.GetPeople).Methods("GET")
	r.HandleFunc("/people/{id}", handler.GetPeopleByID).Methods("GET")
	r.HandleFunc("/people/{id}", handler.UpdatePeople).Methods("PUT")
	r.HandleFunc("/people/{id}", handler.DeletePeople).Methods("DELETE")

	addr := ":" + cfg.Port
	logger.Info("Server is running", zap.String("addr", addr))
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}

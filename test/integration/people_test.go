package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"whois-api/internal/config"
	"whois-api/internal/handlers"
	"whois-api/internal/repositories"
	"whois-api/internal/service"
)

func setupTestHandler(t *testing.T) (*handlers.PeopleHandler, repositories.PeopleRepositoryInterface) {
	t.Helper()

	err := godotenv.Load("../../.env")
	require.NoError(t, err)

	tempLogger := config.NewLogger(config.Config{})
	cfg := config.LoadConfig(tempLogger)
	logger := config.NewLogger(cfg)

	db := config.NewDB(cfg, logger)
	repo := repositories.NewPeopleRepository(db, logger)
	svc := service.NewPeopleService(logger, cfg)
	handler := handlers.NewPeopleHandler(repo, svc, logger)

	return handler, repo
}

func TestCreatePeopleIntegration(t *testing.T) {
	handler, repo := setupTestHandler(t)

	reqBody := map[string]string{
		"name":       "Dmitriy",
		"surname":    "Ushakov",
		"patronymic": "Vasilevich",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/people", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.CreatePeople(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	require.Equal(t, "Dmitriy", result["name"])

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = repo.DeletePeople(ctx, int(result["id"].(float64)))
}

func TestGetPeopleByID_NotFound(t *testing.T) {
	handler, _ := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/people/99999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "99999"})
	w := httptest.NewRecorder()
	handler.GetPeopleByID(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetPeople_Filtered(t *testing.T) {
	handler, repo := setupTestHandler(t)

	reqBody := map[string]string{
		"name":    "Anna",
		"surname": "Testova",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/people", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.CreatePeople(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	var result map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	id := int(result["id"].(float64))

	getReq := httptest.NewRequest(http.MethodGet, "/people?name=Anna", nil)
	getW := httptest.NewRecorder()
	handler.GetPeople(getW, getReq)
	require.Equal(t, http.StatusOK, getW.Code)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = repo.DeletePeople(ctx, id)
}

func TestUpdatePeople_NotFound(t *testing.T) {
	handler, _ := setupTestHandler(t)

	update := map[string]string{
		"name":    "Updated",
		"surname": "User",
	}
	body, _ := json.Marshal(update)
	req := httptest.NewRequest(http.MethodPut, "/people/99999", bytes.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": "99999"})
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.UpdatePeople(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeletePeople_NotFound(t *testing.T) {
	handler, _ := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/people/99999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "99999"})
	w := httptest.NewRecorder()
	handler.DeletePeople(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

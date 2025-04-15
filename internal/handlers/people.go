package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
	"whois-api/internal/models"
	"whois-api/internal/repositories"
	"whois-api/internal/service"

	"github.com/jackc/pgx/v5"
)

type PeopleHandler struct {
	repo   repositories.PeopleRepositoryInterface
	svc    service.PeopleServiceInterface
	logger *zap.Logger
}

func NewPeopleHandler(repo repositories.PeopleRepositoryInterface, svc service.PeopleServiceInterface, logger *zap.Logger) *PeopleHandler {
	return &PeopleHandler{
		repo:   repo,
		svc:    svc,
		logger: logger,
	}
}

// CreatePeople godoc
// @Summary Создание пользователя
// @Description Принимает ФИО, обогащает данные и сохраняет в БД
// @Tags people
// @Accept json
// @Produce json
// @Param input body models.People true "Данные пользователя (ФИО)"
// @Success 200 {object} models.People
// @Failure 400 {object} models.ErrorResponse "Некорректный JSON"
// @Failure 500 {object} models.ErrorResponse "Ошибка при обработке запроса"
// @Router /people [post]
func (h *PeopleHandler) CreatePeople(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("CreatePeople: request received")

	var p models.People
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		h.logger.Warn("CreatePeople: invalid JSON", zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	updated, err := h.svc.EnrichPeople(p)
	if err != nil {
		h.logger.Error("CreatePeople: enrichment failed", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to enrich people data")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id, err := h.repo.CreatePeople(ctx, updated)
	if err != nil {
		h.logger.Error("CreatePeople: failed to save to DB", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to save people to DB")
		return
	}

	updated.ID = id
	h.logger.Info("CreatePeople: people created", zap.Int("id", id))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// GetPeople godoc
// @Summary Получение списка пользователей
// @Description Получает список пользователей с фильтрацией и пагинацией
// @Tags people
// @Accept json
// @Produce json
// @Param name query string false "Фильтр по имени"
// @Param age query int false "Фильтр по возрасту"
// @Param limit query int false "Количество записей (по умолчанию 10)"
// @Param offset query int false "Смещение"
// @Success 200 {array} models.People
// @Failure 500 {object} models.ErrorResponse "Ошибка при получении списка"
// @Router /people [get]
func (h *PeopleHandler) GetPeople(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	ageStr := r.URL.Query().Get("age")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	age, _ := strconv.Atoi(ageStr)
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit == 0 {
		limit = 10
	}

	h.logger.Debug("GetPeople: fetching list", zap.String("name", name), zap.Int("age", age), zap.Int("limit", limit), zap.Int("offset", offset))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	people, err := h.repo.GetPeople(ctx, name, age, limit, offset)
	if err != nil {
		h.logger.Error("GetPeople: failed to fetch", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to get people list")
		return
	}

	h.logger.Info("GetPeople: list fetched", zap.Int("count", len(people)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(people)
}

// GetPeopleByID godoc
// @Summary Получение пользователя по ID
// @Description Возвращает одного пользователя по ID
// @Tags people
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} models.People
// @Failure 400 {object} models.ErrorResponse "Неверный ID"
// @Failure 404 {object} models.ErrorResponse "Пользователь не найден"
// @Router /people/{id} [get]
func (h *PeopleHandler) GetPeopleByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Warn("GetPeopleByID: invalid ID", zap.String("raw_id", idStr))
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	people, err := h.repo.GetPeopleByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "People not found")
		} else {
			h.logger.Error("GetPeopleByID: fetch failed", zap.Int("id", id), zap.Error(err))
			respondWithError(w, http.StatusInternalServerError, "Failed to fetch people")
		}
		return
	}

	h.logger.Info("GetPeopleByID: people found", zap.Int("id", id))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(people)
}

// UpdatePeople godoc
// @Summary Обновление пользователя
// @Description Обновляет данные пользователя по ID
// @Tags people
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Param input body models.People true "Новые данные пользователя"
// @Success 204 "Успешное обновление"
// @Failure 400 {object} models.ErrorResponse "Неверный ID или JSON"
// @Failure 404 {object} models.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} models.ErrorResponse "Ошибка при обновлении"
// @Router /people/{id} [put]
func (h *PeopleHandler) UpdatePeople(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Warn("UpdatePeople: invalid ID", zap.String("raw_id", idStr))
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var p models.People
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		h.logger.Warn("UpdatePeople: invalid JSON", zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	h.logger.Debug("UpdatePeople: request received", zap.Int("id", id))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = h.repo.UpdatePeople(ctx, id, p)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "People not found")
		} else {
			h.logger.Error("UpdatePeople: update failed", zap.Int("id", id), zap.Error(err))
			respondWithError(w, http.StatusInternalServerError, "Failed to update people")
		}
		return
	}

	h.logger.Info("UpdatePeople: people updated", zap.Int("id", id))
	w.WriteHeader(http.StatusNoContent)
}

// DeletePeople godoc
// @Summary Удаление пользователя
// @Description Удаляет пользователя по ID
// @Tags people
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 204 "Пользователь успешно удалён"
// @Failure 400 {object} models.ErrorResponse "Неверный ID"
// @Failure 404 {object} models.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} models.ErrorResponse "Ошибка при удалении"
// @Router /people/{id} [delete]
func (h *PeopleHandler) DeletePeople(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Warn("DeletePeople: invalid ID", zap.String("raw_id", idStr))
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	h.logger.Debug("DeletePeople: request received", zap.Int("id", id))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = h.repo.DeletePeople(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "People not found")
		} else {
			h.logger.Error("DeletePeople: deletion failed", zap.Int("id", id), zap.Error(err))
			respondWithError(w, http.StatusInternalServerError, "Failed to delete people")
		}
		return
	}

	h.logger.Info("DeletePeople: people deleted", zap.Int("id", id))
	w.WriteHeader(http.StatusNoContent)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Code:    code,
		Message: message,
	})
}

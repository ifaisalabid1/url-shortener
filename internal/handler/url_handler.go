package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/ifaisalabid1/url-shortener/internal/model"
	"github.com/ifaisalabid1/url-shortener/internal/repository"
	"github.com/ifaisalabid1/url-shortener/internal/service"
)

type URLHandler struct {
	urlService service.URLService
	validator  *validator.Validate
	logger     *slog.Logger
}

func NewURLHandler(urlService service.URLService, logger *slog.Logger) *URLHandler {
	return &URLHandler{
		urlService: urlService,
		validator:  validator.New(),
		logger:     logger,
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitzero"`
}

func (h *URLHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	var req model.CreateURLRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		h.respondWithError(w, http.StatusOK, err.Error())
		return
	}

	res, err := h.urlService.CreateShortURL(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateCode):
			h.respondWithError(w, http.StatusConflict, "short code already exists")
		default:
			h.logger.Error("failed to create url", "error", err)
			h.respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		return
	}

	h.respondWithJSON(w, http.StatusCreated, res)
}

func (h *URLHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "code")

	original_url, err := h.urlService.GetOriginalURL(r.Context(), shortCode)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrURLNotFound):
			h.respondWithError(w, http.StatusNotFound, "url not found")
		default:
			h.logger.Error("failed to get original url", "error", err)
			h.respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		return
	}

	http.Redirect(w, r, original_url, http.StatusMovedPermanently)
}

func (h *URLHandler) GetURLStats(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "code")

	stats, err := h.urlService.GetURLStats(r.Context(), shortCode)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrURLNotFound):
			h.respondWithError(w, http.StatusNotFound, "url not found")
		default:
			h.logger.Error("Failed to get url stats", "error", err)
			h.respondWithError(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		return
	}

	h.respondWithJSON(w, http.StatusOK, stats)
}

func (h *URLHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondWithJSON(w, http.StatusOK, SuccessResponse{Message: "service is healthy"})
}

func (h *URLHandler) respondWithJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	js, err := json.Marshal(data)
	if err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}

	w.Write(js)
}

func (h *URLHandler) respondWithError(w http.ResponseWriter, status int, message string) {
	h.respondWithJSON(w, status, ErrorResponse{Error: message})
}

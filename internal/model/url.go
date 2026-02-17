package model

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type URL struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ShortCode   string     `json:"short_code" db:"short_code" validate:"required,max=20"`
	OriginalURL string     `json:"original_url" db:"original_url" validate:"required,url"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	Clicks      int64      `json:"clicks" db:"clicks"`
	ExpiresAt   *time.Time `json:"expires_at,omitzero" db:"expires_at"`
}

type CreateURLRequest struct {
	OriginalURL string     `json:"original_url" validate:"required,url"`
	CustomCode  *string    `json:"custom_code,omitzero" validate:"omitzero,max=20,alphanum"`
	ExpiresAt   *time.Time `json:"expires_at,omitzero"`
}

type URLResponse struct {
	ID          string     `json:"id"`
	ShortCode   string     `json:"short_code"`
	ShortURL    string     `json:"short_url"`
	OriginalURL string     `json:"original_url"`
	CreatedAt   time.Time  `json:"created_at"`
	Clicks      int64      `json:"clicks"`
	ExpiresAt   *time.Time `json:"expires_at,omitzero"`
}

type URLStats struct {
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	Clicks      int64     `json:"clicks"`
	CreatedAt   time.Time `json:"created_at"`
}

func (u *URL) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}

func (u *CreateURLRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(u)
}

func (u *URL) ToResponse(baseURL string) *URLResponse {
	return &URLResponse{
		ID:          u.ID.String(),
		ShortCode:   u.ShortCode,
		ShortURL:    baseURL + "/" + u.ShortCode,
		OriginalURL: u.OriginalURL,
		CreatedAt:   u.CreatedAt,
		Clicks:      u.Clicks,
		ExpiresAt:   u.ExpiresAt,
	}
}

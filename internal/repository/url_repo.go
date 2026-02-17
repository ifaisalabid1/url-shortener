package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ifaisalabid1/url-shortener/internal/model"
	"github.com/lib/pq"
)

var (
	ErrURLNotFound   = errors.New("url not found")
	ErrDuplicateCode = errors.New("short code already exists")
)

type URLRepository interface {
	Create(ctx context.Context, url *model.URL) error
	GetByShortCode(ctx context.Context, shortCode string) (*model.URL, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.URL, error)
	IncrementClicks(ctx context.Context, shortCode string) error
	DeleteExpired(ctx context.Context) error
}

type urlRepository struct {
	db *sql.DB
}

func NewURLRepository(db *sql.DB) URLRepository {
	return &urlRepository{db: db}
}

func (r *urlRepository) Create(ctx context.Context, url *model.URL) error {
	query := "INSERT INTO urls (id, short_code, original_url, created_at, updated_at, clicks, expires_at) VALUES ($1, $2, $3, $4, $5, $6, $7)"

	args := []any{url.ID, url.ShortCode, url.OriginalURL, url.CreatedAt, url.UpdatedAt, url.Clicks, url.ExpiresAt}

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return ErrDuplicateCode
			}
		}

		return fmt.Errorf("failed to create url: %w", err)
	}

	return nil

}

func (r *urlRepository) GetByShortCode(ctx context.Context, shortCode string) (*model.URL, error) {
	var url model.URL

	query := `SELECT id, short_code, original_url, created_at, updated_at, clicks, expires_at
			  FROM urls
			  WHERE short_code = $1 AND (expires_at IS NULL OR expires_at > NOW())`

	err := r.db.QueryRowContext(ctx, query, url.ShortCode).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.CreatedAt,
		&url.UpdatedAt,
		&url.Clicks,
		&url.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrURLNotFound
		}

		return nil, fmt.Errorf("failed to get url: %w", err)
	}

	return &url, nil
}

func (r *urlRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.URL, error) {
	var url model.URL

	query := `SELECT id, short_code, original_url, created_at, updated_at, clicks, expires_at
			  FROM urls
			  WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, url.ID).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.CreatedAt,
		&url.UpdatedAt,
		&url.Clicks,
		&url.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrURLNotFound
		}

		return nil, fmt.Errorf("failed to get url: %w", err)
	}

	return &url, nil
}

func (r *urlRepository) IncrementClicks(ctx context.Context, shortCode string) error {
	query := "UPDATE urls SET clicks = clicks + 1 WHERE short_code = $1"

	result, err := r.db.ExecContext(ctx, query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to increment clicks: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrURLNotFound
	}

	return nil
}

func (r *urlRepository) DeleteExpired(ctx context.Context) error {
	query := "DELETE FROM urls WHERE expires_at <= NOW()"

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete url: %w", err)
	}

	return nil
}

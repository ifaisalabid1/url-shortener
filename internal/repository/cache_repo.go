package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ifaisalabid1/url-shortener/internal/model"
	"github.com/redis/go-redis/v9"
)

type CacheRepository interface {
	SetURL(ctx context.Context, shortCode string, url *model.URL, ttl time.Duration) error
	GetURL(ctx context.Context, shortCode string) (*model.URL, error)
	DeleteURL(ctx context.Context, shortCode string) error
	IncrementClicks(ctx context.Context, shortCode string) error
}

type cacheRepository struct {
	client *redis.Client
}

func NewClientRepository(client *redis.Client) CacheRepository {
	return &cacheRepository{client: client}
}

func (r *cacheRepository) SetURL(ctx context.Context, shortCode string, url *model.URL, ttl time.Duration) error {
	data, err := json.Marshal(url)
	if err != nil {
		return fmt.Errorf("failed to marshal URL: %w", err)
	}

	key := fmt.Sprintf("url:%s", shortCode)
	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set URL in cache: %w", err)
	}

	return nil
}

func (r *cacheRepository) GetURL(ctx context.Context, shortCode string) (*model.URL, error) {
	key := fmt.Sprintf("url:%s", shortCode)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get URL from cache: %w", err)
	}

	var url model.URL
	if err := json.Unmarshal(data, &url); err != nil {
		return nil, fmt.Errorf("failed to unmarshal URL: %w", err)
	}

	return &url, nil
}

func (r *cacheRepository) DeleteURL(ctx context.Context, shortCode string) error {
	key := fmt.Sprintf("url:%s", shortCode)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete URL from cache: %w", err)
	}
	return nil
}

func (r *cacheRepository) IncrementClicks(ctx context.Context, shortCode string) error {
	return nil
}

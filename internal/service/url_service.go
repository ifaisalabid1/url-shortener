package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/ifaisalabid1/url-shortener/internal/model"
	"github.com/ifaisalabid1/url-shortener/internal/repository"
	"github.com/itchyny/base58-go"
)

type URLService interface {
	CreateShortURL(ctx context.Context, req *model.CreateURLRequest) (*model.URLResponse, error)
	GetOriginalURL(ctx context.Context, shortCode string) (string, error)
	GetURLStats(ctx context.Context, shortCode string) (*model.URLStats, error)
}

type urlService struct {
	urlRepo   repository.URLRepository
	cacheRepo repository.CacheRepository
	baseURL   string
	shortLen  int
	cacheTTL  time.Duration
}

func NewURLService(urlRepo repository.URLRepository, cacheRepo repository.CacheRepository, baseURL string, shortLen int, cacheTTL time.Duration) URLService {
	return &urlService{
		urlRepo,
		cacheRepo,
		baseURL,
		shortLen,
		cacheTTL,
	}
}

func (s *urlService) CreateShortURL(ctx context.Context, req *model.CreateURLRequest) (*model.URLResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var shortCode string
	if req.CustomCode != nil && *req.CustomCode != "" {
		existing, err := s.urlRepo.GetByShortCode(ctx, *req.CustomCode)
		if err == nil && existing != nil {
			return nil, repository.ErrDuplicateCode
		}

		shortCode = *req.CustomCode
	} else {
		shortCode = s.generateShortCode(req.OriginalURL)
	}

	now := time.Now().UTC()

	url := &model.URL{
		ID:          uuid.New(),
		ShortCode:   shortCode,
		OriginalURL: req.OriginalURL,
		CreatedAt:   now,
		UpdatedAt:   now,
		Clicks:      0,
		ExpiresAt:   req.ExpiresAt,
	}

	if err := s.urlRepo.Create(ctx, url); err != nil {
		return nil, fmt.Errorf("failed to create url: %w", err)
	}

	if err := s.cacheRepo.SetURL(ctx, shortCode, url, s.cacheTTL); err != nil {
		fmt.Printf("failed to cache url: %v\n", err)
	}

	return url.ToResponse(s.baseURL), nil

}

func (s *urlService) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	cachedURL, err := s.cacheRepo.GetURL(ctx, shortCode)
	if err != nil {
		fmt.Printf("failed to get from cache: %v\n", err)
	}

	var url *model.URL

	if cachedURL != nil {
		url = cachedURL
	} else {
		url, err := s.urlRepo.GetByShortCode(ctx, shortCode)
		if err != nil {
			return "", err
		}

		if err := s.cacheRepo.SetURL(ctx, shortCode, url, s.cacheTTL); err != nil {
			fmt.Printf("Failed to cache URL: %v\n", err)
		}
	}

	if url.ExpiresAt != nil && url.ExpiresAt.Before(time.Now().UTC()) {
		return "", repository.ErrURLNotFound
	}

	go func() {
		if err := s.urlRepo.IncrementClicks(ctx, shortCode); err != nil {
			fmt.Printf("failed to increment clicks: %v\n", err)
		}
	}()

	return url.OriginalURL, nil
}

func (s *urlService) GetURLStats(ctx context.Context, shortCode string) (*model.URLStats, error) {
	url, err := s.urlRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	return &model.URLStats{
		ShortCode:   url.ShortCode,
		OriginalURL: url.OriginalURL,
		Clicks:      url.Clicks,
		CreatedAt:   url.CreatedAt,
	}, nil
}

func (s *urlService) generateShortCode(originalURL string) string {

	data := fmt.Sprintf("%s:%d", originalURL, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))

	num := new(big.Int).SetBytes(hash[:])

	encoded, _ := base58.BitcoinEncoding.Encode([]byte(num.String()))

	if len(encoded) > s.shortLen {
		return string(encoded[:s.shortLen])
	}

	return string(encoded)
}

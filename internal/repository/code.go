package repository

import (
	"context"
	"github.com/lutcoding/redbook/internal/repository/cache"
)

var (
	ErrCodeSendTooFrequent    = cache.ErrCodeSendTooFrequent
	ErrCodeVerifyTooManyTimes = cache.ErrCodeVerifyTooManyTimes
)

type CodeRepository interface {
	Store(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type CodeCacheRepository struct {
	cache cache.CodeCache
}

func NewCodeCacheRepository(cache cache.CodeCache) *CodeCacheRepository {
	return &CodeCacheRepository{
		cache: cache,
	}
}

func (r *CodeCacheRepository) Store(ctx context.Context, biz, phone, code string) error {
	return r.cache.Set(ctx, biz, phone, code)
}

func (r *CodeCacheRepository) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return r.cache.Verify(ctx, biz, phone, inputCode)
}

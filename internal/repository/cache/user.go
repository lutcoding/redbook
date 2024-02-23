package cache

import (
	"context"
	"encoding/json"
	"github.com/lutcoding/redbook/common/globalkey"
	"github.com/lutcoding/redbook/internal/domain"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

var (
	ErrKeyNotExist = redis.Nil
)

type UserCache interface {
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, u domain.User) error
}

type UserRedisCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func NewUserRedisCache(client redis.Cmdable) *UserRedisCache {
	return &UserRedisCache{
		client:     client,
		expiration: time.Hour,
	}
}

func (cache *UserRedisCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.key(id)
	val, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal(val, &u)
	return u, err
}

func (cache *UserRedisCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return cache.client.Set(ctx, cache.key(u.Id), val, cache.expiration).Err()
}

func (cache *UserRedisCache) key(id int64) string {
	return globalkey.UserIdCachePrefix + strconv.FormatInt(id, 10)
}

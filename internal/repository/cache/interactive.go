package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/lutcoding/redbook/internal/domain"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"strconv"
)

var (
	//go:embed lua/incr_cnt.lua
	luaIncrCnt string
)

var (
	ErrNotExistKey = errors.New("缓存key不存在")
)

const fieldReadCnt = "read_cnt"
const fieldLikeCnt = "like_cnt"
const fieldCollectCnt = "collect_cnt"

type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	SetInteractiveInfo(ctx context.Context, biz string, bizId int64, info domain.Interactive) error
	GetInteractiveInfo(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
}

type InteractiveRedisCache struct {
	client redis.Cmdable
}

func (cache *InteractiveRedisCache) SetInteractiveInfo(ctx context.Context, biz string, bizId int64, info domain.Interactive) error {
	return cache.client.HSet(ctx, cache.key(biz, bizId),
		fieldReadCnt, info.ReadCnt, fieldLikeCnt, info.LikeCnt, fieldCollectCnt, info.CollectCnt).Err()
}

func (cache *InteractiveRedisCache) GetInteractiveInfo(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	result, err := cache.client.HGetAll(ctx, cache.key(biz, bizId)).Result()
	if err != nil {
		return domain.Interactive{}, err
	}
	if len(result) == 0 {
		return domain.Interactive{}, ErrNotExistKey
	}
	var res domain.Interactive
	res.ReadCnt, err = strconv.ParseInt(result[fieldReadCnt], 10, 64)
	res.LikeCnt, err = strconv.ParseInt(result[fieldLikeCnt], 10, 64)
	res.CollectCnt, err = strconv.ParseInt(result[fieldCollectCnt], 10, 64)
	zap.L().Debug("查看缓存", zap.Int64("ReadCnt", res.ReadCnt))
	return res, nil
}

func (cache *InteractiveRedisCache) IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizId)}, fieldLikeCnt, 1).Err()
}

func (cache *InteractiveRedisCache) DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizId)}, fieldLikeCnt, -1).Err()
}

func (cache *InteractiveRedisCache) IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizId)}, fieldCollectCnt, 1).Err()
}

func (cache *InteractiveRedisCache) DecrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizId)}, fieldCollectCnt, -1).Err()
}

func (cache *InteractiveRedisCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizId)}, fieldReadCnt, 1).Err()
}

func (cache *InteractiveRedisCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}

func NewInteractiveRedisCache(client redis.Cmdable) *InteractiveRedisCache {
	return &InteractiveRedisCache{
		client: client,
	}
}

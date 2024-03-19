package repository

import (
	"context"
	"errors"
	"github.com/lutcoding/redbook/internal/domain"
	"github.com/lutcoding/redbook/internal/repository/cache"
	"github.com/lutcoding/redbook/internal/repository/dao"
	"go.uber.org/zap"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	IncrLikeCnt(ctx context.Context, uid int64, biz string, bizId int64) error
	IncrCollectCnt(ctx context.Context, biz string, bizId int64) error
	DecrLikeCnt(ctx context.Context, uid int64, biz string, bizId int64) error
	DecrCollectCnt(ctx context.Context, biz string, bizId int64) error
	GetInteractiveInfo(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Liked(ctx context.Context, uid int64, biz string, bizId int64) (bool, error)
	Collected(ctx context.Context, uid int64, biz string, bizId int64) (bool, error)
}

type InteractiveCacheRepository struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
}

func (repo *InteractiveCacheRepository) GetInteractiveInfo(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	res, err := repo.cache.GetInteractiveInfo(ctx, biz, bizId)
	if err == nil {
		zap.L().Info("命中缓存", zap.Int64("biz_id", bizId))
		return res, nil
	}
	info, err := repo.dao.GetInteractiveInfo(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}
	res = repo.entityToDomain(info)
	go func() {
		err := repo.cache.SetInteractiveInfo(ctx, biz, bizId, res)
		if err != nil {
			zap.L().Error("设置缓存失败", zap.Int64("biz_id", bizId))
		}
	}()
	return res, nil
}

func (repo *InteractiveCacheRepository) Liked(ctx context.Context, uid int64, biz string, bizId int64) (bool, error) {
	_, err := repo.dao.GetLikeInfo(ctx, uid, biz, bizId)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (repo *InteractiveCacheRepository) Collected(ctx context.Context, uid int64, biz string, bizId int64) (bool, error) {
	return false, nil
}

func (repo *InteractiveCacheRepository) IncrLikeCnt(ctx context.Context, uid int64, biz string, bizId int64) error {
	err := repo.dao.InsertLikeInfo(ctx, uid, biz, bizId)
	if err != nil {
		if errors.Is(err, dao.ErrDuplicateLike) {
			zap.L().Info("用户重复点赞",
				zap.Int64("uid", uid),
				zap.Int64("biz_id", bizId))
			return nil
		}
		return err
	}
	return repo.cache.IncrLikeCntIfPresent(ctx, biz, bizId)
}

func (repo *InteractiveCacheRepository) DecrLikeCnt(ctx context.Context, uid int64, biz string, bizId int64) error {
	err := repo.dao.DelLikeInfo(ctx, uid, biz, bizId)
	if err != nil {
		if errors.Is(err, dao.ErrDislikeNoRow) {
			zap.L().Info("用户取消一篇文章点赞", zap.Int64("uid", uid))
			return nil
		}
		return err
	}
	return repo.cache.DecrLikeCntIfPresent(ctx, biz, bizId)
}

func (repo *InteractiveCacheRepository) IncrCollectCnt(ctx context.Context, biz string, bizId int64) error {
	//TODO implement me
	panic("implement me")
}

func (repo *InteractiveCacheRepository) DecrCollectCnt(ctx context.Context, biz string, bizId int64) error {
	//TODO implement me
	panic("implement me")
}

func (repo *InteractiveCacheRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	err := repo.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	return repo.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (repo *InteractiveCacheRepository) entityToDomain(info dao.Interactive) domain.Interactive {
	return domain.Interactive{
		ReadCnt:    info.ReadCnt,
		LikeCnt:    info.LikeCnt,
		CollectCnt: info.CollectCnt,
	}
}

func NewInteractiveCacheRepository(dao dao.InteractiveDAO, cache cache.InteractiveCache) *InteractiveCacheRepository {
	return &InteractiveCacheRepository{
		dao:   dao,
		cache: cache,
	}
}

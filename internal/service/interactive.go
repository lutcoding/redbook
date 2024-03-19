package service

import (
	"context"
	"github.com/lutcoding/redbook/internal/domain"
	"github.com/lutcoding/redbook/internal/repository"
	"golang.org/x/sync/errgroup"
)

type InteractiveService struct {
	repo repository.InteractiveRepository
}

func NewInteractiveService(repo repository.InteractiveRepository) *InteractiveService {
	return &InteractiveService{
		repo: repo,
	}
}

func (s *InteractiveService) GetInteractiveInfo(ctx context.Context, uid int64, biz string, bizId int64) (domain.Interactive, error) {
	res, err := s.repo.GetInteractiveInfo(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}
	var eg errgroup.Group
	eg.Go(func() error {
		res.Liked, err = s.repo.Liked(ctx, uid, biz, bizId)
		return err
	})
	eg.Go(func() error {
		res.Collected, err = s.repo.Collected(ctx, uid, biz, bizId)
		return err
	})
	err = eg.Wait()
	return res, err
}

func (s *InteractiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return s.repo.IncrReadCnt(ctx, biz, bizId)
}

func (s *InteractiveService) Like(ctx context.Context, uid int64, biz string, bizId int64) error {
	return s.repo.IncrLikeCnt(ctx, uid, biz, bizId)
}

func (s *InteractiveService) CancelLike(ctx context.Context, uid int64, biz string, bizId int64) error {
	return s.repo.DecrLikeCnt(ctx, uid, biz, bizId)
}

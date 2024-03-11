package repository

import (
	"context"
	"encoding/json"
	"github.com/lutcoding/redbook/internal/domain"
	"github.com/lutcoding/redbook/internal/repository/dao"
)

var ErrWaitingSMSNotFound = dao.ErrWaitingSMSNotFound

type AsyncSmsRepository interface {
	Add(ctx context.Context, sms domain.AsyncSms) error
	GetWaitingSms(ctx context.Context) (domain.AsyncSms, error)
	ReportScheduleResult(ctx context.Context, id int64, success bool) error
}

type AsyncSmsCacheRepository struct {
	dao dao.AsyncSmsDAO
}

func NewAsyncSmsCacheRepository(dao dao.AsyncSmsDAO) *AsyncSmsCacheRepository {
	return &AsyncSmsCacheRepository{
		dao: dao,
	}
}

func (repo *AsyncSmsCacheRepository) Add(ctx context.Context, sms domain.AsyncSms) error {
	return repo.dao.Insert(ctx, repo.domainToEntity(sms))
}

func (repo *AsyncSmsCacheRepository) GetWaitingSms(ctx context.Context) (domain.AsyncSms, error) {
	sms, err := repo.dao.GetWaitingSms(ctx)
	if err != nil {
		return domain.AsyncSms{}, err
	}
	return repo.entityToDomain(sms), nil
}

func (repo *AsyncSmsCacheRepository) ReportScheduleResult(ctx context.Context, id int64, success bool) error {
	if success {
		return repo.dao.MarkSuccess(ctx, id)
	}
	return repo.dao.MarkFailed(ctx, id)
}

func (repo *AsyncSmsCacheRepository) domainToEntity(sms domain.AsyncSms) dao.AsyncSms {
	bytes, _ := json.Marshal(sms.AsyncSmsConfig)
	return dao.AsyncSms{
		Id:             sms.Id,
		RetryCnt:       0,
		RetryMax:       sms.RetryMax,
		AsyncSmsConfig: bytes,
	}
}

func (repo *AsyncSmsCacheRepository) entityToDomain(sms dao.AsyncSms) domain.AsyncSms {
	conf := domain.AsyncSmsConfig{}
	json.Unmarshal(sms.AsyncSmsConfig, &conf)
	return domain.AsyncSms{
		Id:             sms.Id,
		RetryMax:       sms.RetryMax,
		AsyncSmsConfig: conf,
	}
}

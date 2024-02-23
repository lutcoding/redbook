package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/lutcoding/redbook/internal/repository"
	"github.com/lutcoding/redbook/internal/service/sms"
	"math/rand"
)

var (
	ErrCodeSendTooFrequent    = repository.ErrCodeSendTooFrequent
	ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
	ErrCodeNotCorrect         = errors.New("code is not correct, plz input code again")
)

type CodeService struct {
	repo   repository.CodeRepository
	smsSvc sms.Service
	tplId  string
}

func NewCodeService(repo repository.CodeRepository, smsSvc sms.Service, tplId string) *CodeService {
	return &CodeService{
		repo:   repo,
		smsSvc: smsSvc,
		tplId:  tplId,
	}
}

func (svc *CodeService) Send(ctx context.Context, biz string, phone string) error {
	code := svc.generateCode()
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	err = svc.smsSvc.Send(ctx, svc.tplId, []string{code}, phone)
	//if err != nil {
	//	// redis成功
	//	// 可能是超时err，不知道是否成功发出短信
	//	// TODO: retry
	//}
	return err
}

func (svc *CodeService) Verify(ctx context.Context, biz string, phone string, inputCode string) error {
	ok, err := svc.repo.Verify(ctx, biz, phone, inputCode)
	if err != nil {
		return err
	}
	if !ok {
		return ErrCodeNotCorrect
	}
	return nil
}

func (svc *CodeService) generateCode() string {
	num := rand.Intn(1000000)
	return fmt.Sprintf("%06d", num)
}

package auth

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lutcoding/redbook/internal/service/sms"
)

var errNotValid = fmt.Errorf("非合法申请短信服务的用户")

type Service struct {
	sms sms.Service
	key string
}

func NewService(sms sms.Service, key string) *Service {
	return &Service{
		sms: sms,
		key: key,
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tplId, claims, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid || claims.Tpl == "" {
		return errNotValid
	}

	return s.sms.Send(ctx, claims.Tpl, args, numbers...)
}

type Claims struct {
	jwt.RegisteredClaims
	Tpl string
}

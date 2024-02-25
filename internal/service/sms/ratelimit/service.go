package ratelimit

import (
	"context"
	"fmt"
	"github.com/lutcoding/redbook/internal/service/sms"
	"github.com/lutcoding/redbook/pkg/ratelimit"
)

var errLimited = fmt.Errorf("trigger sms rate limit")

// Service 装饰器模式
type Service struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

func NewService(svc sms.Service, limiter ratelimit.Limiter) *Service {
	return &Service{
		svc:     svc,
		limiter: limiter,
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	limit, err := s.limiter.Limit(ctx, "sms_ratelimit")
	if err != nil {
		return fmt.Errorf("短信服务判断是否限流出现问题, %w", err)
	}
	if limit {
		return errLimited
	}
	return s.svc.Send(ctx, tplId, args, numbers...)
}

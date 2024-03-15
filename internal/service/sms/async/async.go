package async

import (
	"context"
	"github.com/lutcoding/redbook/internal/domain"
	"github.com/lutcoding/redbook/internal/repository"
	"github.com/lutcoding/redbook/internal/service/sms"
	"time"
)

type Service struct {
	svc      sms.Service
	repo     repository.AsyncSmsRepository
	retryMax int64
}

func NewService(svc sms.Service, repo repository.AsyncSmsRepository, retryMax int64) *Service {
	return &Service{
		svc:      svc,
		repo:     repo,
		retryMax: retryMax,
	}
}

func (s *Service) Start() {
	go func() {
		for {
			s.async()
		}
	}()
}

func (s *Service) async() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	// 抢占一个异步发送消息
	// 比如k8s部署了三个实例，保证只有一个实例抢占到该消息
	waitingSms, err := s.repo.GetWaitingSms(ctx)
	cancelFunc()
	switch err {
	case nil:
		ctx, cancelFunc = context.WithTimeout(context.Background(), time.Second)
		defer cancelFunc()
		err := s.svc.Send(ctx, waitingSms.TplId, waitingSms.Args, waitingSms.Numbers...)
		if err != nil {
			// 记录下日志
		}
		isSuc := err == nil
		err = s.repo.ReportScheduleResult(ctx, waitingSms.Id, isSuc)
		if err != nil {
			// 记录日志
		}
	case repository.ErrWaitingSMSNotFound:
		time.Sleep(time.Second)
	default:
		// 正常来说应该是数据库那边出了问题，
		// 但是为了尽量运行，还是要继续的
		// 你可以稍微睡眠，也可以不睡眠
		// 睡眠的话可以帮你规避掉短时间的网络抖动问题
		time.Sleep(time.Second)
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	if s.needAsync() {
		return s.repo.Add(ctx, domain.AsyncSms{
			RetryMax: s.retryMax,
			AsyncSmsConfig: domain.AsyncSmsConfig{
				TplId:   tplId,
				Args:    args,
				Numbers: numbers,
			},
		})
	}
	return s.svc.Send(ctx, tplId, args, numbers...)
}

// 提前引导你们，开始思考系统容错问题
// 你们面试装逼，赢得竞争优势就靠这一类的东西
func (s *Service) needAsync() bool {
	// 这边就是你要设计的，各种判定要不要触发异步的方案
	// 1. 基于响应时间的，平均响应时间
	// 1.1 使用绝对阈值，比如说直接发送的时候，（连续一段时间，或者连续N个请求）响应时间超过了 500ms，然后后续请求转异步
	// 1.2 变化趋势，比如说当前一秒钟内的所有请求的响应时间比上一秒钟增长了 X%，就转异步
	// 2. 基于错误率：一段时间内，收到 err 的请求比率大于 X%，转异步

	// 什么时候退出异步
	// 1. 进入异步 N 分钟后
	// 2. 保留 1% 的流量（或者更少），继续同步发送，判定响应时间/错误率
	return true
}

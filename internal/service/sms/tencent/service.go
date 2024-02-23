package tencent

import (
	"context"
	"fmt"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"go.uber.org/multierr"
)

type Service struct {
	appId    *string
	SignName *string
	client   *sms.Client
}

func NewService(client *sms.Client, appId string, signName string) *Service {
	return &Service{
		appId:    common.StringPtr(appId),
		SignName: common.StringPtr(signName),
		client:   client,
	}
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId, req.SignName, req.TemplateId = s.appId, s.SignName, common.StringPtr(tplId)
	req.PhoneNumberSet, req.TemplateParamSet = common.StringPtrs(numbers), common.StringPtrs(args)

	response, err := s.client.SendSms(req)
	if err != nil {
		return err
	}
	for _, status := range response.Response.SendStatusSet {
		if status.Code == nil || *(status.Code) != "OK" {
			err = multierr.Append(err, fmt.Errorf("send message failed %s, %s ", *status.Code, *status.Message))
		}
	}
	return err
}

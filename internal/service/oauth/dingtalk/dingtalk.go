package dingtalk

import (
	"context"
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dingtalkoauth2_1_0 "github.com/alibabacloud-go/dingtalk/oauth2_1_0"
	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/google/uuid"
	"net/url"
)

type Service struct {
	appKey    string
	appSecret string
}

func NewService(appKey string, appSecret string) *Service {
	return &Service{
		appKey:    appKey,
		appSecret: appSecret,
	}
}

func (s *Service) AuthURL(ctx context.Context) (string, error) {
	const urlPattern = "https://login.dingtalk.com/oauth2/auth?redirect_uri=%s&response_type=code&client_id=%s&scope=openid&state=%s&prompt=consent"
	var redirectURI = url.PathEscape("http://127.0.0.1:8080/oauth2/dingtalk/callback")
	return fmt.Sprintf(urlPattern, redirectURI, s.appKey, uuid.NewString()), nil
}

func (s *Service) VerifyCode(ctx context.Context, authCode string, state string) error {
	client, err := s.getClient()
	if err != nil {
		return err
	}
	getUserTokenRequest := &dingtalkoauth2_1_0.GetUserTokenRequest{
		ClientId:     tea.String(s.appKey),
		ClientSecret: tea.String(s.appSecret),
		Code:         tea.String(authCode),
		GrantType:    tea.String("authorization_code"),
	}
	tryErr := func() (err error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				err = r
			}
		}()
		var resp *dingtalkoauth2_1_0.GetUserTokenResponse
		resp, err = client.GetUserToken(getUserTokenRequest)
		if err != nil {
			return err
		}
		fmt.Println(resp.Body.String())
		return nil
	}()
	if tryErr != nil {
		var err = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			err = _t
		} else {
			err.Message = tea.String(tryErr.Error())
		}
		if !tea.BoolValue(util.Empty(err.Code)) && !tea.BoolValue(util.Empty(err.Message)) {
			// err 中含有 code 和 message 属性，可帮助开发定位问题
		}
	}
	return err
}

func (s *Service) getClient() (*dingtalkoauth2_1_0.Client, error) {
	config := &openapi.Config{}
	config.Protocol = tea.String("https")
	config.RegionId = tea.String("central")
	return dingtalkoauth2_1_0.NewClient(config)
}

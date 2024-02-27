package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/lutcoding/redbook/internal/domain"
	"net/http"
	"net/url"
)

type Service struct {
	appId     string
	appSecret string
}

func NewService(appId string, appSecret string) *Service {
	return &Service{
		appId:     appId,
		appSecret: appSecret,
	}
}

func (s *Service) AuthURL(ctx context.Context) (string, error) {
	var (
		redirectURI = url.PathEscape("https://meoying.com/oauth2/wechat/callback")
		urlPattern  = "https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect"
	)

	return fmt.Sprintf(urlPattern, s.appId, redirectURI, uuid.NewString()), nil
}

func (s *Service) VerifyCode(ctx context.Context, code string, state string) (domain.Wechat, error) {
	var urlPattern = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
	targetURL := fmt.Sprintf(urlPattern, s.appId, s.appSecret, code)
	resp, err := http.Get(targetURL)
	if err != nil {
		return domain.Wechat{}, err
	}
	var res Result
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return domain.Wechat{}, err
	}
	if res.ErrCode != 0 {
		return domain.Wechat{},
			fmt.Errorf("微信返回错误响应，错误码：%d，错误信息：%s", res.ErrCode, res.ErrMsg)
	}
	return domain.Wechat{OpenID: res.OpenID, UnionID: res.UnionID}, err
}

type Result struct {
	ErrCode int64  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`

	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`

	OpenID  string `json:"openid"`
	Scope   string `json:"scope"`
	UnionID string `json:"unionid"`
}

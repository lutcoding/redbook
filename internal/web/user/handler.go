package user

import (
	regexp "github.com/dlclark/regexp2"
	"github.com/lutcoding/redbook/internal/service"
	"github.com/lutcoding/redbook/internal/web/jwt"
)

const (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	biz               = "login"
)

type Handler struct {
	svc    *service.UserService
	smsSvc *service.CodeService
	// 预编译正则表达式匹配邮箱格式
	emailRegexExp *regexp.Regexp
	jwtHdl        *jwt.Handler
}

func New(userSvc *service.UserService, smsSvc *service.CodeService, jwt *jwt.Handler) *Handler {
	return &Handler{
		emailRegexExp: regexp.MustCompile(emailRegexPattern, regexp.None),
		svc:           userSvc,
		smsSvc:        smsSvc,
		jwtHdl:        jwt,
	}
}

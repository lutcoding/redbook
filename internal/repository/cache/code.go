package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/lutcoding/redbook/common/globalkey"
	"github.com/redis/go-redis/v9"
)

var (
	ErrCodeSendTooFrequent    = errors.New("send code too frequent")
	ErrCodeVerifyTooManyTimes = errors.New("验证次数太多")
	ErrUnknown                = errors.New("unknown error")
)

type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

//go:embed "lua/set_code.lua"
var luaSetCode string

//go:embed "lua/verify_code.lua"
var luaVerifyCode string

type CodeRedisCache struct {
	client redis.Cmdable
}

func NewCodeRedisCache(client redis.Cmdable) *CodeRedisCache {
	return &CodeRedisCache{
		client: client,
	}
}

func (c *CodeRedisCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := c.client.Eval(ctx, luaSetCode, []string{c.key(biz, phone)}, code).Int()
	fmt.Println(c.key(biz, phone))
	if err != nil {
		return err
	}
	switch res {
	case 0:
		return nil
	case -1:
		return ErrCodeSendTooFrequent
	default:
		return errors.New("server internal error")
	}
}

func (c *CodeRedisCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	res, err := c.client.Eval(ctx, luaVerifyCode, []string{c.key(biz, phone)}, inputCode).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case 0:
		return true, nil
	case -1:
		return false, ErrCodeVerifyTooManyTimes
	case -2:
		return false, nil
	default:
		return false, ErrUnknown
	}
}

func (c *CodeRedisCache) key(biz, phone string) string {
	return fmt.Sprintf("%s%s:%s", globalkey.PhoneCodeCachePrefix, biz, phone)
}

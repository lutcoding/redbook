package service

import (
	"context"
	"errors"
	"github.com/lutcoding/redbook/internal/domain"
	"github.com/lutcoding/redbook/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserDuplicate          = repository.ErrUserDuplicate
	ErrInvalidEmailOrPassword = errors.New("invalid email or password")
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) SignUp(ctx context.Context, u domain.User) error {
	password, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(password)
	err = svc.repo.Create(ctx, u)
	if errors.Is(err, repository.ErrUserDuplicate) {
		return ErrUserDuplicate
	}
	return err
}

func (svc *UserService) Login(ctx context.Context, u domain.User) (domain.User, error) {
	user, err := svc.repo.FindByEmail(ctx, u.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return domain.User{}, ErrInvalidEmailOrPassword
		}
		return domain.User{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password))
	if err != nil {
		return domain.User{}, ErrInvalidEmailOrPassword
	}
	return user, nil
}

func (svc *UserService) Profile(ctx context.Context, id int64) (domain.User, error) {
	user, err := svc.repo.FindById(ctx, id)
	return user, err
}

func (svc *UserService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	// 快路径 find比insert快
	user, err := svc.repo.FindByPhone(ctx, phone)
	if !errors.Is(err, repository.ErrUserNotFound) {
		// 绝大部分请求会进入
		// err == nil or err != ErrUserNotFound
		return user, err
	}
	// 慢路径
	// 在系统资源不租, 触发降级后, 不执行慢路径
	//if ctx.Value("降级") == "true" {
	//	return domain.User{}, errors.New("system degraded")
	//}
	err = svc.repo.Create(ctx, domain.User{Phone: phone})
	// 因为这里可能出现并发问题 所以要判断一下 err != repository.ErrUserDuplicate
	if err != nil && err != repository.ErrUserDuplicate {
		return domain.User{}, err
	}
	return svc.repo.FindByPhone(ctx, phone)
}

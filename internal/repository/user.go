package repository

import (
	"context"
	"database/sql"
	"github.com/lutcoding/redbook/internal/domain"
	"github.com/lutcoding/redbook/internal/repository/cache"
	"github.com/lutcoding/redbook/internal/repository/dao"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFound
)

type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	Update(ctx context.Context, u domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindById(ctx context.Context, id int64) (domain.User, error)
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
}

type UserCacheRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserCacheRepository(dao dao.UserDAO, cache cache.UserCache) *UserCacheRepository {
	return &UserCacheRepository{
		dao:   dao,
		cache: cache,
	}
}

func (r *UserCacheRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(u))
}

func (r *UserCacheRepository) Update(ctx context.Context, u domain.User) error {
	return r.dao.Update(ctx, r.domainToEntity(u))
}

func (r *UserCacheRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *UserCacheRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	user, err := r.cache.Get(ctx, id)
	switch err {
	case nil:
		return user, nil
	case cache.ErrKeyNotExist:
		u, err := r.dao.FindById(ctx, id)
		if err != nil {
			return domain.User{}, err
		}
		user = r.entityToDomain(u)
		err = r.cache.Set(ctx, user)
		if err != nil {
			// 打日志 做监控
		}
		return user, nil
	default:
		// err == io.EOF或其他

		// 实际面试：比如redis崩了
		return domain.User{}, err
	}
}

func (r *UserCacheRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *UserCacheRepository) entityToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email.String,
		Phone:    u.Phone.String,
		Password: u.Password,
	}
}

func (r *UserCacheRepository) domainToEntity(u domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password: u.Password,
	}
}

package service

import (
	"context"
	"errors"
	"github.com/lutcoding/redbook/internal/domain"
	"github.com/lutcoding/redbook/internal/repository"
	mock_repository "github.com/lutcoding/redbook/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func TestLogin(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.UserRepository

		user domain.User

		wantUser domain.User
		wantErr  error
	}{
		{
			name: "login success",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := mock_repository.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "6668964@qq.com").
					Return(domain.User{
						Email:    "6668964@qq.com",
						Password: "$2a$10$e4o/td.cvKE3sj0WYxKmB.hUly6nCjQoxB070nRvu63XCp44FFZmy",
					}, nil)
				return repo
			},
			user: domain.User{
				Email:    "6668964@qq.com",
				Password: "123456",
			},
			wantUser: domain.User{
				Email:    "6668964@qq.com",
				Password: "$2a$10$e4o/td.cvKE3sj0WYxKmB.hUly6nCjQoxB070nRvu63XCp44FFZmy",
			},
			wantErr: nil,
		},
		{
			name: "user not found",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := mock_repository.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "6668964@qq.com").
					Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},
			user: domain.User{
				Email:    "6668964@qq.com",
				Password: "123456",
			},
			wantUser: domain.User{},
			wantErr:  ErrInvalidEmailOrPassword,
		},
		{
			name: "err invalid password",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := mock_repository.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "6668964@qq.com").
					Return(domain.User{
						Email:    "6668964@qq.com",
						Password: "123456",
					}, nil)
				return repo
			},
			user: domain.User{
				Email:    "6668964@qq.com",
				Password: "123456",
			},
			wantUser: domain.User{},
			wantErr:  ErrInvalidEmailOrPassword,
		},
		{
			name: "db sys error",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := mock_repository.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "6668964@qq.com").
					Return(domain.User{}, errors.New("db error"))
				return repo
			},
			user: domain.User{
				Email:    "6668964@qq.com",
				Password: "123456",
			},
			wantUser: domain.User{},
			wantErr:  errors.New("db error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewUserService(tc.mock(ctrl))
			user, err := svc.Login(context.Background(), tc.user)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

func TestEncrypt(t *testing.T) {
	password, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	t.Log(string(password))
}

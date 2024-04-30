package service

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"lottery_single/internal/model"
	"lottery_single/internal/pkg/constant"
	"lottery_single/internal/pkg/middlewares/gormcli"
	"lottery_single/internal/pkg/middlewares/log"
	"lottery_single/internal/pkg/utils"
	"lottery_single/internal/repo"
)

// UserService 用户功能
type UserService interface {
	Login(ctx context.Context, userName, passWord string) (*LoginRsp, error)
	Register(ctx context.Context, user *model.User) error
}

type userService struct {
	userReop *repo.UserRepo
}

var userServiceImpl *userService

func NewUserService() {
	userServiceImpl = &userService{
		userReop: repo.NewUserRepo(),
	}
}

func GetUserService() UserService {
	return userServiceImpl
}

func (p *userService) Login(ctx context.Context, userName, passWord string) (*LoginRsp, error) {
	info, err := p.userReop.GetByName(gormcli.GetDB(), userName)
	if err != nil {
		return nil, err
	}
	log.InfoContextf(ctx, "info is: +%v\n", info)
	log.InfoContextf(ctx, "info.Password=%s,passWord=%s\n", info.Password, passWord)
	// 验证密码是否正确
	err = bcrypt.CompareHashAndPassword([]byte(info.Password), []byte(passWord))
	if err != nil {
		return nil, fmt.Errorf("password error: %v", err)
	}
	token, err := utils.GenerateJwtToken(constant.SecretKey, constant.Issuer, info.Id, userName)
	if err != nil {
		return nil, err
	}
	response := &LoginRsp{
		UserID: info.Id,
		Token:  token,
	}
	return response, nil
}
func (p *userService) Register(ctx context.Context, user *model.User) error {
	// Since the method now expects a *model.User, you don't need to create a new user instance
	// inside this method. Instead, you can directly use the provided user object.

	// Hash the password if it's not already hashed
	if !isHashed(user.Password) {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	}

	// Save the user to the database
	err := p.userReop.CreateUser(gormcli.GetDB(), user)
	if err != nil {
		return err
	}

	return nil
}
func isHashed(password string) bool {

	return len(password) > 5
}

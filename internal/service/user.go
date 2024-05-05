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
	// 检查密码是否为空或长度不足
	if len(passWord) < 6 {
		return nil, fmt.Errorf("password is too short")
	}

	log.Infof("Attempting to log in user: %s", userName)

	info, err := p.userReop.GetByName(gormcli.GetDB(), userName)
	if err != nil {
		log.Errorf("Error fetching user: %v", err)
		return nil, err
	}

	log.Infof("Retrieved user info: %+v", info)

	// 验证密码是否正确
	err = bcrypt.CompareHashAndPassword([]byte(info.Password), []byte(passWord))
	if err != nil {
		log.Errorf("Password verification failed: %v", err)
		return nil, fmt.Errorf("password error: %v", err)
	}

	log.Infof("Password verified successfully for user: %s", userName)

	// 生成 JWT Token
	token, err := utils.GenerateJwtToken(constant.SecretKey, constant.Issuer, info.Id, userName)
	if err != nil {
		log.Errorf("Error generating JWT token: %v", err)
		return nil, err
	}

	response := &LoginRsp{
		UserID: info.Id,
		Token:  token,
	}

	log.Infof("User %s logged in successfully", userName)

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

	return len(password) > 1
}

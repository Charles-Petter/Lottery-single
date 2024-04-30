package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"lottery_single/internal/model"
	"lottery_single/internal/pkg/middlewares/cache"
	"lottery_single/internal/pkg/middlewares/log"
	"strconv"
)

type UserRepo struct {
}

func NewUserRepo() *UserRepo {
	return &UserRepo{}
}

func (r *UserRepo) Get(db *gorm.DB, id uint) (*model.User, error) {
	// 优先从缓存获取
	User, err := r.GetFromCache(id)
	if err == nil && User != nil {
		return User, nil
	}
	User = &model.User{
		Id: id,
	}
	err = db.Model(&model.User{}).First(User).Error
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return nil, nil
		}
		return nil, fmt.Errorf("UserRepo|Get:%v", err)
	}
	return User, nil
}

func (r *UserRepo) GetByName(db *gorm.DB, userName string) (*model.User, error) {
	// 优先从缓存获取
	user := &model.User{
		UserName: userName,
	}
	err := db.Model(&model.User{}).First(user).Error
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return nil, nil
		}
		return nil, fmt.Errorf("UserRepo|GetByName:%v", err)
	}
	return user, nil
}

func (r *UserRepo) GetAll(db *gorm.DB) ([]*model.User, error) {
	var Users []*model.User
	err := db.Model(&model.User{}).Where("").Order("sys_updated desc").Find(&Users).Error
	if err != nil {
		return nil, fmt.Errorf("UserRepo|GetAll:%v", err)
	}
	return Users, nil
}

func (r *UserRepo) CountAll(db *gorm.DB) (int64, error) {
	var num int64
	err := db.Model(&model.User{}).Count(&num).Error
	if err != nil {
		return 0, fmt.Errorf("UserRepo|CountAll:%v", err)
	}
	return num, nil
}

func (r *UserRepo) Create(db *gorm.DB, user *model.User) error {
	err := db.Model(&model.User{}).Create(user).Error
	if err != nil {
		return fmt.Errorf("UserRepo|Create:%v", err)
	}
	return nil
}

func (r *UserRepo) Delete(db *gorm.DB, id uint) error {
	User := &model.User{Id: id}
	if err := db.Model(&model.User{}).Delete(User).Error; err != nil {
		return fmt.Errorf("UserRepo|Delete:%v")
	}
	return nil
}

func (r *UserRepo) Update(db *gorm.DB, user *model.User, cols ...string) error {
	var err error
	if len(cols) == 0 {
		err = db.Model(user).Updates(user).Error
	} else {
		err = db.Model(user).Select(cols).Updates(user).Error
	}
	if err != nil {
		return fmt.Errorf("UserRepo|Update:%v", err)
	}
	return nil
}

// GetFromCache 根据id从缓存获取奖品
func (r *UserRepo) GetFromCache(id uint) (*model.User, error) {
	redisCli := cache.GetRedisCli()
	idStr := strconv.FormatUint(uint64(id), 10)
	ret, exist, err := redisCli.Get(context.Background(), idStr)
	if err != nil {
		log.Errorf("UserRepo|GetFromCache:" + err.Error())
		return nil, err
	}

	if !exist {
		return nil, nil
	}

	User := model.User{}
	json.Unmarshal([]byte(ret), &model.User{})

	return &User, nil
}
func (r *UserRepo) CreateUser(db *gorm.DB, user *model.User) error {
	err := db.Model(&model.User{}).Create(user).Error
	if err != nil {
		return fmt.Errorf("UserRepo|CreateUser:%v", err)
	}
	return nil
}

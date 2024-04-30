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

type ResultRepo struct {
}

func NewResultRepo() *ResultRepo {
	return &ResultRepo{}
}

func (r *ResultRepo) Get(db *gorm.DB, id uint) (*model.Result, error) {
	// 优先从缓存获取
	result, err := r.GetFromCache(id)
	if err == nil && result != nil {
		return result, nil
	}
	result = &model.Result{
		Id: id,
	}
	err = db.Model(&model.Result{}).First(result).Error
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return nil, nil
		}
		return nil, fmt.Errorf("ResultRepo|Get:%v", err)
	}
	return result, nil
}

func (r *ResultRepo) GetAll(db *gorm.DB) ([]*model.Result, error) {
	var results []*model.Result
	err := db.Model(&model.Result{}).Where("").Order("sys_updated desc").Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("ResultRepo|GetAll:%v", err)
	}
	return results, nil
}

func (r *ResultRepo) CountAll(db *gorm.DB) (int64, error) {
	var num int64
	err := db.Model(&model.Result{}).Count(&num).Error
	if err != nil {
		return 0, fmt.Errorf("ResultRepo|CountAll:%v", err)
	}
	return num, nil
}

func (r *ResultRepo) Create(db *gorm.DB, result *model.Result) error {
	err := db.Model(&model.Result{}).Create(result).Error
	if err != nil {
		return fmt.Errorf("ResultRepo|Create:%v", err)
	}
	return nil
}

func (r *ResultRepo) Delete(db *gorm.DB, id uint) error {
	result := &model.Result{Id: id}
	if err := db.Model(&model.Result{}).Delete(result).Error; err != nil {
		return fmt.Errorf("ResultRepo|Delete:%v")
	}
	return nil
}

func (r *ResultRepo) Update(db *gorm.DB, result *model.Result, cols ...string) error {
	var err error
	if len(cols) == 0 {
		err = db.Model(result).Updates(result).Error
	} else {
		err = db.Model(result).Select(cols).Updates(result).Error
	}
	if err != nil {
		return fmt.Errorf("ResultRepo|Update:%v", err)
	}
	return nil
}

// GetFromCache 根据id从缓存获取奖品
func (r *ResultRepo) GetFromCache(id uint) (*model.Result, error) {
	redisCli := cache.GetRedisCli()
	idStr := strconv.FormatUint(uint64(id), 10)
	ret, exist, err := redisCli.Get(context.Background(), idStr)
	if err != nil {
		log.Errorf("ResultRepo|GetFromCache:" + err.Error())
		return nil, err
	}

	if !exist {
		return nil, nil
	}

	result := model.Result{}
	json.Unmarshal([]byte(ret), &model.Result{})

	return &result, nil
}

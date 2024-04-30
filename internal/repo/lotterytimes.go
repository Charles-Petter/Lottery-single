package repo

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"lottery_single/internal/model"
	"lottery_single/internal/pkg/constant"
	"lottery_single/internal/pkg/middlewares/cache"
	"lottery_single/internal/pkg/middlewares/log"
	"math"
)

type LotteryTimesRepo struct {
	Db       *gorm.DB
	RedisCli *cache.Client
}

func NewLotteryTimesRepo(db *gorm.DB, redisCli *cache.Client) *LotteryTimesRepo {
	return &LotteryTimesRepo{
		Db:       db,
		RedisCli: redisCli,
	}
}

func (r *LotteryTimesRepo) Get(id uint) (*model.LotteryTimes, error) {
	lotteryTimes := &model.LotteryTimes{
		Id: id,
	}
	err := r.Db.Model(&model.LotteryTimes{}).First(lotteryTimes).Error
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return nil, nil
		}
		return nil, fmt.Errorf("LotteryTimesRepo|Get:%v", err)
	}
	return lotteryTimes, nil
}

func (r *LotteryTimesRepo) GetByUserIDAndDay(uid uint, day uint) (*model.LotteryTimes, error) {
	lotteryTimes := &model.LotteryTimes{
		UserId: uid,
		Day:    day,
	}
	err := r.Db.Model(&model.LotteryTimes{}).First(lotteryTimes).Error
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return nil, nil
		}
		return nil, fmt.Errorf("LotteryTimesRepo|GetByUserID:%v", err)
	}
	return lotteryTimes, nil
}

func (r *LotteryTimesRepo) GetAll() ([]*model.LotteryTimes, error) {
	var lotteryTimesList []*model.LotteryTimes
	err := r.Db.Model(&model.LotteryTimes{}).Where("").Order("sys_updated desc").Find(&lotteryTimesList).Error
	if err != nil {
		return nil, fmt.Errorf("LotteryTimesRepo|GetAll:%v", err)
	}
	return lotteryTimesList, nil
}

func (r *LotteryTimesRepo) CountAll() (int64, error) {
	var num int64
	err := r.Db.Model(&model.LotteryTimes{}).Count(&num).Error
	if err != nil {
		return 0, fmt.Errorf("LotteryTimesRepo|CountAll:%v", err)
	}
	return num, nil
}

func (r *LotteryTimesRepo) Create(lotteryTimes *model.LotteryTimes) error {
	err := r.Db.Model(&model.LotteryTimes{}).Create(lotteryTimes).Error
	if err != nil {
		return fmt.Errorf("LotteryTimesRepo|Create:%v", err)
	}
	return nil
}

func (r *LotteryTimesRepo) Delete(id uint) error {
	lotteryTimes := &model.LotteryTimes{Id: id}
	if err := r.Db.Model(&model.LotteryTimes{}).Delete(lotteryTimes).Error; err != nil {
		return fmt.Errorf("LotteryTimesRepo|Delete:%v")
	}
	return nil
}

func (r *LotteryTimesRepo) Update(lotteryTimes *model.LotteryTimes, cols ...string) error {
	var err error
	if len(cols) == 0 {
		err = r.Db.Model(lotteryTimes).Updates(lotteryTimes).Error
	} else {
		err = r.Db.Model(lotteryTimes).Select(cols).Updates(lotteryTimes).Error
	}
	if err != nil {
		return fmt.Errorf("LotteryTimesRepo|Update:%v", err)
	}
	return nil
}

// IncrUserDayLotteryNum 每天缓存的用户抽奖次数递增，返回递增后的数值
func (r *LotteryTimesRepo) IncrUserDayLotteryNum(uid uint) int64 {
	i := uid % constant.UserFrameSize
	// 集群的redis统计数递增
	key := fmt.Sprintf(constant.UserLotteryDayNumPrefix+"%d", i)
	ret, err := cache.GetRedisCli().HIncrBy(context.Background(), key, fmt.Sprint(uid), 1)
	if err != nil {
		log.Errorf("LotteryTimesRepo|IncrUserDayLotteryNum:%v", err)
		return math.MaxInt32
	}
	return ret
}

// InitUserLuckyNum 从给定的数据直接初始化用户的参与抽奖次数
func (r *LotteryTimesRepo) InitUserLuckyNum(uid uint, num int64) error {
	if num <= 1 {
		return nil
	}
	i := uid % constant.UserFrameSize
	key := fmt.Sprintf(constant.UserLotteryDayNumPrefix+"%d", i)
	_, err := cache.GetRedisCli().HSet(context.Background(), key, fmt.Sprint(uid), num)
	if err != nil {
		log.Errorf("LotteryTimesRepo|InitUserLuckyNum:%v", err)
		return fmt.Errorf("LotteryTimesRepo|InitUserLuckyNum:%v", err)
	}
	return nil
}

package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"lottery_single/internal/model"
	"lottery_single/internal/pkg/constant"
	"lottery_single/internal/pkg/middlewares/cache"
	"lottery_single/internal/pkg/middlewares/log"
	"lottery_single/internal/pkg/utils"
	"strconv"
)

type BlackUserRepo struct {
	Db       *gorm.DB
	RedisCli *cache.Client
}

func NewBlackUserRepo() *BlackUserRepo {
	return &BlackUserRepo{}
}

//func (r *BlackUserRepo) Get(db *gorm.DB, id uint) (*model.BlackUser, error) {
//	// 优先从缓存获取
//	blackUser, err := r.GetFromCache(id)
//	if err == nil && blackUser != nil {
//		return blackUser, nil
//	}
//	blackUser = &model.BlackUser{
//		Id: id,
//	}
//	err = db.Model(&model.BlackUser{}).First(blackUser).Error
//	if err != nil {
//		if err.Error() == gorm.ErrRecordNotFound.Error() {
//			return nil, nil
//		}
//		return nil, fmt.Errorf("BlackUserRepo|Get:%v", err)
//	}
//	return blackUser, nil
//}

func (r *BlackUserRepo) GetByUserID(db *gorm.DB, uid uint) (*model.BlackUser, error) {
	blackUser := &model.BlackUser{
		UserId: uid,
	}
	err := db.Model(&model.BlackUser{}).First(blackUser).Error
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return nil, nil
		}
		return nil, fmt.Errorf("BlackUserRepo|Get:%v", err)
	}
	return blackUser, nil
}

func (r *BlackUserRepo) GetByUserIDWithCache(db *gorm.DB, uid uint) (*model.BlackUser, error) {
	// 优先从缓存获取
	blackUser, err := r.GetByCache(uid)
	// 从缓存获取到用户
	if err == nil && blackUser != nil {
		return blackUser, nil
	}
	// 缓存没有获取到黑明单用户
	blackUser = &model.BlackUser{
		UserId: uid,
	}
	err = db.Model(&model.BlackUser{}).First(blackUser).Error
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return nil, nil
		}
		return nil, fmt.Errorf("BlackUserRepo|Get:%v", err)
	}
	// db获取到了黑明单用户，同步到缓存中
	if err = r.SetByCache(blackUser); err != nil {
		return nil, fmt.Errorf("BlackUserRepo|SetByCache:%v", err)
	}
	return blackUser, nil
}

func (r *BlackUserRepo) GetAll(db *gorm.DB) ([]*model.BlackUser, error) {
	var BlackUsers []*model.BlackUser
	err := db.Model(&model.BlackUser{}).Where("").Order("sys_updated desc").Find(&BlackUsers).Error
	if err != nil {
		return nil, fmt.Errorf("BlackUserRepo|GetAll:%v", err)
	}
	return BlackUsers, nil
}

func (r *BlackUserRepo) CountAll(db *gorm.DB) (int64, error) {
	var num int64
	err := db.Model(&model.BlackUser{}).Count(&num).Error
	if err != nil {
		return 0, fmt.Errorf("BlackUserRepo|CountAll:%v", err)
	}
	return num, nil
}

func (r *BlackUserRepo) Create(db *gorm.DB, BlackUser *model.BlackUser) error {
	err := db.Model(&model.BlackUser{}).Create(BlackUser).Error
	if err != nil {
		return fmt.Errorf("BlackUserRepo|Create:%v", err)
	}
	return nil
}

func (r *BlackUserRepo) Delete(db *gorm.DB, id uint) error {
	BlackUser := &model.BlackUser{Id: id}
	if err := db.Model(&model.BlackUser{}).Delete(BlackUser).Error; err != nil {
		return fmt.Errorf("BlackUserRepo|Delete:%v")
	}
	return nil
}

func (r *BlackUserRepo) DeleteWithCache(db *gorm.DB, uid uint) error {
	blackUser := &model.BlackUser{UserId: uid}
	if err := r.UpdateByCache(blackUser); err != nil {
		return fmt.Errorf("BlackUserRepo|DeleteWithCache:%v", err)
	}
	if err := db.Model(&model.BlackUser{}).Delete(blackUser).Error; err != nil {
		return fmt.Errorf("BlackUserRepo|Delete:%v")
	}
	return nil
}

func (r *BlackUserRepo) Update(db *gorm.DB, userID uint, blackUser *model.BlackUser, cols ...string) error {
	var err error
	if len(cols) == 0 {
		err = db.Model(blackUser).Where("user_id=?", userID).Updates(blackUser).Error
	} else {
		err = db.Model(blackUser).Where("user_id=?", userID).Select(cols).Updates(blackUser).Error
	}
	if err != nil {
		return fmt.Errorf("BlackUserRepo|Update:%v", err)
	}
	return nil
}

func (r *BlackUserRepo) UpdateWithCache(db *gorm.DB, userID uint, blackUser *model.BlackUser, cols ...string) error {
	if err := r.UpdateByCache(&model.BlackUser{UserId: userID}); err != nil {
		return fmt.Errorf("BlackUserRepo|DeleteWithCache:%v", err)
	}
	var err error
	if len(cols) == 0 {
		err = db.Model(blackUser).Where("user_id=?", userID).Updates(blackUser).Error
	} else {
		err = db.Model(blackUser).Where("user_id=?", userID).Select(cols).Updates(blackUser).Error
	}
	if err != nil {
		return fmt.Errorf("BlackUserRepo|Update:%v", err)
	}
	return nil
}

// GetFromCache 根据id从缓存获取奖品
func (r *BlackUserRepo) GetFromCache(id uint) (*model.BlackUser, error) {
	redisCli := cache.GetRedisCli()
	idStr := strconv.FormatUint(uint64(id), 10)
	ret, exist, err := redisCli.Get(context.Background(), idStr)
	if err != nil {
		log.Errorf("BlackUserRepo|GetFromCache:" + err.Error())
		return nil, err
	}

	if !exist {
		return nil, nil
	}

	BlackUser := model.BlackUser{}
	json.Unmarshal([]byte(ret), &model.BlackUser{})

	return &BlackUser, nil
}

func (r *BlackUserRepo) GetByCache(uid uint) (*model.BlackUser, error) {
	key := fmt.Sprintf(constant.UserCacheKeyPrefix+"%d", uid)
	valueMap, err := cache.GetRedisCli().HGetAll(context.Background(), key)
	if err != nil {
		return nil, fmt.Errorf("BlackUserRepo|GetByCache:%v", err)
	}
	userIdStr := valueMap["UserId"]
	num, _ := strconv.Atoi(userIdStr)
	userID := uint(num)
	if userID <= 0 {
		return nil, nil
	}
	idStr := valueMap["Id"]
	id, _ := strconv.Atoi(idStr)
	blackUser := &model.BlackUser{
		Id:       uint(id),
		UserId:   userID,
		UserName: valueMap["UserName"],
		RealName: valueMap["RealName"],
		Mobile:   valueMap["Mobile"],
		Address:  valueMap["Address"],
		SysIp:    valueMap["SysIp"],
	}
	blackTime, err := utils.ParseTime(valueMap["BlackTime"])
	if err != nil {
		return nil, fmt.Errorf("BlackUserRepo|GetByCache:%v", err)
	}
	blackUser.BlackTime = blackTime
	sysCreated, err := utils.ParseTime(valueMap["SysCreated"])
	if err != nil {
		return nil, fmt.Errorf("BlackUserRepo|GetByCache:%v", err)
	}
	blackUser.SysCreated = &sysCreated
	sysUpdated, err := utils.ParseTime(valueMap["SysUpdated"])
	if err != nil {
		return nil, fmt.Errorf("BlackUserRepo|GetByCache:%v", err)
	}
	blackUser.SysUpdated = &sysUpdated
	return blackUser, nil
}

func (r *BlackUserRepo) SetByCache(blackUser *model.BlackUser) error {
	if blackUser == nil || blackUser.UserId <= 0 {
		return fmt.Errorf("BlackUserRepo|SetByCache invalid user")
	}
	key := fmt.Sprintf(constant.UserCacheKeyPrefix+"%d", blackUser.UserId)
	valueMap := make(map[string]interface{})
	valueMap["Id"] = strconv.Itoa(int(blackUser.Id))
	valueMap["UserId"] = strconv.Itoa(int(blackUser.UserId))
	valueMap["UserName"] = blackUser.UserName
	valueMap["BlackTime"] = utils.FormatFromUnixTime(blackUser.BlackTime.Unix())
	valueMap["RealName"] = blackUser.RealName
	valueMap["Mobile"] = blackUser.Mobile
	valueMap["Address"] = blackUser.Address
	valueMap["SysCreated"] = utils.FormatFromUnixTime(blackUser.SysCreated.Unix())
	valueMap["SysUpdated"] = utils.FormatFromUnixTime(blackUser.SysUpdated.Unix())
	valueMap["SysIp"] = blackUser.SysIp
	_, err := cache.GetRedisCli().HMSet(context.Background(), key, valueMap)
	if err != nil {
		fmt.Errorf("BlackUserRepo|SetByCache invalid user")
	}
	return nil
}

func (r *BlackUserRepo) UpdateByCache(blackUser *model.BlackUser) error {
	if blackUser == nil || blackUser.UserId <= 0 {
		return fmt.Errorf("BlackUserRepo|UpdateByCache invalid blackUser")
	}
	key := fmt.Sprintf(constant.UserCacheKeyPrefix+"%d", blackUser.UserId)
	if err := cache.GetRedisCli().Delete(context.Background(), key); err != nil {
		return fmt.Errorf("BlackUserRepo|UpdateByCache:%v", err)
	}
	return nil
}

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

type BlackIpRepo struct {
	Db       *gorm.DB
	RedisCli *cache.Client
}

func NewBlackIpRepo() *BlackIpRepo {
	return &BlackIpRepo{}
}

func (r *BlackIpRepo) Get(db *gorm.DB, id uint) (*model.BlackIp, error) {
	// 优先从缓存获取
	BlackIp, err := r.GetFromCache(id)
	if err == nil && BlackIp != nil {
		return BlackIp, nil
	}
	BlackIp = &model.BlackIp{
		Id: id,
	}
	err = db.Model(&model.BlackIp{}).First(BlackIp).Error
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return nil, nil
		}
		return nil, fmt.Errorf("BlackIpRepo|Get:%v", err)
	}
	return BlackIp, nil
}

func (r *BlackIpRepo) GetByIP(db *gorm.DB, ip string) (*model.BlackIp, error) {
	// 优先从缓存获取
	blackIP := &model.BlackIp{
		Ip: ip,
	}
	err := db.Model(&model.BlackIp{}).Where("ip = ?", ip).First(blackIP).Error
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return nil, nil
		}
		return nil, fmt.Errorf("BlackIpRepo|GetByIp:%v", err)
	}
	return blackIP, nil
}

func (r *BlackIpRepo) GetByIPWithCache(db *gorm.DB, ip string) (*model.BlackIp, error) {
	// 优先从缓存获取
	blackIp, err := r.GetByCache(ip)
	// 从缓存获取到IP
	if err == nil && blackIp != nil {
		return blackIp, nil
	}
	// 缓存中没有获取到ip
	blackIP := &model.BlackIp{
		Ip: ip,
	}
	err = db.Model(&model.BlackIp{}).Where("ip = ?", ip).First(blackIP).Error
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return nil, nil
		}
		return nil, fmt.Errorf("BlackIpRepo|GetByIp:%v", err)
	}
	// 数据库中正确读到数据，设置到缓存中
	if err = r.SetByCache(blackIP); err != nil {
		return nil, fmt.Errorf("BlackIpRepo|SetByCache:%v", err)
	}
	return blackIP, nil
}

func (r *BlackIpRepo) GetAll(db *gorm.DB) ([]*model.BlackIp, error) {
	var BlackIps []*model.BlackIp
	err := db.Model(&model.BlackIp{}).Where("").Order("sys_updated desc").Find(&BlackIps).Error
	if err != nil {
		return nil, fmt.Errorf("BlackIpRepo|GetAll:%v", err)
	}
	return BlackIps, nil
}

func (r *BlackIpRepo) CountAll(db *gorm.DB) (int64, error) {
	var num int64
	err := db.Model(&model.BlackIp{}).Count(&num).Error
	if err != nil {
		return 0, fmt.Errorf("BlackIpRepo|CountAll:%v", err)
	}
	return num, nil
}

func (r *BlackIpRepo) Create(db *gorm.DB, BlackIp *model.BlackIp) error {
	err := db.Model(&model.BlackIp{}).Create(BlackIp).Error
	if err != nil {
		return fmt.Errorf("BlackIpRepo|Create:%v", err)
	}
	return nil
}

func (r *BlackIpRepo) Delete(db *gorm.DB, id uint) error {
	BlackIp := &model.BlackIp{Id: id}
	if err := db.Model(&model.BlackIp{}).Delete(BlackIp).Error; err != nil {
		return fmt.Errorf("BlackIpRepo|Delete:%v")
	}
	return nil
}

func (r *BlackIpRepo) DeleteWithCache(db *gorm.DB, id uint) error {
	blackIp := &model.BlackIp{Id: id}
	if err := db.Model(&model.BlackIp{}).Delete(blackIp).Error; err != nil {
		return fmt.Errorf("BlackIpRepo|Delete:%v")
	}
	return nil
}

func (r *BlackIpRepo) Update(db *gorm.DB, ip string, blackIp *model.BlackIp, cols ...string) error {
	if err := r.UpdateByCache(&model.BlackIp{Ip: ip}); err != nil {
		return fmt.Errorf("BlackIpRepo|UpdateWithCache:%v", err)
	}
	var err error
	if len(cols) == 0 {
		err = db.Model(blackIp).Where("ip=?", ip).Updates(blackIp).Error
	} else {
		err = db.Model(blackIp).Where("ip=?", ip).Select(cols).Updates(blackIp).Error
	}
	if err != nil {
		return fmt.Errorf("BlackIpRepo|Update:%v", err)
	}
	return nil
}

func (r *BlackIpRepo) UpdateWithCache(db *gorm.DB, ip string, blackIp *model.BlackIp, cols ...string) error {
	if err := r.UpdateByCache(&model.BlackIp{Ip: ip}); err != nil {
		return fmt.Errorf("BlackIpRepo|UpdateWithCache:%v", err)
	}
	var err error
	if len(cols) == 0 {
		err = db.Model(blackIp).Where("ip=?", ip).Updates(blackIp).Error
	} else {
		err = db.Model(blackIp).Where("ip=?", ip).Select(cols).Updates(blackIp).Error
	}
	if err != nil {
		return fmt.Errorf("BlackIpRepo|Update:%v", err)
	}
	return nil
}

// GetFromCache 根据id从缓存获取奖品
func (r *BlackIpRepo) GetFromCache(id uint) (*model.BlackIp, error) {
	redisCli := cache.GetRedisCli()
	idStr := strconv.FormatUint(uint64(id), 10)
	ret, exist, err := redisCli.Get(context.Background(), idStr)
	if err != nil {
		log.Errorf("BlackIpRepo|GetFromCache:" + err.Error())
		return nil, err
	}

	if !exist {
		return nil, nil
	}

	BlackIp := model.BlackIp{}
	json.Unmarshal([]byte(ret), &model.BlackIp{})

	return &BlackIp, nil
}

func (s *BlackIpRepo) SetByCache(blackIp *model.BlackIp) error {
	if blackIp == nil || blackIp.Ip == "" {
		return fmt.Errorf("BlackIpRepo|SetByCache invalid user")
	}
	key := fmt.Sprintf(constant.IpCacheKeyPrefix+"%s", blackIp.Ip)
	valueMap := make(map[string]interface{})
	valueMap["Id"] = strconv.Itoa(int(blackIp.Id))
	valueMap["BlackTime"] = utils.FormatFromUnixTime(blackIp.BlackTime.Unix())
	valueMap["SysCreated"] = utils.FormatFromUnixTime(blackIp.SysCreated.Unix())
	valueMap["SysUpdated"] = utils.FormatFromUnixTime(blackIp.SysUpdated.Unix())
	valueMap["Ip"] = blackIp.Ip
	_, err := cache.GetRedisCli().HMSet(context.Background(), key, valueMap)
	if err != nil {
		fmt.Errorf("BlackUserRepo|SetByCache invalid user")
	}
	return nil
}

func (s *BlackIpRepo) GetByCache(ip string) (*model.BlackIp, error) {
	key := fmt.Sprintf(constant.IpCacheKeyPrefix+"%s", ip)
	valueMap, err := cache.GetRedisCli().HGetAll(context.Background(), key)
	if err != nil {
		return nil, fmt.Errorf("BlackIpRepo|GetByCache:%v", err)
	}
	idStr := valueMap["Id"]
	id, _ := strconv.Atoi(idStr)
	blackIp := &model.BlackIp{
		Id: uint(id),
		Ip: ip,
	}
	blackTime, err := utils.ParseTime(valueMap["BlackTime"])
	if err != nil {
		return nil, fmt.Errorf("BlackIpRepo|GetByCache:%v", err)
	}
	blackIp.BlackTime = blackTime
	sysCreated, err := utils.ParseTime(valueMap["SysCreated"])
	if err != nil {
		return nil, fmt.Errorf("BlackIpRepo|GetByCache:%v", err)
	}
	blackIp.SysCreated = &sysCreated
	sysUpdated, err := utils.ParseTime(valueMap["SysUpdated"])
	if err != nil {
		return nil, fmt.Errorf("BlackIpRepo|GetByCache:%v", err)
	}
	blackIp.SysUpdated = &sysUpdated
	return blackIp, nil
}

func (r *BlackIpRepo) UpdateByCache(blackIp *model.BlackIp) error {
	if blackIp == nil || blackIp.Ip == "" {
		return fmt.Errorf("BlackIpRepo|UpdateByCache invalid blackUser")
	}
	key := fmt.Sprintf(constant.UserCacheKeyPrefix+"%s", blackIp.Ip)
	if err := cache.GetRedisCli().Delete(context.Background(), key); err != nil {
		return fmt.Errorf("BlackIpRepo|UpdateByCache:%v", err)
	}
	return nil
}

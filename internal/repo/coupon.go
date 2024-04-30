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
	"strconv"
)

type CouponRepo struct {
}

func NewCouponRepo() *CouponRepo {
	return &CouponRepo{}
}

func (r *CouponRepo) Get(db *gorm.DB, id uint) (*model.Coupon, error) {
	// 优先从缓存获取
	coupon, err := r.GetFromCache(id)
	if err == nil && coupon != nil {
		return coupon, nil
	}
	coupon = &model.Coupon{
		Id: id,
	}
	err = db.Model(&model.Coupon{}).First(coupon).Error
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return nil, nil
		}
		return nil, fmt.Errorf("CouponRepo|Get:%v", err)
	}
	return coupon, nil
}

func (r *CouponRepo) GetAll(db *gorm.DB) ([]*model.Coupon, error) {
	var coupons []*model.Coupon
	err := db.Model(&model.Coupon{}).Order("sys_updated desc").Find(&coupons).Error
	if err != nil {
		return nil, fmt.Errorf("CouponRepo|GetAll:%v", err)
	}
	return coupons, nil
}

func (r *CouponRepo) GetCouponListByPrizeID(db *gorm.DB, prizeID uint) ([]*model.Coupon, error) {
	var coupons []*model.Coupon
	err := db.Model(&model.Coupon{}).Where("prize_id=?", prizeID).Order("id desc").Find(&coupons).Error
	if err != nil {
		return nil, fmt.Errorf("CouponRepo|GetAll:%v", err)
	}
	return coupons, nil
}

func (r *CouponRepo) CountAll(db *gorm.DB) (int64, error) {
	var num int64
	err := db.Model(&model.Coupon{}).Count(&num).Error
	if err != nil {
		return 0, fmt.Errorf("CouponRepo|CountAll:%v", err)
	}
	return num, nil
}

func (r *CouponRepo) Create(db *gorm.DB, coupon *model.Coupon) error {
	err := db.Model(&model.Coupon{}).Create(coupon).Error
	if err != nil {
		return fmt.Errorf("CouponRepo|Create:%v", err)
	}
	return nil
}

func (r *CouponRepo) Delete(db *gorm.DB, id uint) error {
	coupon := &model.Coupon{Id: id}
	if err := db.Model(&model.Coupon{}).Delete(coupon).Error; err != nil {
		return fmt.Errorf("CouponRepo|Delete:%v")
	}
	return nil
}

func (r *CouponRepo) Update(db *gorm.DB, coupon *model.Coupon, cols ...string) error {
	var err error
	if len(cols) == 0 {
		err = db.Model(coupon).Updates(coupon).Error
	} else {
		err = db.Model(coupon).Select(cols).Updates(coupon).Error
	}
	if err != nil {
		return fmt.Errorf("CouponRepo|Update:%v", err)
	}
	return nil
}

func (r *CouponRepo) UpdateByCode(db *gorm.DB, code string, coupon *model.Coupon, cols ...string) error {
	var err error
	if len(cols) == 0 {
		err = db.Model(coupon).Where("code = ?", code).Updates(coupon).Error
	} else {
		err = db.Model(coupon).Where("code = ?", code).Select(cols).Updates(coupon).Error
	}
	if err != nil {
		return fmt.Errorf("CouponRepo|Update:%v", err)
	}
	return nil
}

// GetFromCache 根据id从缓存获取奖品
func (r *CouponRepo) GetFromCache(id uint) (*model.Coupon, error) {
	redisCli := cache.GetRedisCli()
	idStr := strconv.FormatUint(uint64(id), 10)
	ret, exist, err := redisCli.Get(context.Background(), idStr)
	if err != nil {
		log.Errorf("CouponRepo|GetFromCache:" + err.Error())
		return nil, err
	}

	if !exist {
		return nil, nil
	}

	coupon := model.Coupon{}
	json.Unmarshal([]byte(ret), &model.Coupon{})

	return &coupon, nil
}

// GetGetNextUsefulCoupon 获取下一个可用编码的优惠券
func (r *CouponRepo) GetGetNextUsefulCoupon(db *gorm.DB, prizeID, couponID int) (*model.Coupon, error) {
	coupon := &model.Coupon{}
	err := db.Model(coupon).Where("prize_id=?", prizeID).Where("id > ?", couponID).
		Where("sys_status = ?", 1).First(coupon).Error
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return nil, nil
		}
		return nil, fmt.Errorf("CouponRepo|GetGetNextUsefulCoupon err:%v", err)
	}
	return coupon, nil
}

// ImportCacheCoupon 往缓存导入优惠券
func (r *CouponRepo) ImportCacheCoupon(prizeID uint, code string) (bool, error) {
	key := fmt.Sprintf(constant.PrizeCouponCacheKey+"%d", prizeID)
	cnt, err := cache.GetRedisCli().SAdd(context.Background(), key, code)
	if err != nil {
		return false, fmt.Errorf("CouponRepo|ImportCacheCoupon:%v", err)
	}
	if cnt == 0 {
		return false, nil
	}
	return true, nil
}

// ReSetCacheCoupon 根据库存优惠券重置优惠券缓存
func (r *CouponRepo) ReSetCacheCoupon(db *gorm.DB, prizeID uint) (int64, int64, error) {
	var successNum, failureNum int64 = 0, 0
	couponList, err := r.GetCouponListByPrizeID(db, prizeID)
	if err != nil {
		return 0, 0, fmt.Errorf("CouponRepo")
	}
	if couponList == nil || len(couponList) == 0 {
		return 0, 0, nil
	}
	key := fmt.Sprintf(constant.PrizeCouponCacheKey+"%d", prizeID)
	// 这里先用临时keu统计，在原key上统计的话，因为db里的数量可能变化，没有同部到缓存中，比如db里面减少了10条数据，如果在原key上增加，那么缓存就会多处10条数据，所以根据db全部统计完了之后，在覆盖
	tmpKey := "tmp_" + key
	for _, coupon := range couponList {
		code := coupon.Code
		if coupon.SysStatus == 1 {
			cnt, err := cache.GetRedisCli().SAdd(context.Background(), tmpKey, code)
			if err != nil {
				return 0, 0, fmt.Errorf("CouponRepo|ReSetCacheCoupon:%v", err)
			}
			if cnt <= 0 {
				failureNum++
			} else {
				successNum++
			}
		}
	}
	_, err = cache.GetRedisCli().Rename(context.Background(), tmpKey, key)
	if err != nil {
		return 0, 0, fmt.Errorf("CouponRepo|ReSetCacheCoupon:%v", err)
	}
	return successNum, failureNum, nil
}

// GetCacheCouponNum 获取缓存中的剩余优惠券数量以及数据库中的剩余优惠券数量
func (r *CouponRepo) GetCacheCouponNum(db *gorm.DB, prizeID uint) (int64, int64, error) {
	var dbNum, cacheNum int64 = 0, 0
	couponList, err := r.GetCouponListByPrizeID(db, prizeID)
	if err != nil {
		return 0, 0, fmt.Errorf("CouponRepo|GetCacheCouponNum:%v", err)
	}
	if couponList == nil {
		return 0, 0, nil
	}
	for _, coupon := range couponList {
		if coupon.SysStatus == 1 {
			dbNum++
		}
	}
	key := fmt.Sprintf(constant.PrizeCouponCacheKey+"%d", prizeID)
	cacheNum, err = cache.GetRedisCli().SCard(context.Background(), key)
	if err != nil {
		return 0, 0, fmt.Errorf("CouponRepo|GetCacheCouponNum:%v", err)
	}
	return dbNum, cacheNum, nil
}

// GetNextUsefulCouponFromCache 从缓存中拿出一个可用优惠券
func (r *CouponRepo) GetNextUsefulCouponFromCache(prizeID int) (string, error) {
	key := fmt.Sprintf(constant.PrizeCouponCacheKey+"%d", prizeID)
	code, err := cache.GetRedisCli().SPop(context.Background(), key)
	if err != nil {
		if err.Error() == "redis: nil" {
			log.Infof("coupon not left")
			return "", nil
		}
		return "", fmt.Errorf("lotteryService|PrizeCouponDiffByCache:%v", err)
	}
	if code == "" {
		log.Infof("lotteryService|PrizeCouponDiffByCache code is nil with prize_id=%d", prizeID)
		return "", nil
	}
	return code, nil
}

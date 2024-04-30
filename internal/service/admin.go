package service

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"lottery_single/internal/model"
	"lottery_single/internal/pkg/constant"
	"lottery_single/internal/pkg/middlewares/cache"
	"lottery_single/internal/pkg/middlewares/gormcli"
	"lottery_single/internal/pkg/middlewares/log"
	"lottery_single/internal/pkg/utils"
	"lottery_single/internal/repo"
	"strconv"
	"strings"
	"time"
)

// AdminService 后将系统后台管理功能
type AdminService interface {
	// 用户操作
	AddUser(ctx context.Context, user *model.User) error

	// 奖品操作
	AddPrize(ctx context.Context, viewPrize *ViewPrize) error
	AddPrizeWithCache(ctx context.Context, viewPrize *ViewPrize) error
	AddPrizeWithPool(ctx context.Context, viewPrize *ViewPrize) error
	GetPrizeList(ctx context.Context) ([]*model.Prize, error)
	GetPrizeListWithCache(ctx context.Context) ([]*model.Prize, error)
	GetViewPrizeList(ctx context.Context) ([]*ViewPrize, error)
	GetViewPrizeListWithCache(ctx context.Context) ([]*ViewPrize, error)
	GetPrize(ctx context.Context, id uint) (*ViewPrize, error)
	UpdatePrize(ctx context.Context, viewPrize *ViewPrize) error
	UpdateDbPrizeWithCache(ctx context.Context, db *gorm.DB, prize *model.Prize, cols ...string) error
	UpdateDbPrize(ctx context.Context, db *gorm.DB, prize *model.Prize, cols ...string) error
	ResetPrizePlan(ctx context.Context, prize *model.Prize) error

	// 优惠券操作
	GetCouponList(ctx context.Context, prizeID uint) ([]*ViewCouponInfo, int64, int64, error)
	ImportCoupon(ctx context.Context, prizeID uint, codes string) (int, int, error)
	ImportCouponWithCache(ctx context.Context, prizeID uint, codes string) (int, int, error)
}

type adminService struct {
	couponRepo *repo.CouponRepo
	prizeRepo  *repo.PrizeReop
	userRepo   *repo.UserRepo
}

var adminServiceImpl *adminService

func InitAdminService() {
	adminServiceImpl = &adminService{
		couponRepo: repo.NewCouponRepo(),
		prizeRepo:  repo.NewPrizeRepo(),
		userRepo:   repo.NewUserRepo(),
	}
}

func GetAdminService() AdminService {
	return adminServiceImpl
}

func (a *adminService) AddUser(ctx context.Context, user *model.User) error {
	return a.userRepo.Create(gormcli.GetDB(), user)
}

// GetPrizeList 获取db奖品列表
func (a *adminService) GetPrizeList(ctx context.Context) ([]*model.Prize, error) {
	log.InfoContextf(ctx, "GetPrizeList!!!!!")
	db := gormcli.GetDB()
	list, err := a.prizeRepo.GetAll(db)
	if err != nil {
		log.ErrorContextf(ctx, "prizeService|GetPrizeList err:%v", err)
		return nil, fmt.Errorf("prizeService|GetPrizeList: %v", err)
	}
	return list, nil
}

// GetPrizeListWithCache 获取db奖品列表
func (a *adminService) GetPrizeListWithCache(ctx context.Context) ([]*model.Prize, error) {
	log.InfoContextf(ctx, "GetPrizeListWithCache!!!!!")
	db := gormcli.GetDB()
	list, err := a.prizeRepo.GetAllWithCache(db)
	if err != nil {
		log.ErrorContextf(ctx, "prizeService|GetPrizeList err:%v", err)
		return nil, fmt.Errorf("prizeService|GetPrizeList: %v", err)
	}
	return list, nil
}

// GetViewPrizeList 获取奖品列表,这个方法用于管理后台使用，因为管理后台不需要高性能，所以不走缓存
func (a *adminService) GetViewPrizeList(ctx context.Context) ([]*ViewPrize, error) {
	log.InfoContextf(ctx, "GetPrizeList!!!!!")
	db := gormcli.GetDB()
	list, err := a.prizeRepo.GetAll(db)
	if err != nil {
		log.ErrorContextf(ctx, "prizeService|GetPrizeList err:%v", err)
		return nil, fmt.Errorf("prizeService|GetPrizeList: %v", err)
	}
	prizeList := make([]*ViewPrize, 0)
	for _, prize := range list {
		if prize.SysStatus != constant.PrizeStatusNormal {
			continue
		}
		num, err := a.prizeRepo.GetPrizePoolNum(prize.Id)
		if err != nil {
			return nil, fmt.Errorf("prizeService|GetPrizeList: %v", err)
		}
		title := fmt.Sprintf("【%d】%s", num, prize.Title)
		prizeList = append(prizeList, &ViewPrize{
			Id:        prize.Id,
			Title:     title,
			Img:       prize.Img,
			PrizeNum:  prize.PrizeNum,
			LeftNum:   prize.LeftNum,
			PrizeType: prize.PrizeType,
		})

	}
	return prizeList, nil
}

// GetViewPrizeListWithCache 获取奖品列表,优先从缓存获取
func (a *adminService) GetViewPrizeListWithCache(ctx context.Context) ([]*ViewPrize, error) {
	log.InfoContextf(ctx, "GetViewPrizeListWithCache!!!!!")
	db := gormcli.GetDB()
	list, err := a.prizeRepo.GetAllWithCache(db)
	if err != nil {
		log.ErrorContextf(ctx, "prizeService|GetPrizeList err:%v", err)
		return nil, fmt.Errorf("prizeService|GetPrizeList: %v", err)
	}
	prizeList := make([]*ViewPrize, 0)
	for _, prize := range list {
		if prize.SysStatus != constant.PrizeStatusNormal {
			continue
		}
		prizeList = append(prizeList, &ViewPrize{
			Id:        prize.Id,
			Title:     prize.Title,
			Img:       prize.Img,
			PrizeNum:  prize.PrizeNum,
			LeftNum:   prize.LeftNum,
			PrizeType: prize.PrizeType,
		})
	}
	return prizeList, nil
}

// GetPrize 获取某个奖品
func (a *adminService) GetPrize(ctx context.Context, id uint) (*ViewPrize, error) {
	prizeModel, err := a.prizeRepo.Get(gormcli.GetDB(), id)
	if err != nil {
		log.ErrorContextf(ctx, "prizeService|GetPrize:%v", err)
		return nil, fmt.Errorf("prizeService|GetPrize:%v", err)
	}
	prize := &ViewPrize{
		Id:        prizeModel.Id,
		Title:     prizeModel.Title,
		Img:       prizeModel.Img,
		PrizeNum:  prizeModel.PrizeNum,
		LeftNum:   prizeModel.LeftNum,
		PrizeType: prizeModel.PrizeType,
	}
	return prize, nil
}

// AddPrize 新增奖品
func (a *adminService) AddPrize(ctx context.Context, viewPrize *ViewPrize) error {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("AddPrize panic%v\n", err)
		}
	}()
	prize := model.Prize{
		Title:        viewPrize.Title,
		PrizeNum:     viewPrize.PrizeNum,
		LeftNum:      viewPrize.PrizeNum,
		PrizeCode:    viewPrize.PrizeCode,
		PrizeTime:    viewPrize.PrizeTime,
		Img:          viewPrize.Img,
		DisplayOrder: viewPrize.DisplayOrder,
		PrizeType:    viewPrize.PrizeType,
		BeginTime:    viewPrize.BeginTime,
		EndTime:      viewPrize.EndTime,
		PrizePlan:    viewPrize.PrizePlan,
		SysStatus:    1,
	}
	// 因为奖品是全量string缓存，新增奖品之后缓存有变动，所有要更新
	if err := a.prizeRepo.Create(gormcli.GetDB(), &prize); err != nil {
		log.Errorf("adminService|AddPrize err:%v", err)
		return fmt.Errorf("adminService|AddPrize:%v", err)
	}
	return nil
}

// AddPrizeWithPool 带奖品池的新增奖品实现
func (a *adminService) AddPrizeWithPool(ctx context.Context, viewPrize *ViewPrize) error {
	prize := model.Prize{
		Title:        viewPrize.Title,
		PrizeNum:     viewPrize.PrizeNum,
		LeftNum:      viewPrize.PrizeNum,
		PrizeCode:    viewPrize.PrizeCode,
		PrizeTime:    viewPrize.PrizeTime,
		Img:          viewPrize.Img,
		DisplayOrder: viewPrize.DisplayOrder,
		PrizeType:    viewPrize.PrizeType,
		BeginTime:    viewPrize.BeginTime,
		EndTime:      viewPrize.EndTime,
		PrizePlan:    viewPrize.PrizePlan,
		SysStatus:    1,
		//SysUpdated:   time.Now(),
	}
	// 因为奖品是全量string缓存，新增奖品之后缓存有变动，所有要更新
	if err := a.prizeRepo.CreateWithCache(gormcli.GetDB(), &prize); err != nil {
		log.Errorf("adminService|AddPrize err:%v", err)
		return fmt.Errorf("adminService|AddPrize:%v", err)
	}
	if err := a.ResetPrizePlan(ctx, &prize); err != nil {
		log.Errorf("adminService|AddPrize ResetPrizePlan prize err:%v", err)
		return fmt.Errorf("adminService|AddPrize ResetPrizePlan prize err:%v", err)
	}
	return nil
}

// AddPrizeWithCache 带缓存优化的新增奖品
func (a *adminService) AddPrizeWithCache(ctx context.Context, viewPrize *ViewPrize) error {
	prize := model.Prize{
		Title:        viewPrize.Title,
		PrizeNum:     viewPrize.PrizeNum,
		LeftNum:      viewPrize.PrizeNum,
		PrizeCode:    viewPrize.PrizeCode,
		PrizeTime:    viewPrize.PrizeTime,
		Img:          viewPrize.Img,
		DisplayOrder: viewPrize.DisplayOrder,
		PrizeType:    viewPrize.PrizeType,
		BeginTime:    viewPrize.BeginTime,
		EndTime:      viewPrize.EndTime,
		PrizePlan:    viewPrize.PrizePlan,
		SysStatus:    1,
		//SysUpdated:   time.Now(),
	}
	// 因为奖品是全量string缓存，新增奖品之后缓存有变动，所有要更新
	if err := a.prizeRepo.CreateWithCache(gormcli.GetDB(), &prize); err != nil {
		log.Errorf("adminService|AddPrize err:%v", err)
		return fmt.Errorf("adminService|AddPrize:%v", err)
	}
	return nil
}

func (a *adminService) UpdateDbPrizeWithCache(ctx context.Context, db *gorm.DB, prize *model.Prize, cols ...string) error {
	return a.prizeRepo.UpdateWithCache(db, prize, cols...)
}

func (a *adminService) UpdateDbPrize(ctx context.Context, db *gorm.DB, prize *model.Prize, cols ...string) error {
	return a.prizeRepo.Update(db, prize, cols...)
}

func (a *adminService) UpdatePrize(ctx context.Context, viewPrize *ViewPrize) error {
	if viewPrize == nil || viewPrize.Id <= 0 {
		log.Errorf("adminService|UpdatePrize invalid prize err:%v", viewPrize)
		return fmt.Errorf("adminService|UpdatePrize invalid prize")
	}
	prize := model.Prize{
		Title:        viewPrize.Title,
		PrizeNum:     viewPrize.PrizeNum,
		LeftNum:      viewPrize.LeftNum,
		PrizeCode:    viewPrize.PrizeCode,
		PrizeTime:    viewPrize.PrizeTime,
		Img:          viewPrize.Img,
		DisplayOrder: viewPrize.DisplayOrder,
		PrizeType:    viewPrize.PrizeType,
		BeginTime:    viewPrize.BeginTime,
		EndTime:      viewPrize.EndTime,
		PrizePlan:    viewPrize.PrizePlan,
		SysStatus:    viewPrize.SysStatus,
	}
	oldPrize, err := a.prizeRepo.Get(gormcli.GetDB(), viewPrize.Id)
	if err != nil {
		log.Errorf("adminService|UpdatePrize get old prize err:%v", err)
		return fmt.Errorf("adminService|UpdatePrize:%v", err)
	}
	if oldPrize == nil {
		log.Errorf("adminService|UpdatePrize prize not exists with id: %d", viewPrize.Id)
		return fmt.Errorf("adminService|UpdatePrize prize not exists with id: %d", viewPrize.Id)
	}
	// 奖品数量发生了改变
	if prize.PrizeNum != oldPrize.PrizeNum {
		if prize.PrizeNum <= 0 {
			prize.PrizeNum = 0
		}
		if prize.LeftNum <= 0 {
			prize.LeftNum = 0
		}
	}
	if a.prizeRepo.Update(gormcli.GetDB(), &prize, "title", "prize_num", "left_num", "prize_code", "prize_time", "img",
		"display_order", "prize_type", "begin_time", "end_time", "prize_plan"); err != nil {
		log.Errorf("adminService|UpdatePrize Update prize err:%v", err)
		return fmt.Errorf("adminService|UpdatePrize Update prize:%v", err)
	}
	return nil
}

func (a *adminService) UpdatePrizeWithPool(ctx context.Context, viewPrize *ViewPrize) error {
	if viewPrize == nil || viewPrize.Id <= 0 {
		log.Errorf("adminService|UpdatePrize invalid prize err:%v", viewPrize)
		return fmt.Errorf("adminService|UpdatePrize invalid prize")
	}
	prize := model.Prize{
		Title:        viewPrize.Title,
		PrizeNum:     viewPrize.PrizeNum,
		LeftNum:      viewPrize.LeftNum,
		PrizeCode:    viewPrize.PrizeCode,
		PrizeTime:    viewPrize.PrizeTime,
		Img:          viewPrize.Img,
		DisplayOrder: viewPrize.DisplayOrder,
		PrizeType:    viewPrize.PrizeType,
		BeginTime:    viewPrize.BeginTime,
		EndTime:      viewPrize.EndTime,
		PrizePlan:    viewPrize.PrizePlan,
		SysStatus:    viewPrize.SysStatus,
	}
	oldPrize, err := a.prizeRepo.Get(gormcli.GetDB(), viewPrize.Id)
	if err != nil {
		log.Errorf("adminService|UpdatePrize get old prize err:%v", err)
		return fmt.Errorf("adminService|UpdatePrize:%v", err)
	}
	if oldPrize == nil {
		log.Errorf("adminService|UpdatePrize prize not exists with id: %d", viewPrize.Id)
		return fmt.Errorf("adminService|UpdatePrize prize not exists with id: %d", viewPrize.Id)
	}
	// 奖品数量发生了改变
	if prize.PrizeNum != oldPrize.PrizeNum {
		if prize.PrizeNum <= 0 {
			prize.PrizeNum = 0
		}
		if prize.LeftNum <= 0 {
			prize.LeftNum = 0
		}
		if err := a.ResetPrizePlan(ctx, &prize); err != nil {
			log.Errorf("adminService|UpdatePrize ResetPrizePlan prize err:%v", err)
			return fmt.Errorf("adminService|UpdatePrize ResetPrizePlan prize err:%v", err)
		}
	}
	if prize.PrizeTime != oldPrize.PrizeTime {
		if err := a.ResetPrizePlan(ctx, &prize); err != nil {
			log.Errorf("adminService|UpdatePrize ResetPrizePlan prize err:%v", err)
			return fmt.Errorf("adminService|UpdatePrize ResetPrizePlan prize err:%v", err)
		}
	}
	if a.prizeRepo.Update(gormcli.GetDB(), &prize, "title", "prize_num", "left_num", "prize_code", "prize_time", "img",
		"display_order", "prize_type", "begin_time", "end_time", "prize_plan"); err != nil {
		log.Errorf("adminService|UpdatePrize Update prize err:%v", err)
		return fmt.Errorf("adminService|UpdatePrize Update prize:%v", err)
	}
	return nil
}

// GetCouponList 获取优惠券列表,库存优惠券数量和缓存优惠券数量，当这两个数量不一致的时候，需要重置缓存优惠券数量
func (a *adminService) GetCouponList(ctx context.Context, prizeID uint) ([]*ViewCouponInfo, int64, int64, error) {
	var (
		viewCouponList []*ViewCouponInfo
		couponList     []*model.Coupon
		err            error
		dbNum          int64
		cacheNum       int64
	)
	if prizeID > 0 {
		couponList, err = a.couponRepo.GetCouponListByPrizeID(gormcli.GetDB(), prizeID)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("adminService|GetCouponList invalid prize_id:%d", prizeID)
		}
		dbNum, cacheNum, err = a.couponRepo.GetCacheCouponNum(gormcli.GetDB(), prizeID)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("adminService|GetCouponList invalid prize_id:%d", prizeID)
		}
	} else {
		couponList, err = a.couponRepo.GetAll(gormcli.GetDB())
		if err != nil {
			return nil, 0, 0, fmt.Errorf("adminService|GetCouponList invalid prize_id:%d", prizeID)
		}
	}
	for _, coupon := range couponList {
		viewCouponList = append(viewCouponList, &ViewCouponInfo{
			Id:      coupon.Id,
			PrizeId: coupon.PrizeId,
			Code:    coupon.Code,
			//SysCreated: coupon.SysCreated,
			//SysUpdated: coupon.SysUpdated,
			SysStatus: coupon.SysStatus,
		})
	}
	return viewCouponList, dbNum, cacheNum, nil
}

// ImportCoupon 导入优惠券
func (a *adminService) ImportCoupon(ctx context.Context, prizeID uint, codes string) (int, int, error) {
	if prizeID <= 0 {
		return 0, 0, fmt.Errorf("adminService|ImportCoupon invalid prizeID:%d", prizeID)
	}
	prize, err := a.prizeRepo.Get(gormcli.GetDB(), prizeID)
	if err != nil {
		return 0, 0, fmt.Errorf("adminService|ImportCoupon invalid prizeID:%d", prizeID)
	}
	if prize == nil || prize.PrizeType != constant.PrizeTypeCouponDiff {
		log.InfoContext(ctx, "adminService|ImportCoupon invalid prize type:%d with prize_id %d", prize.PrizeType, prizeID)
		return 0, 0, fmt.Errorf("adminService|ImportCoupon prize_type is not coupon with prize_id %d", prizeID)
	}
	var (
		successNum int
		failNum    int
	)
	codeList := strings.Split(codes, "\n")
	for _, code := range codeList {
		code = strings.TrimSpace(code)
		coupon := &model.Coupon{
			PrizeId: prizeID,
			Code:    code,
			//SysCreated: time.Now(),
			SysStatus: 1,
		}
		if err = a.couponRepo.Create(gormcli.GetDB(), coupon); err != nil {
			failNum++
		} else {
			successNum++
		}
	}
	return successNum, failNum, nil
}

// ImportCouponWithCache 导入优惠券
func (a *adminService) ImportCouponWithCache(ctx context.Context, prizeID uint, codes string) (int, int, error) {
	if prizeID <= 0 {
		return 0, 0, fmt.Errorf("adminService|ImportCoupon invalid prizeID:%d", prizeID)
	}
	prize, err := a.prizeRepo.GetWithCache(gormcli.GetDB(), prizeID)
	if err != nil {
		return 0, 0, fmt.Errorf("adminService|ImportCoupon invalid prizeID:%d", prizeID)
	}
	if prize == nil || prize.PrizeType != constant.PrizeTypeCouponDiff {
		log.InfoContext(ctx, "adminService|ImportCoupon invalid prize type:%d with prize_id %d", prize.PrizeType, prizeID)
		return 0, 0, fmt.Errorf("adminService|ImportCoupon prize_type is not coupon with prize_id %d", prizeID)
	}
	var (
		successNum int
		failNum    int
	)
	codeList := strings.Split(codes, "\n")
	for _, code := range codeList {
		code = strings.TrimSpace(code)
		coupon := &model.Coupon{
			PrizeId: prizeID,
			Code:    code,
			//SysCreated: time.Now(),
			SysStatus: 1,
		}
		if err = a.couponRepo.Create(gormcli.GetDB(), coupon); err != nil {
			failNum++
		} else {
			// db导入成功之后，再导入缓存
			ok, err := a.couponRepo.ImportCacheCoupon(prizeID, code)
			if err != nil {
				return 0, 0, fmt.Errorf("adminService|ImportCoupon prize_type is not coupon with prize_id %d", prizeID)
			}
			if !ok {
				failNum++
			} else {
				successNum++
			}
		}
	}
	return successNum, failNum, nil
}

// ReCacheCoupon 根据数据库重置某种奖品的优惠券数据到缓存中
func (a *adminService) ReCacheCoupon(ctx *context.Context, prizeID uint) (int64, int64, error) {
	if prizeID <= 0 {
		return 0, 0, fmt.Errorf("adminService|ReCacheCoupon invalid prizeID:%d", prizeID)
	}
	successNum, failureNum, err := a.couponRepo.ReSetCacheCoupon(gormcli.GetDB(), prizeID)
	if err != nil {
		return 0, 0, fmt.Errorf("adminService|ReCacheCoupon:%v", err)
	}
	return successNum, failureNum, nil
}

// ResetPrizePlan 重置某种奖品的发奖计划
func (a *adminService) ResetPrizePlan(ctx context.Context, prize *model.Prize) error {
	if prize == nil || prize.Id < 1 {
		return fmt.Errorf("limitService|ResetGiftPrizePlan invalid prize")
	}
	now := time.Now()
	// 奖品状态不对，不能发奖
	if prize.SysStatus == 2 ||
		prize.BeginTime.After(now) || // 还未开始
		prize.EndTime.Before(now) || // 已经结束
		prize.LeftNum <= 0 ||
		prize.PrizeNum <= 0 {
		if prize.PrizePlan != "" {
			// 在重置的时候，如果发现原来奖品的发奖计划不为空，需要清空发奖计划
			a.clearPrizePlan(ctx, prize)
		}
		log.InfoContext(ctx, "prize can not be given out")
		return nil
	}
	// PrizeTime, 发奖周期，这类奖品需要在多少天内发完
	prizePlanDays := int(prize.PrizeTime)
	if prizePlanDays <= 0 {
		a.setPrizePool(ctx, prize.Id, prize.LeftNum)
		log.InfoContext(ctx, "adminService|ResetGiftPrizePlan|prizePlanDays <= 0")
		return nil
	}
	// 对于设置发奖周期的奖品重新计算出来合适的奖品发放节奏
	// 奖品池的剩余数先设置为空
	a.setPrizePool(ctx, prize.Id, 0)
	// 发奖周期中的每天的发奖概率一样，一天内24小时，每个小时的概率是不一样的，每个小时内的每一分钟的概率一样
	prizeNum := prize.PrizeNum
	// 先计算每天至少发多少奖
	avgPrizeNum := prizeNum / prizePlanDays

	// 每天可以分配到的奖品数量
	dayPrizeNumMap := make(map[int]int)
	// 发奖周期大雨1天，并且平均每天发的奖品书大于等于1
	if prizePlanDays > 0 && avgPrizeNum >= 1 {
		for day := 0; day < prizePlanDays; day++ {
			dayPrizeNumMap[day] = avgPrizeNum
		}
	}
	// 剩下的奖品一个一个的随机分配到任意哪天
	prizeNum -= prizePlanDays * avgPrizeNum
	for prizeNum > 0 {
		prizeNum--
		day := utils.Random(prizePlanDays)
		dayPrizeNumMap[day] += 1
	}
	// 发奖map：map[int]map[int][60]int
	//map[天]map[小时][60]奖品数量：后一个map表示value是一个60大小的数组，表示一个小时中每分钟要发的奖品数量
	prizePlanMap := make(map[int]map[int][60]int)
	log.Infof("prize_id = %d\ndayPrizeNumMap = %+v", prize.Id, dayPrizeNumMap)
	for day, num := range dayPrizeNumMap {
		//计算一天的发奖计划
		dayPrizePlan := a.prizePlanOneDay(num)
		prizePlanMap[day] = dayPrizePlan
	}
	log.Infof("prize_id = %d\nprizePlanMap = %+v", prize.Id, prizePlanMap)
	// 格式化 dayPrizePlan数据，序列化成为一个[时间:数量]二元组的数组
	planList, err := a.formatPrizePlan(now, prizePlanDays, prizePlanMap)
	if err != nil {
		log.ErrorContextf(ctx, "limitService|ResetPrizePlan|formatPrizePlan err:", err)
		return fmt.Errorf("limitService|ResetGiftPrizePlan:%v", err)
	}
	bytes, err := json.Marshal(planList)
	if err != nil {
		log.ErrorContextf(ctx, "limitService|ResetPrizePlan|planList json marshal error=", err)
		return fmt.Errorf("limitService|ResetGiftPrizePlan:%v", err)
	}
	// 保存奖品的分布计划数据
	info := &model.Prize{
		Id:         prize.Id,
		LeftNum:    prize.PrizeNum,
		PrizePlan:  string(bytes),
		PrizeBegin: now,
		PrizeEnd:   now.Add(time.Second * time.Duration(86400*prizePlanDays)),
	}
	err = a.prizeRepo.UpdateWithCache(gormcli.GetDB(), info, "prize_plan", "prize_begin", "prize_end")
	if err != nil {
		log.ErrorContextf(ctx, "limitService|ResetPrizePlan|prizeRepo.Update err:", err)
		return fmt.Errorf("limitService|ResetPrizePlan:%v", err)
	}
	return nil
}

// clearPrizeData 清空奖品的发放计划
func (a *adminService) clearPrizePlan(ctx context.Context, prize *model.Prize) error {
	info := &model.Prize{
		Id:        prize.Id,
		PrizePlan: "",
	}
	err := a.prizeRepo.UpdateWithCache(gormcli.GetDB(), info, "prize_plan")
	if err != nil {
		log.ErrorContextf(ctx, "limitService|clearPrizePlan|prizeRepo.Update err", err)
		return fmt.Errorf("limitService|clearPrizePlan:%v", err)
	}
	//奖品池也设为0
	if err = a.setPrizePool(ctx, prize.Id, 0); err != nil {
		return fmt.Errorf("limitService|clearPrizePlan:%v", err)
	}
	return nil
}

// setGiftPool 设置奖品池中某种奖品的数量
func (a *adminService) setPrizePool(ctx context.Context, id uint, num int) error {
	key := constant.PrizePoolCacheKey
	idStr := strconv.Itoa(int(id))
	_, err := cache.GetRedisCli().HSet(ctx, key, idStr, strconv.Itoa(num))
	if err != nil {
		return fmt.Errorf("adminService|setPrizePool:%v", err)
	}
	return nil
}

// prizePlanOneDay 计算一天内具体到每小时每分钟应该发出的奖品，map[int][60]int： map[hour][minute]num
func (a *adminService) prizePlanOneDay(num int) map[int][60]int {
	resultMap := make(map[int][60]int)
	hourPrizeNumList := [24]int{} // 长度为24的数组表示1天中每个小时对应的奖品数
	// 计算一天中的24个小时，每个小时应该发出的奖品数，为什么是100，100表示每一天的权重百分比
	if num > 100 {
		for _, h := range DayPrizeWeights {
			hourPrizeNumList[h]++
		}
		for h := 0; h < 24; h++ {
			d := hourPrizeNumList[h]
			n := num * d / 100 // d / 100 每个小时所占的奖品数量概率
			hourPrizeNumList[h] = n
			num -= n
		}
	}
	log.Infof("num = %d", num)
	for num > 0 {
		num--
		// 随机将这个奖品分配到某一个小时上
		hourIndex := utils.Random(100)
		log.Infof("hourIndex = %d", hourIndex)
		h := DayPrizeWeights[hourIndex]
		hourPrizeNumList[h]++
	}
	log.Infof("hourPrizeNumList = %v", hourPrizeNumList)
	// 将每个小时内的奖品数量分配到60分钟
	for h, hourPrizenum := range hourPrizeNumList {
		if hourPrizenum <= 0 {
			continue
		}
		minutePrizeNumList := [60]int{}
		if hourPrizenum >= 60 {
			avgMinutePrizeNum := hourPrizenum / 60
			for i := 0; i < 60; i++ {
				minutePrizeNumList[i] = avgMinutePrizeNum
			}
			hourPrizenum -= avgMinutePrizeNum * 60
		}
		for hourPrizenum > 0 {
			hourPrizenum--
			// 随机将这个奖品分配到某一分钟上
			m := utils.Random(60)
			log.Infof("minuteIndex = %d", m)
			minutePrizeNumList[m]++
		}
		log.Infof("minutePrizeNumList = %v", minutePrizeNumList)
		resultMap[h] = minutePrizeNumList
	}
	log.Infof("resultMap=%v", resultMap)
	log.Infof("-----------------------------------------------------------")
	return resultMap
}

// 将prizeData格式化成具体到一个时间（分钟）的奖品数量
// 结构为： [day][hour][minute]num
// result: [][时间, 数量]
func (a *adminService) formatPrizePlan(now time.Time, prizePlanDays int, prizePlan map[int]map[int][60]int) ([]*TimePrizeInfo, error) {
	result := make([]*TimePrizeInfo, 0)
	nowHour := now.Hour()
	for i := 0; i < prizePlanDays; i++ {
		dayPrizePlanMap, ok := prizePlan[i]
		if !ok {
			continue
		}
		dayTimeStamp := int(now.Unix()) + i*86400 // dayTimeStamp 为发奖周期中的每一天对应当前时间的时刻
		for h := 0; h < 24; h++ {
			hourPrizePlanMap, ok := dayPrizePlanMap[(h+nowHour)%24]
			if !ok {
				continue
			}
			hourTimeStamp := dayTimeStamp + h*3600 // hourTimeStamp 为发奖周期中的每一天中每个小时对应的时刻
			for m := 0; m < 60; m++ {
				num := hourPrizePlanMap[m]
				if num <= 0 {
					continue
				}
				// 找到特定一个时间的计划数据
				minuteTimeStamp := hourTimeStamp + m*60 // minuteTimeStamp 为发奖周期中的每一分钟对应的时刻
				result = append(result, &TimePrizeInfo{
					Time: utils.FormatFromUnixTime(int64(minuteTimeStamp)),
					Num:  num,
				})
			}
		}
	}
	return result, nil
}

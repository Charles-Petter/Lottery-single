package service

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"lottery_single/internal/model"
	"lottery_single/internal/pkg/constant"
	"lottery_single/internal/pkg/middlewares/gormcli"
	"lottery_single/internal/pkg/middlewares/lock"
	"lottery_single/internal/pkg/middlewares/log"
	"lottery_single/internal/repo"
	"strconv"
	"strings"
	"time"
)

// LotteryService 抽发奖功能
type LotteryService interface {
	GetPrize(ctx context.Context, prizeCode int) (*LotteryPrize, error)
	GetPrizeWithCache(ctx context.Context, prizeCode int) (*LotteryPrize, error)
	GetAllUsefulPrizes(ctx context.Context) ([]*LotteryPrize, error)
	GetAllUsefulPrizesWithCache(ctx context.Context) ([]*LotteryPrize, error)
	PrizeCouponDiff(ctx context.Context, prizeID int) (string, error)
	PrizeCouponDiffWithCache(ctx context.Context, prizeID int) (string, error)
	PrizeLargeBlackLimit(ctx context.Context, blackUser *model.BlackUser, blackIp *model.BlackIp, info *LotteryUserInfo) error
	GiveOutPrize(ctx context.Context, prizeID int) (bool, error)
	GiveOutPrizeWithCache(ctx context.Context, prizeID int) (bool, error)
	GiveOutPrizeWithPool(ctx context.Context, prizeID int) (bool, error)
	GetPrizeNumWithPool(ctx context.Context, prizeID uint) (int, error)
}

type lotteryService struct {
	prizeReop     *repo.PrizeReop
	couponReop    *repo.CouponRepo
	blackUserRepo *repo.BlackUserRepo
	blackIpRepo   *repo.BlackIpRepo
}

var lotteryServiceImpl *lotteryService

func NewLotteryService() {
	lotteryServiceImpl = &lotteryService{
		prizeReop:     repo.NewPrizeRepo(),
		couponReop:    repo.NewCouponRepo(),
		blackUserRepo: repo.NewBlackUserRepo(),
		blackIpRepo:   repo.NewBlackIpRepo(),
	}
}

func GetLotteryService() LotteryService {
	return lotteryServiceImpl
}

func (l *lotteryService) Lottery(ctx context.Context, db *gorm.DB) {

}

func (l *lotteryService) GetPrize(ctx context.Context, prizeCode int) (*LotteryPrize, error) {
	var prize *LotteryPrize
	lotteryPrizeList, err := l.GetAllUsefulPrizes(ctx)
	if err != nil {
		log.ErrorContextf(ctx, "lotteryService|ToLotteryPrize:%v", err)
		return nil, err
	}
	for _, lotteryPrize := range lotteryPrizeList {
		if lotteryPrize.PrizeCodeLow <= prizeCode &&
			lotteryPrize.PrizeCodeHigh >= prizeCode {
			// 中奖编码区间满足条件，说明可以中奖
			if lotteryPrize.PrizeType < constant.PrizeTypeEntitySmall { //如果非实物奖直接发，实物奖需要看是不是在黑名单外
				prize = lotteryPrize
				break
			}
		}
	}
	return prize, nil
}

// GetPrizeWithCache 获取中奖的奖品类型
func (l *lotteryService) GetPrizeWithCache(ctx context.Context, prizeCode int) (*LotteryPrize, error) {
	var prize *LotteryPrize
	lotteryPrizeList, err := l.GetAllUsefulPrizesWithCache(ctx)
	if err != nil {
		log.ErrorContextf(ctx, "lotteryService|ToLotteryPrize:%v", err)
		return nil, err
	}
	for _, lotteryPrize := range lotteryPrizeList {
		if lotteryPrize.PrizeCodeLow <= prizeCode &&
			lotteryPrize.PrizeCodeHigh >= prizeCode {
			// 中奖编码区间满足条件，说明可以中奖
			//if lotteryPrize.PrizeType < constant.PrizeTypeEntitySmall { //如果非实物奖直接发，实物奖需要看是不是在黑名单外
			prize = lotteryPrize
			break
			//}
		}
	}
	return prize, nil
}

// GiveOutPrize 发奖，奖品数量减1
func (l *lotteryService) GiveOutPrize(ctx context.Context, prizeID int) (bool, error) {
	// 该类奖品的库存数量减1
	ok, err := l.prizeReop.DecrLeftNum(gormcli.GetDB(), prizeID, 1)
	if err != nil {
		log.ErrorContextf(ctx, "lotteryService|GiveOutPrize err:%v", err)
		return false, fmt.Errorf("lotteryService|GiveOutPrize:%v", err)
	}
	if !ok {
		return false, nil
	}
	return true, nil
}

// GiveOutPrizeWithCache 发奖，奖品数量减1,并且同步更新缓存
func (l *lotteryService) GiveOutPrizeWithCache(ctx context.Context, prizeID int) (bool, error) {
	// 该类奖品的库存数量减1
	ok, err := l.prizeReop.DecrLeftNum(gormcli.GetDB(), prizeID, 1)
	if err != nil {
		log.ErrorContextf(ctx, "lotteryService|GiveOutPrize err:%v", err)
		return false, fmt.Errorf("lotteryService|GiveOutPrize:%v", err)
	}
	if !ok {
		return false, nil
	}
	// 扣减库存成功
	if err = l.prizeReop.UpdateByCache(&model.Prize{Id: uint(prizeID)}); err != nil {
		log.ErrorContextf(ctx, "lotteryService|GiveOutPrize|UpdateByCache err:%v", err)
		return false, fmt.Errorf("lotteryService|GiveOutPrize|UpdateByCache:%v", err)
	}
	return true, nil
}

func (l *lotteryService) GiveOutPrizeWithPool(ctx context.Context, prizeID int) (bool, error) {
	cnt, err := l.prizeReop.DecrLeftNumByPool(prizeID)
	if err != nil {
		log.ErrorContextf(ctx, "lotteryService|GiveOutPrizeWithPool err:%v", err)
	}
	// 扣减完之后剩余奖品池中该奖品的数量小与0，所以当前时段该奖品不足了，不能发奖
	if cnt < 0 {
		return false, nil
	}
	// 奖品池成功之后再周数据库发奖逻辑
	return l.GiveOutPrize(ctx, prizeID)
}

// GetAllUsefulPrizes 获取所有可用奖品
func (l *lotteryService) GetAllUsefulPrizes(ctx context.Context) ([]*LotteryPrize, error) {
	list, err := l.prizeReop.GetAllUsefulPrizeList(gormcli.GetDB())
	if err != nil {
		log.ErrorContextf(ctx, "lotteryService|GetAllUsefulPrizes:%v", err)
		return nil, fmt.Errorf("lotteryService|GetAllUsefulPrizes:%v", err)
	}
	if len(list) == 0 {
		return nil, nil
	}
	lotteryPrizeList := make([]*LotteryPrize, 0)
	for _, prize := range list {
		codes := strings.Split(prize.PrizeCode, "-")
		if len(codes) == 2 {
			// 设置了获奖编码范围 a-b 才可以进行抽奖
			codeA := codes[0]
			codeB := codes[1]
			low, err1 := strconv.Atoi(codeA)
			high, err2 := strconv.Atoi(codeB)
			if err1 == nil && err2 == nil && high >= low && low >= 0 && high < constant.PrizeCodeMax {
				lotteryPrize := &LotteryPrize{
					Id:            prize.Id,
					Title:         prize.Title,
					PrizeNum:      prize.PrizeNum,
					LeftNum:       prize.LeftNum,
					PrizeCodeLow:  low,
					PrizeCodeHigh: high,
					Img:           prize.Img,
					DisplayOrder:  prize.DisplayOrder,
					PrizeType:     prize.PrizeType,
					PrizeProfile:  prize.PrizeProfile,
				}
				lotteryPrizeList = append(lotteryPrizeList, lotteryPrize)
			}
		}
	}
	return lotteryPrizeList, nil
}

func (l *lotteryService) GetAllUsefulPrizesWithCache(ctx context.Context) ([]*LotteryPrize, error) {
	// 筛选出符合条件的奖品列表
	list, err := l.prizeReop.GetAllUsefulPrizeListWithCache(gormcli.GetDB())
	if err != nil {
		log.ErrorContextf(ctx, "lotteryService|GetAllUsefulPrizes:%v", err)
		return nil, fmt.Errorf("lotteryService|GetAllUsefulPrizes:%v", err)
	}
	if len(list) == 0 {
		return nil, nil
	}
	// 对db的prize做一个类型转换，转化为LotteryPrize
	lotteryPrizeList := make([]*LotteryPrize, 0)
	for _, prize := range list {
		codes := strings.Split(prize.PrizeCode, "-")
		if len(codes) == 2 {
			// 设置了获奖编码范围 a-b 才可以进行抽奖
			codeA := codes[0]
			codeB := codes[1]
			low, err1 := strconv.Atoi(codeA)
			high, err2 := strconv.Atoi(codeB)
			if err1 == nil && err2 == nil && high >= low && low >= 0 && high < constant.PrizeCodeMax {
				lotteryPrize := &LotteryPrize{
					Id:            prize.Id,
					Title:         prize.Title,
					PrizeNum:      prize.PrizeNum,
					LeftNum:       prize.LeftNum,
					PrizeCodeLow:  low,
					PrizeCodeHigh: high,
					Img:           prize.Img,
					DisplayOrder:  prize.DisplayOrder,
					PrizeType:     prize.PrizeType,
					PrizeProfile:  prize.PrizeProfile,
				}
				lotteryPrizeList = append(lotteryPrizeList, lotteryPrize)
			}
		}
	}
	return lotteryPrizeList, nil
}

// PrizeCouponDiff 发放不同编码的优惠券
func (l *lotteryService) PrizeCouponDiff(ctx context.Context, prizeID int) (string, error) {
	// 分布式锁保证查询和更新操作的原子性，并且保证每个连续操作串行执行
	// 因为需要更新数据的信息，所以要select，单纯用条件update只会返回受影响的记录数，不会返回具体信息，就拿不到优惠券的编码，所以需要两个操作，先select，再update
	key := fmt.Sprint(0 - prizeID - constant.CouponDiffLockLimit)
	lock1 := lock.NewRedisLock(key, lock.WithExpireSeconds(5), lock.WithWatchDogMode())
	if err := lock1.Lock(ctx); err != nil {
		log.ErrorContextf(ctx, "lotteryService|PrizeCouponDiff:%v", err)
		return "", fmt.Errorf("lotteryService|PrizeCouponDiff:%v", err)
	}
	defer lock1.Unlock(ctx)

	db := gormcli.GetDB()
	// 查询
	couponID := 0
	coupon, err := l.couponReop.GetGetNextUsefulCoupon(db, prizeID, couponID)
	if err != nil {
		log.ErrorContextf(ctx, "lotteryService|PrizeCouponDiff:%v\n", err)
		return "", err
	}
	if coupon == nil {
		log.InfoContextf(ctx, "lotteryService|PrizeCouponDiff: coupon is nil")
		return "", nil
	}
	// 更新
	coupon.SysStatus = 2
	if err := l.couponReop.Update(db, coupon, "sys_status"); err != nil {
		log.ErrorContextf(ctx, "lotteryService|PrizeCouponDiff:%v\n", err)
		return "", err
	}
	return coupon.Code, nil
}

// PrizeCouponDiffWithCache 带缓存的优惠券发奖，从缓存中拿出一个优惠券,要用缓存的话，需要再项目启动的时候将优惠券导入到缓存
func (l *lotteryService) PrizeCouponDiffWithCache(ctx context.Context, prizeID int) (string, error) {
	code, err := l.couponReop.GetNextUsefulCouponFromCache(prizeID)
	if err != nil {
		return "", fmt.Errorf("lotteryService|PrizeCouponDiffByCache:%v", err)
	}
	if code == "" {
		log.InfoContextf(ctx, "lotteryService|PrizeCouponDiffByCache code is nil with prize_id=%d", prizeID)
		return "", nil
	}
	coupon := model.Coupon{
		Code:      code,
		SysStatus: 2,
	}
	if err = l.couponReop.UpdateByCode(gormcli.GetDB(), code, &coupon, "sys_status"); err != nil {
		return "", fmt.Errorf("lotteryService|PrizeCouponDiffByCache:%v", err)
	}
	return code, nil
}

func (l *lotteryService) PrizeLargeBlackLimit(ctx context.Context, blackUser *model.BlackUser,
	blackIp *model.BlackIp, lotteryUserInfo *LotteryUserInfo) error {
	now := time.Now()
	blackTime := constant.DefaultBlackTime
	// 用户黑明单限制
	if blackUser == nil || blackUser.UserId <= 0 {
		blackUserInfo := &model.BlackUser{
			Id:        lotteryUserInfo.UserID,
			UserName:  lotteryUserInfo.UserName,
			BlackTime: now.Add(time.Second * time.Duration(blackTime)),
			//SysCreated: time.Time{},
			//SysUpdated: time.Time{},
			SysIp: lotteryUserInfo.IP,
		}
		if err := l.blackUserRepo.Create(gormcli.GetDB(), blackUserInfo); err != nil {
			log.ErrorContextf(ctx, "lotteryService|PrizeLargeBlackLimit:%v", err)
			return fmt.Errorf("lotteryService|PrizeLargeBlackLimit:%v", err)
		}
	} else {
		blackUserInfo := &model.BlackUser{
			UserId:    lotteryUserInfo.UserID,
			BlackTime: now.Add(time.Second * time.Duration(blackTime)),
		}
		if err := l.blackUserRepo.Update(gormcli.GetDB(), lotteryUserInfo.UserID, blackUserInfo, "black_time"); err != nil {
			log.ErrorContextf(ctx, "lotteryService|PrizeLargeBlackLimit:%v", err)
			return fmt.Errorf("lotteryService|PrizeLargeBlackLimit:%v", err)
		}
	}
	// ip黑明但限制
	if blackIp == nil || blackIp.Ip == "" {
		blackIPInfo := &model.BlackIp{
			Ip:        lotteryUserInfo.IP,
			BlackTime: now.Add(time.Second * time.Duration(blackTime)),
			//SysCreated: time.Time{},
			//SysUpdated: time.Time{},
		}
		if err := l.blackIpRepo.Create(gormcli.GetDB(), blackIPInfo); err != nil {
			log.ErrorContextf(ctx, "lotteryService|PrizeLargeBlackLimit:%v", err)
			return fmt.Errorf("lotteryService|PrizeLargeBlackLimit:%v", err)
		}
	} else {
		blackIPInfo := &model.BlackIp{
			Ip:        lotteryUserInfo.IP,
			BlackTime: now.Add(time.Second * time.Duration(blackTime)),
			//SysUpdated: time.Time{},
		}
		if err := l.blackIpRepo.Update(gormcli.GetDB(), lotteryUserInfo.IP, blackIPInfo, "black_time"); err != nil {
			log.ErrorContextf(ctx, "lotteryService|PrizeLargeBlackLimit:%v", err)
			return fmt.Errorf("lotteryService|PrizeLargeBlackLimit:%v", err)
		}
	}
	return nil
}

func (l *lotteryService) GetPrizeNumWithPool(ctx context.Context, prizeID uint) (int, error) {
	num, err := l.prizeReop.GetPrizePoolNum(prizeID)
	if err != nil {
		log.ErrorContextf(ctx, "lotteryService|GetPrizeNumWithPool err: %v", err)
		return 0, fmt.Errorf("lotteryService|GetPrizeNumWithPool:%v", err)
	}
	return num, nil
}

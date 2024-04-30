package service

import (
	"context"
	"fmt"
	"lottery_single/internal/model"
	"lottery_single/internal/pkg/constant"
	"lottery_single/internal/pkg/middlewares/cache"
	"lottery_single/internal/pkg/middlewares/gormcli"
	"lottery_single/internal/pkg/middlewares/log"
	"lottery_single/internal/pkg/utils"
	"lottery_single/internal/repo"
	"math"
	"strconv"
	"time"
)

// LimitService 用户功能
type LimitService interface {
	GetUserCurrentLotteryTimes(ctx context.Context, uid uint) (*model.LotteryTimes, error)
	CheckUserDayLotteryTimes(ctx context.Context, uid uint) (bool, error)
	CheckUserDayLotteryTimesWithCache(ctx context.Context, uid uint) (bool, error)
	CheckIPLimit(ctx context.Context, ip string) int64
	CheckBlackIP(ctx context.Context, ip string) (bool, *model.BlackIp, error)
	CheckBlackIPWithCache(ctx context.Context, ip string) (bool, *model.BlackIp, error)
	CheckBlackUser(ctx context.Context, uid uint) (bool, *model.BlackUser, error)
	CheckBlackUserWithCache(ctx context.Context, uid uint) (bool, *model.BlackUser, error)
}

type limitService struct {
	lotteryTimesReop *repo.LotteryTimesRepo
	blackIpRepo      *repo.BlackIpRepo
	blackUserRepo    *repo.BlackUserRepo
}

var limitServiceImpl *limitService

func InitLimitService() {
	limitServiceImpl = &limitService{
		lotteryTimesReop: repo.NewLotteryTimesRepo(gormcli.GetDB(), cache.GetRedisCli()),
		blackIpRepo:      repo.NewBlackIpRepo(),
		blackUserRepo:    repo.NewBlackUserRepo(),
	}
}

func GetLimitService() LimitService {
	return limitServiceImpl
}

// GetUserCurrentLotteryTimes 获取当天该用户的抽奖次数
func (l *limitService) GetUserCurrentLotteryTimes(ctx context.Context, uid uint) (*model.LotteryTimes, error) {
	y, m, d := time.Now().Date()
	strDay := fmt.Sprintf("%d%02d%02d", y, m, d)
	day, _ := strconv.Atoi(strDay)
	lotteryTimes, err := l.lotteryTimesReop.GetByUserIDAndDay(uid, uint(day))
	if err != nil {
		log.ErrorContextf(ctx, "lotteryTimesService|GetUserCurrentLotteryTimes:%v", err)
		return nil, err
	}
	return lotteryTimes, nil
}

// CheckUserDayLotteryTimes 判断当天是否还可以进行抽奖
func (l *limitService) CheckUserDayLotteryTimes(ctx context.Context, uid uint) (bool, error) {
	userLotteryTimes, err := l.GetUserCurrentLotteryTimes(ctx, uid)
	if err != nil {
		return false, fmt.Errorf("checkUserDayLotteryTimes|err:%v", err)
	}
	if userLotteryTimes != nil {
		// 今天的抽奖记录已经达到了抽奖次数限制
		if userLotteryTimes.Num >= constant.UserPrizeMax {
			return false, nil
		} else {
			userLotteryTimes.Num++
			if err := l.lotteryTimesReop.Update(userLotteryTimes, "num"); err != nil {
				return false, fmt.Errorf("updateLotteryTimes｜update:%v", err)
			}
		}
		return true, nil
	}
	y, m, d := time.Now().Date()
	strDay := fmt.Sprintf("%d%02d%02d", y, m, d)
	day, _ := strconv.Atoi(strDay)
	lotteryTimeesInfo := &model.LotteryTimes{
		UserId: uid,
		Day:    uint(day),
		Num:    1,
	}
	if err := l.lotteryTimesReop.Create(lotteryTimeesInfo); err != nil {
		return false, fmt.Errorf("updateLotteryTimes｜create:%v", err)
	}
	return true, nil
}

func (l *limitService) CheckUserDayLotteryTimesWithCache(ctx context.Context, uid uint) (bool, error) {
	// 通过缓存验证
	userLotteryNum := l.lotteryTimesReop.IncrUserDayLotteryNum(uid)
	log.InfoContextf(ctx, "CheckUserDayLotteryTimesWithCache|userLotteryNum = %d", userLotteryNum)
	// 缓存验证没通过，直接返回
	if userLotteryNum > constant.UserPrizeMax {
		return false, nil
	}
	// 通过数据库验证，还要在数据库中做一次验证
	userLotteryTimes, err := l.GetUserCurrentLotteryTimes(ctx, uid)
	if err != nil {
		return false, fmt.Errorf("checkUserDayLotteryTimes|err:%v", err)
	}
	if userLotteryTimes != nil {
		// 数据库验证今天的抽奖记录已经达到了抽奖次数限制，不能在抽奖
		if userLotteryTimes.Num >= constant.UserPrizeMax {
			// 缓存数据不可靠，不对，需要更新
			if int64(userLotteryTimes.Num) > userLotteryNum {
				if err = l.lotteryTimesReop.InitUserLuckyNum(uid, int64(userLotteryTimes.Num)); err != nil {
					return false, fmt.Errorf("limitService|CheckUserDayLotteryTimesWithCache:%v", err)
				}
			}
			return false, nil
		} else { // 数据库验证通过，今天还可以抽奖
			userLotteryTimes.Num++
			// 此时次数抽奖次数增加了，需要更新缓存
			if int64(userLotteryTimes.Num) > userLotteryNum {
				if err = l.lotteryTimesReop.InitUserLuckyNum(uid, int64(userLotteryTimes.Num)); err != nil {
					return false, fmt.Errorf("limitService|CheckUserDayLotteryTimesWithCache:%v", err)
				}
			}
			// 更新数据库
			if err = l.lotteryTimesReop.Update(userLotteryTimes); err != nil {
				return false, fmt.Errorf("updateLotteryTimes｜update:%v", err)
			}
		}
		return true, nil
	}
	y, m, d := time.Now().Date()
	strDay := fmt.Sprintf("%d%02d%02d", y, m, d)
	day, _ := strconv.Atoi(strDay)
	lotteryTimesInfo := &model.LotteryTimes{
		UserId: uid,
		Day:    uint(day),
		Num:    1,
	}
	if err = l.lotteryTimesReop.Create(lotteryTimesInfo); err != nil {
		return false, fmt.Errorf("updateLotteryTimes｜create:%v", err)
	}
	if err = l.lotteryTimesReop.InitUserLuckyNum(uid, 1); err != nil {
		return false, fmt.Errorf("limitService|CheckUserDayLotteryTimesWithCache:%v", err)
	}
	return true, nil
}

// CheckIPLimit 验证ip抽奖是否受限制
func (l *limitService) CheckIPLimit(ctx context.Context, strIp string) int64 {
	ip := utils.Ip4toInt(strIp)
	i := ip % constant.IpFrameSize
	key := fmt.Sprintf("day_ip_num_%d", i)
	ret, err := cache.GetRedisCli().HIncrBy(ctx, key, strIp, 1)
	if err != nil {
		log.ErrorContextf(ctx, "CheckIPLimit|Incr:%v", err)
		return math.MaxInt32
	}
	return ret
}

func (l *limitService) CheckBlackIP(ctx context.Context, ip string) (bool, *model.BlackIp, error) {
	info, err := l.blackIpRepo.GetByIP(gormcli.GetDB(), ip)
	if err != nil {
		log.ErrorContextf(ctx, "CheckBlackIP|GetByIP:%v", err)
		return false, nil, fmt.Errorf("CheckBlackIP|GetByIP:%v", err)
	}
	if info == nil || info.Ip == "" {
		return true, nil, nil
	}
	if time.Now().Before(info.BlackTime) {
		// IP黑名单存在，而且还在黑名单有效期内
		return false, info, nil
	}
	return true, info, nil
}

func (l *limitService) CheckBlackIPWithCache(ctx context.Context, ip string) (bool, *model.BlackIp, error) {
	info, err := l.blackIpRepo.GetByIPWithCache(gormcli.GetDB(), ip)
	if err != nil {
		log.ErrorContextf(ctx, "CheckBlackIP|GetByIP:%v", err)
		return false, nil, fmt.Errorf("CheckBlackIP|GetByIP:%v", err)
	}
	if info == nil || info.Ip == "" {
		return true, nil, nil
	}
	if time.Now().Before(info.BlackTime) {
		// IP黑名单存在，而且还在黑名单有效期内
		return false, info, nil
	}
	return true, info, nil
}

func (l *limitService) CheckBlackUser(ctx context.Context, uid uint) (bool, *model.BlackUser, error) {
	info, err := l.blackUserRepo.GetByUserID(gormcli.GetDB(), uid)
	if err != nil {
		log.ErrorContextf(ctx, "CheckBlackUser|Get:%v", err)
		return false, nil, fmt.Errorf("CheckBlackUser|Get:%v", err)
	}
	// 黑名单存在并且有效，不能通过
	if info != nil && time.Now().Before(info.BlackTime) {
		return false, info, nil
	}
	return true, info, nil
}

func (l *limitService) CheckBlackUserWithCache(ctx context.Context, uid uint) (bool, *model.BlackUser, error) {
	info, err := l.blackUserRepo.GetByUserIDWithCache(gormcli.GetDB(), uid)
	if err != nil {
		log.ErrorContextf(ctx, "CheckBlackUser|Get:%v", err)
		return false, nil, fmt.Errorf("CheckBlackUser|Get:%v", err)
	}
	if info != nil && info.BlackTime.Unix() > time.Now().Unix() {
		// 黑名单存在并且有效
		return false, info, nil
	}
	return true, info, nil
}

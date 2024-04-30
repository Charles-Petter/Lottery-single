package handlers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"lottery_single/internal/handlers/params"
	"lottery_single/internal/pkg/constant"
	"lottery_single/internal/pkg/middlewares/lock"
	"lottery_single/internal/pkg/middlewares/log"
	"lottery_single/internal/pkg/utils"
	"lottery_single/internal/service"
	"net/http"
)

type LotteryHandlerV3 struct {
	req  *params.LotteryReq
	resp HttpResponse

	// 需要什么Service，就在这里声明
	limitService   service.LimitService
	lotteryService service.LotteryService
	resultService  service.ResultService
}

func LotteryV3(c *gin.Context) {
	h := LotteryHandlerV3{
		limitService:   service.GetLimitService(),
		lotteryService: service.GetLotteryService(),
		resultService:  service.GetResultService(),
	}
	// HTTP响应
	defer func() {
		// 通过对应的Code，获取Msg
		h.resp.Msg = constant.GetErrMsg(h.resp.Code)
		c.JSON(http.StatusOK, h.resp)
	}()
	if err := c.ShouldBind(&h.req); err != nil {
		log.Errorf("LotteryV3|Error binding:%v", err)
		h.resp.Code = constant.ErrShouldBind
		return
	}
	Run(&h)
}

func (l *LotteryHandlerV3) CheckInput(ctx context.Context) error {
	if l.req == nil {
		l.resp.Code = constant.ErrInputInvalid
		log.Errorf("lottery params is nil")
		return fmt.Errorf(constant.GetErrMsg(constant.ErrInputInvalid))
	}
	if l.req.Token == "" || l.req.UserID <= 0 {
		l.resp.Code = constant.ErrInputInvalid
		log.Errorf("lottery params invalid, token=%s,user_id=%s\n", l.req.Token, l.req.UserID)
		return fmt.Errorf(constant.GetErrMsg(constant.ErrInputInvalid))
	}
	return nil
}

// Process 抽奖逻辑实现
func (l *LotteryHandlerV3) Process(ctx context.Context) {
	var (
		ok  bool
		err error
	)
	// 1. 根据token解析出用户信息
	jwtClaims, err := utils.ParseJwtToken(l.req.Token, constant.SecretKey)
	if err != nil || jwtClaims == nil {
		l.resp.Code = constant.ErrJwtParse
		log.Errorf("jwt parse err, token=%s,user_id=%s\n", l.req.Token, l.req.UserID)
		return
	}
	userID := jwtClaims.UserID

	lockKey := getLotteryLockKey(userID)
	lock1 := lock.NewRedisLock(lockKey, lock.WithExpireSeconds(5), lock.WithWatchDogMode())

	// 1. 用户抽奖分布式锁定,防重入
	if err := lock1.Lock(ctx); err != nil {
		l.resp.Code = constant.ErrInternalServer
		log.ErrorContextf(ctx, "LotteryHandler|Process:%v", err)
		return
	}
	defer lock1.Unlock(ctx)

	// 2. 验证用户今日抽奖次数
	ok, err = l.limitService.CheckUserDayLotteryTimesWithCache(ctx, userID)
	if err != nil {
		l.resp.Code = constant.ErrInternalServer
		log.ErrorContextf(ctx, "LotteryHandler|CheckUserDayLotteryTimes:%v", err)
		return
	}
	if !ok {
		l.resp.Code = constant.ErrUserLimitInvalid
		log.InfoContextf(ctx, "LotteryHandler|CheckUserDayLotteryTimes:%v", err)
		return
	}

	// 3. 验证当天IP参与的抽奖次数
	ipDayLotteryTimes := l.limitService.CheckIPLimit(ctx, l.req.IP)
	if ipDayLotteryTimes > constant.IpLimitMax {
		l.resp.Code = constant.ErrIPLimitInvalid
		log.InfoContextf(ctx, "LotteryHandler|CheckUserDayLotteryTimes:%v", err)
		return
	}

	// 4. 验证IP是否在ip黑名单
	ok, blackIpInfo, err := l.limitService.CheckBlackIPWithCache(ctx, l.req.IP)
	if err != nil {
		l.resp.Code = constant.ErrInternalServer
		log.ErrorContextf(ctx, "LotteryHandler|CheckBlackIP:%v", err)
		return
	}
	// ip黑明单生效
	if !ok {
		l.resp.Code = constant.ErrBlackedIP
		log.InfoContextf(ctx, "LotteryHandler|CheckBlackIP blackIpInfo is %v\n", blackIpInfo)
		return
	}

	// 5. 验证用户是否在黑明单中
	ok, blackUserInfo, err := l.limitService.CheckBlackUserWithCache(ctx, userID)
	if err != nil {
		l.resp.Code = constant.ErrInternalServer
		log.ErrorContextf(ctx, "LotteryHandler|CheckBlackUser:%v", err)
		return
	}
	// 用户黑明单生效
	if !ok {
		l.resp.Code = constant.ErrBlackedUser
		log.ErrorContextf(ctx, "LotteryHandler|CheckBlackUser blackUserInfo is %v\n", blackUserInfo)
		return
	}

	// 6. 抽奖逻辑实现
	prizeCode := utils.Random(constant.PrizeCodeMax)
	prize, err := l.lotteryService.GetPrizeWithCache(ctx, prizeCode)
	if err != nil {
		l.resp.Code = constant.ErrInternalServer
		log.ErrorContextf(ctx, "LotteryHandler|CheckBlackUser:%v", err)
		return
	}
	if prize == nil || prize.PrizeNum < 0 || (prize.PrizeNum > 0 && prize.LeftNum <= 0) {
		l.resp.Code = constant.ErrNotWon
		return
	}

	// 7. 有剩余奖品发放
	if prize.PrizeNum > 0 {
		num, err := l.lotteryService.GetPrizeNumWithPool(ctx, prize.Id)
		if err != nil {
			l.resp.Code = constant.ErrInternalServer
			log.ErrorContextf(ctx, "LotteryHandler|GiveOutPrize:%v", err)
			return
		}
		// 奖品池奖品不够，不能发奖
		if num <= 0 {
			l.resp.Code = constant.ErrNotWon
			log.InfoContextf(ctx, "LotteryHandler|GiveOutPrize|prize num not enough")
			return
		}
		// 奖品池奖品足够，可以发奖
		ok, err = l.lotteryService.GiveOutPrizeWithPool(ctx, int(prize.Id))
		if err != nil {
			l.resp.Code = constant.ErrInternalServer
			log.ErrorContextf(ctx, "LotteryHandler|GiveOutPrize:%v", err)
			return
		}
		// 奖品不足，发放失败
		if !ok {
			l.resp.Code = constant.ErrPrizeNotEnough
			log.InfoContextf(ctx, "LotteryHandler|GiveOutPrize:%v", err)
			return
		}
	}

	/***如果中奖记录重要的的话，可以考虑用事务将下面逻辑包裹*****/
	// 8. 发优惠券
	if prize.PrizeType == constant.PrizeTypeCouponDiff {
		code, err := l.lotteryService.PrizeCouponDiffWithCache(ctx, int(prize.Id))
		if err != nil {
			l.resp.Code = constant.ErrInternalServer
			log.InfoContextf(ctx, "LotteryHandler|PrizeCouponDiff:%v", err)
			return
		}
		if code == "" {
			l.resp.Code = constant.ErrNotWon
			log.InfoContext(ctx, "LotteryHandler|PrizeCouponDiff coupon left is nil")
			return
		}
	}
	l.resp.Data = prize

	// 9 记录中奖纪录
	if err := l.resultService.LotteryResult(ctx, prize, userID, jwtClaims.UserName, l.req.IP, prizeCode); err != nil {
		l.resp.Code = constant.ErrInternalServer
		log.InfoContextf(ctx, "LotteryHandler|PrizeCouponDiff:%v", err)
		return
	}

	// 10. 如果中了实物大奖，需要把ip和用户置于黑明单中一段时间，防止同一个用户频繁中大奖
	if prize.PrizeType == constant.PrizeTypeEntityLarge {
		lotteryUserInfo := service.LotteryUserInfo{
			UserID:   userID,
			UserName: jwtClaims.UserName,
			IP:       l.req.IP,
		}
		if err := l.lotteryService.PrizeLargeBlackLimit(ctx, blackUserInfo, blackIpInfo, &lotteryUserInfo); err != nil {
			l.resp.Code = constant.ErrInternalServer
			log.InfoContextf(ctx, "LotteryHandler|PrizeCouponDiff:%v", err)
			return
		}
	}
}

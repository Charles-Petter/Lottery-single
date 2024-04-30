package service

import (
	"context"
	"fmt"
	"lottery_single/internal/model"
	"lottery_single/internal/pkg/middlewares/gormcli"
	"lottery_single/internal/pkg/middlewares/log"
	"lottery_single/internal/repo"
)

type ResultService interface {
	LotteryResult(ctx context.Context, prize *LotteryPrize, uid uint, userName, ip string, prizeCode int) error
}

type resultService struct {
	resultReop *repo.ResultRepo
}

var resultServiceImpl *resultService

func GetResultService() ResultService {
	resultServiceImpl = &resultService{
		resultReop: repo.NewResultRepo(),
	}
	return resultServiceImpl
}

func (r *resultService) LotteryResult(ctx context.Context, prize *LotteryPrize, uid uint, userName, ip string, prizeCode int) error {
	result := model.Result{
		PrizeId:   prize.Id,
		PrizeName: prize.Title,
		PrizeType: prize.PrizeType,
		UserId:    uid,
		UserName:  userName,
		PrizeCode: uint(prizeCode),
		PrizeData: prize.PrizeProfile,
		//SysCreated: time.Now(),
		SysIp:     ip,
		SysStatus: 0,
	}
	if err := r.resultReop.Create(gormcli.GetDB(), &result); err != nil {
		log.ErrorContextf(ctx, "resultService|LotteryResult:%v", err)
		return fmt.Errorf("resultService|LotteryResult:%v", err)
	}
	return nil
}

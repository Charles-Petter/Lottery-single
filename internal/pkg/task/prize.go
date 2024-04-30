package task

import (
	"context"
	"encoding/json"
	"fmt"
	"lottery_single/internal/model"
	"lottery_single/internal/pkg/constant"
	"lottery_single/internal/pkg/middlewares/cache"
	"lottery_single/internal/pkg/middlewares/gormcli"
	"lottery_single/internal/pkg/middlewares/log"
	"lottery_single/internal/pkg/utils"
	"lottery_single/internal/service"
	"strconv"
	"time"
)

/**
 * 根据奖品的发奖计划，把设定的奖品数量放入奖品池
 * 需要每分钟执行一次
 * 【难点】定时程序，根据奖品设置的数据，更新奖品池的数据
 */

func DoPrizePlanTask() {
	go ResetAllPrizePlan()
	go FillAllPrizePool()
}

// ResetAllPrizePlan 重置所有奖品的发奖计划
func ResetAllPrizePlan() {
	log.Infof("Resetting all prizes!!!!!")
	adminService := service.GetAdminService()
	prizeList, err := adminService.GetPrizeList(context.Background())
	if err != nil {
		log.Errorf("ResetAllPrizePlan err:%v", err)
	}
	now := time.Now()
	for _, prize := range prizeList {
		if prize.PrizeTime > 0 && (prize.PrizePlan == "" || prize.PrizeEnd.Before(now)) {
			// ResetPrizePlan只会更新db的数据
			if err = adminService.ResetPrizePlan(context.Background(), prize); err != nil {
				log.Errorf("ResetAllPrizePlan err:%v", err)
			}
			// 通过读取缓存将db的数据同步到缓存中
			_, err = adminService.GetPrizeListWithCache(context.Background())
			if err != nil {
				log.Errorf("ResetAllPrizePlan err:%v", err)
			}
		}
	}
	// 每5分钟执行一次
	time.AfterFunc(5*time.Minute, ResetAllPrizePlan)
}

func FillAllPrizePool() {
	log.Infof("FillAllPrizePool!!!!")
	totalNum, err := fillPrizePool()
	if err != nil {
		log.Errorf("FillAllPrizePool err:%v", err)
	}
	log.Infof("FillAllPrizePool with num:%d", totalNum)
	time.AfterFunc(time.Minute, FillAllPrizePool)
}

func fillPrizePool() (int, error) {
	totalNum := 0
	adminService := service.GetAdminService()
	prizeList, err := adminService.GetPrizeList(context.Background())
	now := time.Now()
	if err != nil {
		log.Errorf("FillPrizePool err:%v", err)
		return 0, fmt.Errorf("FillPrizePool|GetPrizeList:%v", err)
	}
	if prizeList == nil || len(prizeList) == 0 {
		return 0, nil
	}
	for _, prize := range prizeList {
		if prize.SysStatus != 1 {
			continue
		}
		if prize.PrizeNum <= 0 {
			continue
		}
		if prize.BeginTime.After(now) || prize.EndTime.Before(now) {
			continue
		}
		// 发奖计划数据不正确
		if len(prize.PrizePlan) <= 7 {
			continue
		}
		prizePlanList := []*service.TimePrizeInfo{}
		if err = json.Unmarshal([]byte(prize.PrizePlan), &prizePlanList); err != nil {
			log.Errorf("FillPrizePool|Unmarshal TimePrizeInfo err:%v", err)
			return 0, fmt.Errorf("FillPrizePool|Unmarshal TimePrizeInfo:%v", err)
		}
		index := 0
		prizeNum := 0
		for i, prizePlanInfo := range prizePlanList {
			log.Infof("fillPrizePool|prize_id=%d\n, prizePlanInfo=%+v", prize.Id, prizePlanInfo)
			t, err := utils.ParseTime(prizePlanInfo.Time)
			if err != nil {
				log.Errorf("FillPrizePool|ParseTime err:%v", err)
				return 0, fmt.Errorf("FillPrizePool|ParseTime:%v", err)
			}
			if t.After(now) {
				break
			}
			// 该类奖品中，之前没有发放的奖品数量都要放入奖品池
			prizeNum += prizePlanInfo.Num
			index = i + 1
		}
		if prizeNum > 0 {
			incrPrizePool(prize.Id, prizeNum)
			totalNum += prizeNum
		}
		// 更新发奖计划
		if index > 0 {
			if index < len(prizePlanList) {
				prizePlanList = prizePlanList[index:]
			} else {
				prizePlanList = make([]*service.TimePrizeInfo, 0)
			}
			// 将新的发奖计划更新到数据库
			bytes, err := json.Marshal(&prizePlanList)
			if err != nil {
				log.Errorf("FillPrizePool|Marshal err:%v", err)
				return 0, fmt.Errorf("FillPrizePool|Marshal:%v", err)
			}
			updatePrize := &model.Prize{
				Id:        prize.Id,
				PrizePlan: string(bytes),
			}
			if err = adminService.UpdateDbPrizeWithCache(context.Background(), gormcli.GetDB(), updatePrize, "prize_plan"); err != nil {
				log.Errorf("FillPrizePool|UpdateDbPrizeWithCache err:%v", err)
				return 0, fmt.Errorf("FillPrizePool|UpdateDbPrizeWithCache:%v", err)
			}
		}
		if totalNum > 0 {
			// 将更新后的数据加载到缓存中
			_, err = adminService.GetPrizeListWithCache(context.Background())
			if err != nil {
				log.Errorf("FillPrizePool|GetPrizeListWithCache err:%v", err)
				return 0, fmt.Errorf("FillPrizePool|GetPrizeListWithCache:%v", err)
			}
		}
	}
	return totalNum, nil
}

// incrPrizePool 根据计划数据，往奖品池增加奖品数量
func incrPrizePool(prizeID uint, num int) (int, error) {
	key := constant.PrizePoolCacheKey
	idStr := strconv.Itoa(int(prizeID))
	cnt, err := cache.GetRedisCli().HIncrBy(context.Background(), key, idStr, int64(num))
	if err != nil {
		log.Errorf("incrPrizePool err:%v", err)
		return 0, fmt.Errorf("incrPrizePool err:%v", err)
	}
	if int(cnt) < num {
		log.Infof("incrPrizePool twice,num=%d,cnt=%d", num, int(cnt))
		left := num - int(cnt)
		// 数量不等，存在没有成功的情况，补偿一次
		cnt, err = cache.GetRedisCli().HIncrBy(context.Background(), key, idStr, int64(left))
		if err != nil {
			log.Errorf("incrPrizePool twice err:%v", err)
			return 0, fmt.Errorf("incrPrizePool err:%v", err)
		}
	}
	return int(cnt), nil
}

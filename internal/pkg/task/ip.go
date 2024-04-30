package task

import (
	"context"
	"fmt"
	"lottery_single/internal/pkg/constant"
	"lottery_single/internal/pkg/middlewares/cache"
	"lottery_single/internal/pkg/middlewares/log"
	"lottery_single/internal/pkg/utils"
	"time"
)

//func init() {
//	// 本地开发测试的时候，每次启动归零
//	ResetIPLotteryNums()
//

func DoResetIPLotteryNumsTask() {
	go ResetIPLotteryNums()
}

func ResetIPLotteryNums() {
	log.Infof("重置所有的IP抽奖次数")
	for i := 0; i < constant.IpFrameSize; i++ {
		key := fmt.Sprintf("day_ip_num_%d", i)
		if err := cache.GetRedisCli().Delete(context.Background(), key); err != nil {
			log.Errorf("ResetIPLotteryNums err:%v", err)
		}
	}

	// IP当天的统计数，整点归零，设置定时器
	duration := utils.NextDayDuration()
	time.AfterFunc(duration, ResetIPLotteryNums)
}

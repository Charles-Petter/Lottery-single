package utils

import (
	"fmt"
	"lottery_single/internal/pkg/constant"
)

func GetLotteryLockKey(uid uint) string {
	return fmt.Sprintf(constant.LotteryLockKeyPrefix+"%d", uid)
}

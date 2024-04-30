package utils

import (
	_ "fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"lottery_single/internal/model"
	"lottery_single/internal/repo"
)

func TestPrizeRepo(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	prizeRepo := repo.NewPrizeRepo()

	// 在这里使用奖品仓库的方法进行测试
	t.Run("TestGetAllWithCache", func(t *testing.T) {
		prizes, err := prizeRepo.GetAllWithCache(db)
		assert.NoError(t, err)
		assert.NotNil(t, prizes)
	})

	t.Run("TestDecrLeftNumByPool", func(t *testing.T) {
		prizeID := 1 // 奖品ID，请确保该ID在测试数据库中存在
		// 尝试递减奖品池中奖品的数量
		cnt, err := prizeRepo.DecrLeftNumByPool(prizeID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, cnt, int64(0))
	})

	t.Run("TestSetAllByCache", func(t *testing.T) {
		// 创建奖品列表
		now := time.Date(2024, 4, 24, 0, 0, 0, 0, time.UTC)
		prizes := []*model.Prize{
			{Id: 12, Title: "奖品1", PrizeNum: 10, LeftNum: 10, SysCreated: &now, SysUpdated: &now},
			{Id: 21, Title: "奖品2", PrizeNum: 20, LeftNum: 20, SysCreated: &now, SysUpdated: &now},
			// 在这里添加更多奖品
		}
		// 保存奖品列表到缓存中
		err := prizeRepo.SetAllByCache(prizes)
		assert.NoError(t, err)
	})

	t.Run("TestUpdateByCache", func(t *testing.T) {
		// 创建一个奖品用于更新缓存
		now := time.Date(2024, 4, 24, 0, 0, 0, 0, time.UTC)
		prize := &model.Prize{Id: 1, Title: "更新后的奖品1", PrizeNum: 5, LeftNum: 5, SysCreated: &now, SysUpdated: &now}
		// 更新缓存
		err := prizeRepo.UpdateByCache(prize)
		assert.NoError(t, err)
	})

	t.Run("TestGetAllByCache", func(t *testing.T) {
		// 从缓存中获取所有奖品
		prizes, err := prizeRepo.GetAllByCache()
		assert.NoError(t, err)
		assert.NotNil(t, prizes)
	})

	// 添加更多测试用例以测试其他方法

	// 关闭数据库连接
	//err = db.Close()
	if err != nil {
		log.Fatalf("failed to close database connection: %v", err)
	}
}

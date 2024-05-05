package router

import (
	"github.com/gin-gonic/gin"
	"lottery_single/internal/handlers"
	"net/http"
)

const SessionKey = "lottery_session" // 鉴权session

// AuthMiddleWare 鉴权中间件
func AuthMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		if session, err := c.Cookie(SessionKey); err == nil {
			if session != "" {
				c.Next()
				return
			}
		}
		// 返回错误
		c.JSON(http.StatusUnauthorized, gin.H{"error": "err"})
		c.Abort()
		return
	}
}

func setRoutes(r *gin.Engine) {
	setAdminRoutes(r)
	setLotteryRoutes(r)
}

func setAdminRoutes(r *gin.Engine) {
	adminGroup := r.Group("admin")
	// 获取奖品列表
	adminGroup.GET("/get_prize_list", handlers.GetPrizeList)
	// 添加奖品
	adminGroup.POST("/add_prize", handlers.PrizeAdd)
	// 导入优惠券
	adminGroup.POST("/import_coupon", handlers.CouponImport)
	// 用户登录
	adminGroup.POST("/login", handlers.Login)
	//注册
	adminGroup.POST("/register", handlers.Register)

	adminGroup.POST("/get_lucky", handlers.LotteryV1)
	//用户管理
	adminGroup.POST("/add_user", handlers.AddUser)
	adminGroup.PUT("/update_user/:userID", handlers.UpdateUser)
	adminGroup.DELETE("/delete_user/:userID", handlers.DeleteUser)
	adminGroup.GET("/get_all_users", handlers.GetAllUsers)
}

func setLotteryRoutes(r *gin.Engine) {
	lotteryGroup := r.Group("lottery")
	// 基础版获取中奖
	lotteryGroup.POST("/v1/get_lucky", handlers.LotteryV1)
	// 优化V1版中奖逻辑
	lotteryGroup.POST("/v2/get_lucky", handlers.LotteryV2)

	//lotteryGroup.Use(AuthMiddleWare()) // Apply authentication middleware to this group
	// 新增抽奖结果展示路由
	lotteryGroup.GET("/show_results", handlers.ShowLotteryResult)
}

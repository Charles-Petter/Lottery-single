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
	setBlackIpRoutes(r)
}

func setAdminRoutes(r *gin.Engine) {
	adminGroup := r.Group("admin")

	// 获取奖品列表
	adminGroup.GET("/get_prize_list", handlers.GetPrizeList)
	// 设置添加抽奖奖品
	adminGroup.POST("/add_prize", handlers.PrizeAdd)

	// 上传图片
	adminGroup.POST("/upload", handlers.UploadImage)
	// 删除奖品
	adminGroup.DELETE("/delete_prize/:id", handlers.DeletePrize)

	// 导入优惠券
	adminGroup.POST("/import_coupon", handlers.CouponImport)
	// 获取优惠券列表
	adminGroup.GET("/get_coupon_list", handlers.GetCouponList)

	// 用户登录
	adminGroup.POST("/login", handlers.Login)
	//注册
	adminGroup.POST("/register", handlers.Register)

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

	//lotteryGroup.Use(AuthMiddleWare())
	// 抽奖结果展示
	lotteryGroup.GET("/show_results", handlers.ShowLotteryResult)
}

func setBlackIpRoutes(r *gin.Engine) {
	blackIpGroup := r.Group("/admin/blackip")
	// 添加IP到黑名单
	blackIpGroup.POST("/add", handlers.AddBlackIP)
	// 删除黑名单中的IP
	blackIpGroup.DELETE("/delete/:id", handlers.DeleteBlackIP)
	// 查看所有黑名单IP
	blackIpGroup.GET("/list", handlers.ListBlackIP)
}

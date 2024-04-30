package router

import (
	"github.com/gin-gonic/gin"

	"io"
	"lottery_single/configs"
	"lottery_single/internal/pkg/constant"
	"lottery_single/internal/pkg/middlewares/log"
	"lottery_single/internal/pkg/utils"
	"net/http"
	"strconv"
	"time"
)

// InitRouterAndServe 路由配置、启动服务
func InitRouterAndServe() {
	setAppRunMode()
	r := gin.Default()

	setMiddleWare(r) // 修改此处

	// 设置路由
	setRoutes(r)

	// 启动server
	port := configs.GetGlobalConfig().AppConfig.Port
	if err := r.Run(":" + strconv.Itoa(port)); err != nil {
		log.Errorf("start server err:" + err.Error())
	}
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"status": -1, "msg": "请求未携带token，无权限访问"})
			c.Abort()
			return

		}
		claims, err := utils.ParseJwtToken(token, constant.SecretKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"status": -1, "msg": err.Error()})
			c.Abort()
			return
		}
		//token超时
		if time.Now().Unix() > claims.StandardClaims.ExpiresAt {
			c.JSON(http.StatusUnauthorized, gin.H{"status": -1, "msg": "token过期"})
			c.Abort() //阻止执行
			return
		}
		//可以在这里做权限校验
		c.Set("jwtUser", claims)
		c.Next()
	}
}

// setAppRunMode 设置运行模式
func setAppRunMode() {
	if configs.GetGlobalConfig().AppConfig.RunMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
}

func setGinLog(out io.Writer) {
	gin.DefaultWriter = out
	gin.DefaultErrorWriter = out
}

func setMiddleWare(r *gin.Engine) {
	r.Use(func(c *gin.Context) {
		method := c.Request.Method
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
		} else {
			c.Next()
		}
	})
}

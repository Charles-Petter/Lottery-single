package handlers

import (
	"github.com/gin-gonic/gin"
	"lottery_single/internal/handlers/params"
	"lottery_single/internal/service"
	"net/http"
)

// GetCouponList 处理获取优惠券列表的请求
func GetCouponList(c *gin.Context) {
	var req params.CouponListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		// 参数绑定失败，返回错误响应
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用服务层方法获取优惠券列表数据
	coupons, dbCouponNum, cacheCouponNum, err := service.GetAdminService().GetCouponList(c.Request.Context(), req.PrizeID)
	if err != nil {
		// 获取优惠券列表失败，返回错误响应
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 构建响应
	resp := params.CouponListResponse{
		Coupons:        coupons,
		DBCouponNum:    dbCouponNum,
		CacheCouponNum: cacheCouponNum,
	}

	// 返回成功响应
	c.JSON(http.StatusOK, resp)
}
